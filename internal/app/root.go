package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/nats"
)

type Root struct {
	Server     server.ServerConf     `mapstructure:"server"`
	Log        log.Config            `mapstructure:"log"`
	Mongo      mongo.MongoConfig     `mapstructure:"mongo"`
	Retry      *mq.RetryConfig       `mapstructure:"retry"`
	Subscriber SubscriberConfig      `mapstructure:"subscriber"`
	Publisher  *nats.PublisherConfig `mapstructure:"publisher"`
}

type SubscriberConfig struct {
	NATS   nats.SubscriberConfig `mapstructure:"nats"`
	Config mq.HandlerConfig      `mapstructure:"retry"`
}
