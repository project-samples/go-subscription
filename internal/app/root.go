package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/kafka"
	"github.com/core-go/mq/zap"
)

type Root struct {
	Server      server.ServerConf   `mapstructure:"server"`
	Log         log.Config          `mapstructure:"log"`
	Mongo       mongo.MongoConfig   `mapstructure:"mongo"`
	Retry       *mq.RetryConfig     `mapstructure:"retry"`
	Reader      ReaderConfig        `mapstructure:"reader"`
	KafkaWriter *kafka.WriterConfig `mapstructure:"writer"`
}

type ReaderConfig struct {
	KafkaConsumer kafka.ReaderConfig `mapstructure:"kafka"`
	Config        mq.HandlerConfig   `mapstructure:"retry"`
}
