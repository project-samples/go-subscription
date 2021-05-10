package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mongo"
	"github.com/core-go/mq"
	"github.com/core-go/mq/amq"
	"github.com/core-go/mq/log"
)

type Root struct {
	Server server.ServerConf `mapstructure:"server"`
	Log    log.Config        `mapstructure:"log"`
	Mongo  mongo.MongoConfig `mapstructure:"mongo"`
	Retry  *mq.RetryConfig   `mapstructure:"retry"`
	Amq    amq.Config        `mapstructure:"amq"`
}
