package main

import (
	"github.com/micro/go-micro"
	"time"
	"github.com/dayan-be/access-service/proto"
	"github.com/dayan-be/access-service/logic/access-micro"
	"github.com/micro/go-log"
)

func main() {
	service := micro.NewService(
		micro.Name("go.micro.srv.greeter"),
		micro.RegisterTTL(time.Second*30),
		micro.RegisterInterval(time.Second*10),
	)

	// optionally setup command line usage
	service.Init()

	// Register Handlers
	access.RegisterPushHandler(service.Server(), new(logic.server))

	// Run server
	if err := service.Run(); err != nil {
		log.Fatal(err)
	}
}
