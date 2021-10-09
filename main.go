package main

import (
	"context"
	"fmt"
	"github.com/core-go/config"
	"github.com/core-go/health/server"
	"go-service/internal/app"
)

func main() {
	var conf app.Root
	er1 := config.Load(&conf, "configs/config")
	if er1 != nil {
		panic(er1)
	}
	ctx := context.Background()

	applicationContext, er2 := app.NewApp(ctx, conf)
	if er2 != nil {
		panic(er2)
	}
	go applicationContext.Receive(ctx, applicationContext.Handler.Handle)

	err := server.Serve(conf.Server, applicationContext.HealthHandler.Check)
	if err != nil {
		fmt.Println(err)
		return
	}
}
