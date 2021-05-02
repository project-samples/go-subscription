package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/rabbitmq"
)

type Root struct {
	Server    server.ServerConf         `mapstructure:"server"`
	Log       log.Config                `mapstructure:"log"`
	Mongo     mongo.MongoConfig         `mapstructure:"mongo"`
	Retry     *mq.RetryConfig           `mapstructure:"retry"`
	Consumer  ConsumerConfig            `mapstructure:"consumer"`
	Publisher *rabbitmq.PublisherConfig `mapstructure:"publisher"`
}

type ConsumerConfig struct {
	RabbitMQ rabbitmq.ConsumerConfig `mapstructure:"rabbitmq"`
	Config   mq.HandlerConfig        `mapstructure:"retry"`
}
