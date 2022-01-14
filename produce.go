package main

import (
	"context"
	"encoding/json"
	lg "log"

	"github.com/core-go/config"
	"github.com/core-go/mq/log"
	"github.com/core-go/mq/sarama"
	"github.com/hamba/avro"

	"go-service/internal/app"
)

func main() {
	var conf app.Root
	er1 := config.Load(&conf, "configs/config")
	if er1 != nil {
		panic(er1)
	}

	ctx := context.Background()
	schema, er2 := avro.Parse(conf.Avro)
	if er2 != nil {
		log.Error(ctx, "Cannot create a schema. Error: "+er2.Error())
		panic(er2)
	}
	sender, er3 := kafka.NewWriterByConfig(*conf.KafkaWriter, nil)
	if er2 != nil {
		log.Error(ctx, "Cannot new a new sender. Error:"+er3.Error())
		panic(er3)
	}
	data := []byte(`{"id": "01", "username": "user01", "active": true, "locked": false, "dateOfBirth": "2009-12-31T23:59:59.999+07:00", "email": "test@gmail.com", "url": "http://test.com", "phone": "0987654321"}`)
	user := app.User{}
	er4 := json.Unmarshal(data, &user)
	if er4 != nil {
		log.Debug(ctx, er4)
		panic(er4)
	}
	d2, er5 := avro.Marshal(schema, &user)
	if er5 != nil {
		log.Debug(ctx, er5)
		panic(er5)
	}
	lg.Println(string(d2))
	sender.Write(ctx, d2, map[string]string{})
	u2 := app.User{}
	avro.Unmarshal(schema, d2, &u2)
	d3, _ := json.Marshal(&u2)
	lg.Println(string(d3))
}
