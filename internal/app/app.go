package app

import (
	"context"
	"reflect"

	"github.com/core-go/health"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/amq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/validator"
	v "github.com/go-playground/validator/v10"
	"github.com/go-stomp/stomp"
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

	receiver, er2 := amq.NewSubscriberByConfig(root.Amq, stomp.AckAuto, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
		return nil, er2
	}
	userType := reflect.TypeOf(User{})
	writer := mongo.NewInserter(db, "user")
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	mongoChecker := mongo.NewHealthChecker(db)
	receiverChecker := amq.NewHealthChecker(receiver.Conn, "amq_subscriber")
	handler := mq.NewHandlerWithRetryConfig(userType, writer.Write, val.Validate, root.Retry, true, logError, logInfo)
	healthHandler := health.NewHealthHandler(mongoChecker, receiverChecker)

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
