package app

import (
	"context"
	"reflect"

	"github.com/core-go/firestore"
	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/kafka"
	"github.com/core-go/mq/log"
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

	receiver, er2 := kafka.NewReaderByConfig(root.Reader.KafkaConsumer, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := firestore.NewFirestoreWriter(client, "user", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	firestoreChecker := firestore.NewHealthChecker(ctx, []byte(root.Firestore.Credentials))
	receiverChecker := kafka.NewKafkaHealthChecker(root.Reader.KafkaConsumer.Brokers, "kafka_consumer")
	var healthHandler *health.Handler
	var handler *mq.Handler
	if root.KafkaWriter != nil {
		sender, er3 := kafka.NewWriterByConfig(*root.KafkaWriter)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Write, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Reader.Config, userType, writer.Write, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := kafka.NewKafkaHealthChecker(root.KafkaWriter.Brokers, "kafka_producer")
		healthHandler = health.NewHandler(firestoreChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(firestoreChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(userType, writer.Write, val.Validate, root.Retry, true, logError, logInfo)
	}

	return &ApplicationContext{
		HealthHandler: healthHandler,
		Receive:       receiver.Read,
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
