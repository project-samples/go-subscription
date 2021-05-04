package app

import (
	"github.com/core-go/dynamodb"
	"github.com/core-go/health/server"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sarama"
)

type Root struct {
	Server      server.ServerConf   `mapstructure:"server"`
	Log         log.Config          `mapstructure:"log"`
	Dynamodb    dynamodb.Config     `mapstructure:"dynamodb"`
	Retry       *mq.RetryConfig     `mapstructure:"retry"`
	Reader      ReaderConfig        `mapstructure:"reader"`
	KafkaWriter *kafka.WriterConfig `mapstructure:"writer"`
}

type ReaderConfig struct {
	KafkaConsumer kafka.ReaderConfig `mapstructure:"kafka"`
	Config        mq.HandlerConfig   `mapstructure:"retry"`
}
