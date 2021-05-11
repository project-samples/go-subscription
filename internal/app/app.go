package app

import (
	"context"
	"reflect"

	"github.com/core-go/health"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sqs"
	"github.com/core-go/mq/validator"
	v "github.com/go-playground/validator/v10"
)

type ApplicationContext struct {
	HealthHandler *health.HealthHandler
	Receive       func(ctx context.Context, handle func(context.Context, *mq.Message, error) error)
	Handler       *mq.Handler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er1 := mongo.SetupMongo(ctx, root.Mongo)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to MongoDB: Error: "+er1.Error())
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	client, er2 := sqs.Connect(root.Receiver.SQS)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new sqs. Error: "+er2.Error())
		return nil, er2
	}
	receiver,_ := sqs.NewReceiverByQueueName(client, root.Receiver.SQS.QueueName, true, 20, 1)

	userType := reflect.TypeOf(User{})
	writer := mongo.NewInserter(db, "user")
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	mongoChecker := mongo.NewHealthChecker(db)
	receiverChecker := sqs.NewHealthChecker(client, root.Receiver.SQS.QueueName, "sqs_receiver")
	var healthHandler *health.HealthHandler
	var handler *mq.Handler
	if root.Sender != nil {
		senderClient, er3 := sqs.Connect(*root.Sender)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sqs sender. Error:"+er3.Error())
			return nil, er3
		}
		var delaySecond int64
		delaySecond = 1
		sender,_ := sqs.NewSenderByQueueName(senderClient, root.Sender.QueueName, &delaySecond)
		retryService := mq.NewRetryService(sender.Send, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Receiver.Config, userType, writer.Write, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := sqs.NewHealthChecker(senderClient, root.Sender.QueueName, "sqs_sender")
		healthHandler = health.NewHealthHandler(mongoChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHealthHandler(mongoChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(userType, writer.Write, val.Validate, root.Retry, true, logError, logInfo)
	}

	return &ApplicationContext{
		HealthHandler: healthHandler,
		Receive:       receiver.Receive,
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
