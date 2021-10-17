package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	kafka "github.com/core-go/mq/confluent"
	"github.com/core-go/mq/log"
)

type Root struct {
	Server      server.ServerConf     `mapstructure:"server"`
	Log         log.Config            `mapstructure:"log"`
	Mongo       mongo.MongoConfig     `mapstructure:"mongo"`
	Retry       *mq.RetryConfig       `mapstructure:"retry"`
	Reader      ReaderConfig          `mapstructure:"reader"`
	KafkaWriter *kafka.ProducerConfig `mapstructure:"writer"`
}

type ReaderConfig struct {
	KafkaConsumer kafka.ConsumerConfig `mapstructure:"kafka"`
	Config        mq.HandlerConfig     `mapstructure:"retry"`
}
