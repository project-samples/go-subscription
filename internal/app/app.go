package app

import (
	"context"
	"net/http"
	"reflect"

	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sarama"
	"github.com/core-go/mq/validator"
	"github.com/core-go/sql"
	v "github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
)

type ApplicationContext struct {
	Check   func(w http.ResponseWriter, r *http.Request)
	Receive func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handle  func(ctx context.Context, data []byte, header map[string]string, err error) error
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er1 := sql.OpenByConfig(root.Sql)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to SQL: Error: "+er1.Error())
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver, er2 := kafka.NewSimpleReaderByConfig(root.Reader.KafkaConsumer, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := sql.NewSqlWriter(db, "users", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	sqlChecker := sql.NewHealthChecker(db)
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
		handler = mq.NewHandlerByConfig(root.Reader.Config, writer.Write, &userType, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := kafka.NewKafkaHealthChecker(root.KafkaWriter.Brokers, "kafka_producer")
		healthHandler = health.NewHandler(sqlChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(sqlChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(writer.Write, &userType, val.Validate, root.Retry, true, nil, logError, logInfo)
	}

	return &ApplicationContext{
		Check:   healthHandler.Check,
		Receive: receiver.Read,
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
