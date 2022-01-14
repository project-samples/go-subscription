package app

import (
	"context"
	"net/http"
	"reflect"
	"strings"

	"github.com/core-go/health"
	mgo "github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/avro"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sarama"
	v "github.com/core-go/mq/validator"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	av "github.com/hamba/avro"
)

type ApplicationContext struct {
	Check   func(w http.ResponseWriter, r *http.Request)
	Receive func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handle  func(ctx context.Context, data []byte, header map[string]string, err error) error
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er0 := mgo.Setup(ctx, root.Mongo)
	if er0 != nil {
		log.Error(ctx, "Cannot connect to MongoDB: Error: "+er0.Error())
		return nil, er0
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	schema, er1 := av.Parse(root.Avro)
	if er1 != nil {
		log.Error(ctx, "Cannot create a schema. Error: "+er1.Error())
		return nil, er1
	}
	marshaller := avro.NewMarshaller(schema)
	receiver, er2 := kafka.NewSimpleReaderByConfig(root.Reader.KafkaConsumer, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := mgo.NewInserter(db, "user")
	checker := v.NewErrorChecker(NewUserValidator().Validate)
	validator := mq.NewValidator(userType, checker.Check, marshaller.Unmarshal)

	mongoChecker := mgo.NewHealthChecker(db)
	receiverChecker := kafka.NewKafkaHealthChecker(root.Reader.KafkaConsumer.Brokers, "kafka_consumer")
	var healthHandler *health.Handler
	var handler *mq.Handler
	if root.KafkaWriter != nil {
		sender, er3 := kafka.NewWriterByConfig(*root.KafkaWriter, nil, Generate)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Write, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Reader.Config, writer.Write, &userType, retryService.Retry, validator.Validate, nil, marshaller.Unmarshal, logError, logInfo)
		senderChecker := kafka.NewKafkaHealthChecker(root.KafkaWriter.Brokers, "kafka_producer")
		healthHandler = health.NewHandler(mongoChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(mongoChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(writer.Write, &userType, validator.Validate, root.Retry, true, nil, marshaller.Unmarshal, logError, logInfo)
	}

	return &ApplicationContext{
		Check:   healthHandler.Check,
		Receive: receiver.Read,
		Handle:  handler.Handle,
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
func Generate() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
