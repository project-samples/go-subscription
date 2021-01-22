package main

import (
	"context"
	"github.com/common-go/config"
	"github.com/common-go/health"
	logger "github.com/common-go/log"
	"log"
	"net/http"
	"strconv"

	"go-service/internal/app"
)

func main() {
	var conf app.Root
	er1 := config.Load(&conf, "configs/config")
	if er1 != nil {
		panic(er1)
	}
	logger.Initialize(conf.Log)
	ctx := context.Background()

	app, er2 := app.NewApp(ctx, conf)
	if er2 != nil {
		panic(er2)
	}

	go serve(conf.Server, app.HealthHandler)

	app.Consumer.Consume(ctx, app.ConsumerCaller)
}

// Start a http server to serve HTTP requests
func serve(config app.ServerConfig, healthHandler *health.HealthHandler) {
	server := ""
	if config.Port > 0 {
		server = ":" + strconv.Itoa(config.Port)
	}
	log.Println("Start " + config.Name)
	http.HandleFunc("/health", healthHandler.Check)
	http.HandleFunc("/", healthHandler.Check)
	http.ListenAndServe(server, nil)
}
