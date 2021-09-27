package app

import (
	"context"
	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/ibm-mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/validator"
	"github.com/core-go/sql"
	v "github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"reflect"
)

type ApplicationContext struct {
	HealthHandler *health.Handler
	Receive       func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handler       *mq.Handler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	db, er1 := sql.OpenByConfig(root.Sql)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to MySQL: Error: "+er1.Error())
		return nil, er1
	}

	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if log.IsInfoEnable() {
		logInfo = log.InfoMsg
	}

	receiver , er2 := ibmmq.NewSimpleSubscriberByConfig(root.IBMMQ.SubscriberConfig,root.IBMMQ.MQAuth)
	if er2 != nil {
		log.Error(ctx,"Cannot create a new receiver. Error: " + er2.Error())
	}
	userType := reflect.TypeOf(User{})
	writer := sql.NewInserter(db, "users", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	validator := mq.NewValidator(userType, checker.Check)

	sqlChecker := sql.NewHealthChecker(db)
	receiverChecker := ibmmq.NewHealthChecker(receiver.QueueManager, "ibmmq_subscriber")

	healthHandler := health.NewHandler(sqlChecker, receiverChecker)
	handler := mq.NewHandlerWithRetryConfig(writer.Write, &userType, validator.Validate, root.Retry, true, logError, logInfo)

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
