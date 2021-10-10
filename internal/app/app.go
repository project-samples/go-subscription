package app

import (
	"context"
	"net/http"
	"reflect"

	"github.com/core-go/health"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/pubsub"
	"github.com/core-go/mq/validator"
	v "github.com/go-playground/validator/v10"
)

type ApplicationContext struct {
	Check   func(w http.ResponseWriter, r *http.Request)
	Receive func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handle  func(ctx context.Context, data []byte, header map[string]string, err error) error
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er1 := mongo.Setup(ctx, root.Mongo)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to MongoDB: Error: "+er1.Error())
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver, er2 := pubsub.NewSimpleSubscriberByConfig(ctx, root.Sub.Subscriber, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := mongo.NewMongoWriter(db, "user", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	mongoChecker := mongo.NewHealthChecker(db)
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
		handler = mq.NewHandlerByConfig(root.Sub.Config, writer.Write, &userType, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := pubsub.NewPubHealthChecker("pubsub_publisher", sender.Client, root.Pub.TopicId)
		healthHandler = health.NewHandler(mongoChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(mongoChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(writer.Write, &userType, val.Validate, root.Retry, true, logError, logInfo)
	}

	return &ApplicationContext{
		Check:   healthHandler.Check,
		Receive: receiver.Subscribe,
		Handle:  handler.Handle,
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
