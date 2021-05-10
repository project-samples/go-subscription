package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sqs"
)

type Root struct {
	Server   server.ServerConf `mapstructure:"server"`
	Log      log.Config        `mapstructure:"log"`
	Mongo    mongo.MongoConfig `mapstructure:"mongo"`
	Retry    *mq.RetryConfig   `mapstructure:"retry"`
	Receiver ReceiverConfig    `mapstructure:"receiver"`
	Sender   *sqs.Config       `mapstructure:"sender"`
}

type ReceiverConfig struct {
	SQS    sqs.Config       `mapstructure:"sqs"`
	Config mq.HandlerConfig `mapstructure:"retry"`
}
