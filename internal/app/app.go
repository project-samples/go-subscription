package app

import (
	"context"
	"reflect"

	"github.com/core-go/health"
	mgo "github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/rabbitmq"
	v "github.com/core-go/mq/validator"
	"github.com/go-playground/validator/v10"
)

type ApplicationContext struct {
	HealthHandler *health.HealthHandler
	Receive       func(ctx context.Context, handle func(context.Context, *mq.Message, error) error)
	Handler       *mq.Handler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er1 := mgo.SetupMongo(ctx, root.Mongo)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to MongoDB: Error: "+er1.Error())
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver, er2 := rabbitmq.NewConsumerByConfig(root.Consumer.RabbitMQ, true, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := mgo.NewInserter(db, "users")
	checker := v.NewErrorChecker(NewUserValidator().Validate)
	validator := mq.NewValidator(userType, checker.Check)

	mongoChecker := mgo.NewHealthChecker(db)
	receiverChecker := rabbitmq.NewHealthChecker(root.Consumer.RabbitMQ.Url, "rabbitmq_consumer")
	var healthHandler *health.HealthHandler
	var handler *mq.Handler
	if root.Publisher != nil {
		sender, er3 := rabbitmq.NewPublisherByConfig(*root.Publisher)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Publish, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Consumer.Config, userType, writer.Write, retryService.Retry, validator.Validate, nil, logError, logInfo)
		senderChecker := rabbitmq.NewHealthChecker(root.Publisher.Url, "rabbitmq_publisher")
		healthHandler = health.NewHealthHandler(mongoChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHealthHandler(mongoChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(userType, writer.Write, validator.Validate, root.Retry, true, logError, logInfo)
	}

	return &ApplicationContext{
		HealthHandler: healthHandler,
		Receive:       receiver.Consume,
		Handler:       handler,
	}, nil
}

func NewUserValidator() v.Validator {
	validator := v.NewDefaultValidator()
	validator.CustomValidateList = append(validator.CustomValidateList, v.CustomValidate{Fn: CheckActive, Tag: "active"})
	return validator
}
func CheckActive(fl validator.FieldLevel) bool {
	return fl.Field().Bool()
}
