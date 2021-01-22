package app

import (
	"context"
	"reflect"

	"github.com/common-go/health"
	"github.com/common-go/kafka"
	"github.com/common-go/log"
	"github.com/common-go/mongo"
	"github.com/common-go/mq"
	v "github.com/common-go/validator"
	"github.com/sirupsen/logrus"
	"gopkg.in/go-playground/validator.v9"
)

type ApplicationContext struct {
	Consumer       mq.Consumer
	ConsumerCaller mq.ConsumerCaller
	HealthHandler  *health.HealthHandler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	mongoDb, er1 := mongo.SetupMongo(ctx, root.Mongo)
	if er1 != nil {
		log.Error(ctx, "Cannot connect to MongoDB: Error: "+er1.Error())
		return nil, er1
	}

	consumer, er2 := kafka.NewConsumerByConfig(root.KafkaConsumer, true)
	if er2 != nil {
		log.Error(ctx, "Cannot create a new consumer: Error: "+er2.Error())
		return nil, er2
	}
	logError := log.ErrorMsg
	var logInfo func(context.Context, string)
	if logrus.IsLevelEnabled(logrus.InfoLevel) {
		logInfo = log.InfoMsg
	}

	userTypeOf := reflect.TypeOf(User{})
	writer := mongo.NewMongoInserter(mongoDb, "users")
	v := NewUserValidator()
	validator := mq.NewValidator(userTypeOf, v)
	consumerCaller := mq.NewConsumerCaller(userTypeOf, writer, 3, nil, "", validator, nil, true, logError, logInfo)

	mongoChecker := mongo.NewHealthChecker(mongoDb)
	consumerChecker := kafka.NewDefaultKafkaHealthChecker(root.KafkaConsumer.Brokers)
	healthCheckers := []health.HealthChecker{mongoChecker, consumerChecker}
	healthHandler := health.NewHealthHandler(healthCheckers)
	return &ApplicationContext{
		Consumer:       consumer,
		ConsumerCaller: consumerCaller,
		HealthHandler:  healthHandler,
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
