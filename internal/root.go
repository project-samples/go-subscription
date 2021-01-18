package internal

import (
	"github.com/common-go/kafka"
	"github.com/common-go/log"
	"github.com/common-go/mongo"
)

type Root struct {
	Server        ServerConfig         `mapstructure:"server"`
	Log           log.Config           `mapstructure:"log"`
	Mongo         mongo.MongoConfig    `mapstructure:"mongo"`
	KafkaConsumer kafka.ConsumerConfig `mapstructure:"kafka_consumer"`
}

type ServerConfig struct {
	Name string `mapstructure:"name"`
	Port int    `mapstructure:"port"`
}
