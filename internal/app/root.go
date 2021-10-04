package app

import (
	"github.com/core-go/mongo"

	"go-service/pkg/kafka"
	"go-service/pkg/log"
	"go-service/pkg/mq"
	"go-service/pkg/server"
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
