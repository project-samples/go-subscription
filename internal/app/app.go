package app

import (
	"context"
	"reflect"
	"strings"

	"github.com/core-go/cassandra"
	"github.com/core-go/health"
	cas "github.com/core-go/health/cassandra"
	"github.com/core-go/mq"
	"github.com/core-go/mq/kafka"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/validator"
	v "github.com/go-playground/validator/v10"
	"github.com/gocql/gocql"
	"github.com/google/uuid"
)

type ApplicationContext struct {
	HealthHandler *health.Handler
	Receive       func(ctx context.Context, handle func(context.Context, []byte, map[string]string, error) error)
	Handler       *mq.Handler
}

func NewApp(ctx context.Context, root Root) (*ApplicationContext, error) {
	log.Initialize(root.Log)
	cluster := gocql.NewCluster(root.Cassandra.Uri)
	cluster.Authenticator = gocql.PasswordAuthenticator{
		Username: root.Cassandra.Username,
		Password: root.Cassandra.Password,
	}
	session, er1 := cluster.CreateSession()
	if er1 != nil {
		log.Error(ctx, "Cannot connect to Cassandra, Error: " + er1.Error())
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
	writer := cassandra.NewCassandraWriter(session, "user.user", userType)
	checker := validator.NewErrorChecker(NewUserValidator().Validate)
	val := mq.NewValidator(userType, checker.Check)

	cassandraChecker := cas.NewHealthChecker(cluster)
	receiverChecker := kafka.NewKafkaHealthChecker(root.Reader.KafkaConsumer.Brokers, "kafka_consumer")
	var healthHandler *health.Handler
	var handler *mq.Handler
	if root.KafkaWriter != nil {
		sender, er3 := kafka.NewWriterByConfig(*root.KafkaWriter, Generate)
		if er3 != nil {
			log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
			return nil, er3
		}
		retryService := mq.NewRetryService(sender.Write, logError, logInfo)
		handler = mq.NewHandlerByConfig(root.Reader.Config, writer.Write, &userType, retryService.Retry, val.Validate, nil, logError, logInfo)
		senderChecker := kafka.NewKafkaHealthChecker(root.KafkaWriter.Brokers, "kafka_producer")
		healthHandler = health.NewHandler(cassandraChecker, receiverChecker, senderChecker)
	} else {
		healthHandler = health.NewHandler(cassandraChecker, receiverChecker)
		handler = mq.NewHandlerWithRetryConfig(writer.Write, &userType, val.Validate, root.Retry, true, logError, logInfo)
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
func Generate() string {
	return strings.Replace(uuid.New().String(), "-", "", -1)
}
