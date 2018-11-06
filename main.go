package main

import (
	"github.com/dayan-be/access-service/logic"
	"github.com/sirupsen/logrus"
	_ "github.com/dayan-be/golibs/log"
)

func main() {

	//1.load configer
	cfg := logic.Config()
	cfg.Load()

	//2.log
	logrus.Info("aaa")

}
