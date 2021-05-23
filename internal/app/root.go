package app

import (
	"github.com/core-go/firestore"
	"github.com/core-go/health/server"
	"github.com/core-go/mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/pubsub"
)

type Root struct {
	Server    server.ServerConf       `mapstructure:"server"`
	Log       log.Config              `mapstructure:"log"`
	Firestore firestore.Config        `mapstructure:"firestore"`
	Retry     *mq.RetryConfig         `mapstructure:"retry"`
	Sub       SubConfig               `mapstructure:"sub"`
	Pub       *pubsub.PublisherConfig `mapstructure:"pub"`
}

type SubConfig struct {
	Subscriber pubsub.SubscriberConfig `mapstructure:"pubsub"`
	Config     mq.HandlerConfig        `mapstructure:"retry"`
}
