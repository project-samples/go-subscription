package app

import (
	"context"
	"fmt"
	"github.com/core-go/health"
	"github.com/core-go/mq"
	"github.com/core-go/mq/ibm-mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/validator"
	"github.com/core-go/sql"
	v "github.com/go-playground/validator/v10"
	_ "github.com/go-sql-driver/mysql"
	"net/http"
	"reflect"
)

type ApplicationContext struct {
	Check   func(w http.ResponseWriter, r *http.Request)
	Receive func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handle  func(ctx context.Context, data []byte, header map[string]string, err error) error
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	fmt.Println(root)
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
	receiver, er2 := ibmmq.NewSimpleSubscriberByConfig(root.IBMMQ.SubscriberConfig, root.IBMMQ.MQAuth, log.ErrorMsg)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new receiver. Error: "+er2.Error())
	}
	userType := reflect.TypeOf(User{})
	writer := sql.NewInserter(db, "", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	newValidator := mq.NewValidator(userType, checker.Check)

	sqlChecker := sql.NewHealthChecker(db)
	receiverChecker := ibmmq.NewHealthCheckerByConfig(&root.IBMMQ.QueueConfig, &root.IBMMQ.MQAuth, root.IBMMQ.SubscriberConfig.Topic)

	healthHandler := health.NewHandler(sqlChecker, receiverChecker)
	handler := mq.NewHandlerWithRetryConfig(writer.Write, &userType, newValidator.Validate, root.Retry, true, nil, logError, logInfo)

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
