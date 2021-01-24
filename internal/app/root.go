package app

import (
	"github.com/common-go/config"
	"github.com/common-go/health"
	"github.com/common-go/kafka"
	"github.com/common-go/log"
	"github.com/common-go/mongo"
)

type Root struct {
	Server        health.ServerConfig  `mapstructure:"server"`
	Log           log.Config           `mapstructure:"log"`
	Mongo         mongo.MongoConfig    `mapstructure:"mongo"`
	KafkaConsumer kafka.ConsumerConfig `mapstructure:"kafka_consumer"`
	Retry         *config.Retry        `mapstructure:"retry"`
}
