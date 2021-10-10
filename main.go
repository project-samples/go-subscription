package main

import (
	"context"
	"github.com/core-go/config"

	"go-service/internal/app"
	"go-service/pkg/server"
)

func main() {
	var conf app.Root
	er1 := config.Load(&conf, "configs/config")
	if er1 != nil {
		panic(er1)
	}
	ctx := context.Background()

	app, er2 := app.NewApp(ctx, conf)
	if er2 != nil {
		panic(er2)
	}

	go server.Serve(conf.Server, app.Check)
	app.Receive(ctx, app.Handle)
}
