package app

import (
	"context"
	"reflect"

	"github.com/core-go/firestore"
	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/pubsub"
	"github.com/core-go/mq/validator"
	v "github.com/go-playground/validator/v10"
)

type ApplicationContext struct {
	HealthHandler *health.Handler
	Receive       func(ctx context.Context, handle func(context.Context, *mq.Message, error) error)
	Handler       *mq.Handler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	client, er1 := firestore.Connect(ctx, []byte(root.Firestore.Credentials))
	if er1 != nil {
		return nil, er1
	}
	if er1 != nil {
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver, er2 := pubsub.NewSubscriberByConfig(ctx, root.Sub.Subscriber, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := firestore.NewFirestoreWriter(client, "user", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	firestoreChecker := firestore.NewHealthChecker(ctx, []byte(root.Firestore.Credentials))
	receiverChecker := pubsub.NewSubHealthChecker("pubsub_subscriber", receiver.Client, root.Sub.Subscriber.SubscriptionId)
	var healthHandler *health.Handler
	var handler *mq.Handler
	if root.Pub != nil {
		sender, er3 := pubsub.NewPublisherByConfig(ctx, *root.Pub)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Publish, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Sub.Config, userType, writer.Write, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := pubsub.NewPubHealthChecker("pubsub_publisher", sender.Client, root.Pub.TopicId)
		healthHandler = health.NewHandler(firestoreChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(firestoreChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(userType, writer.Write, val.Validate, root.Retry, true, logError, logInfo)
	}

	return &ApplicationContext{
		HealthHandler: healthHandler,
		Receive:       receiver.Subscribe,
		Handler:       handler,
	}, nil
}

func NewUserValidator() validator.Validator {
	val := validator.NewDefaultValidator()
	val.CustomValidateList = append(val.CustomValidateList, validator.CustomValidate{Fn: CheckActive, Tag: "active"})
	return val
}
func CheckActive(fl v.FieldLevel) bool {
	return fl.Field().Bool()
}
