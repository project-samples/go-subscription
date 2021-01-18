package main

import (
	"context"
	"github.com/common-go/config"
	"github.com/common-go/health"
	"github.com/common-go/log"
	"net/http"
	"strconv"

	c "go-service/internal"
)

func main() {
	var conf c.Root
	er1 := config.Load(&conf, "configs/config")
	if er1 != nil {
		panic(er1)
	}
	log.Initialize(conf.Log)
	ctx := context.Background()

	app, er2 := c.NewApp(ctx, conf)
	if er2 != nil {
		panic(er2)
	}

	go serve(ctx, conf.Server, app.HealthHandler)

	app.Consumer.Consume(ctx, app.ConsumerCaller)
}

// Start a http server to serve HTTP requests
func serve(ctx context.Context, config c.ServerConfig, healthHandler *health.HealthHandler) {
	server := ""
	if config.Port > 0 {
		server = ":" + strconv.Itoa(config.Port)
	}
	log.Info(ctx, "Start "+config.Name)
	http.HandleFunc("/health", healthHandler.Check)
	http.HandleFunc("/", healthHandler.Check)
	http.ListenAndServe(server, nil)
}
