package app

import (
	"github.com/core-go/health/server"
	"github.com/core-go/mq"
	"github.com/core-go/mq/ibm-mq"
	"github.com/core-go/mq/log"
	"github.com/core-go/sql"
)

type Root struct {
	Server server.ServerConf `mapstructure:"server"`
	Log    log.Config        `mapstructure:"log"`
	Sql    sql.Config        `mapstructure:"sql"`
	Retry  *mq.RetryConfig   `mapstructure:"retry"`
	IBMMQ  IBMMQConfig       `mapstructure:"ibmmq"`
}

type IBMMQConfig struct {
	QueueConfig      ibmmq.QueueConfig      `mapstructure:"queue_config"`
	SubscriberConfig ibmmq.SubscriberConfig `mapstructure:"subscriber_config"`
	MQAuth           ibmmq.MQAuth           `mapstructure:"mq_auth"`
}
