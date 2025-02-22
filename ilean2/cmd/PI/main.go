package main

import (
	"iLean/agent"
	"iLean/config"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"
)

func main() {

	if err := run(); err != nil {
		logrus.Fatal(err)
	}

}

func run() error {

	logrus.Info("run application PI")

	var st time.Time
	defer func() {
		logrus.WithField("shutdown_time", time.Now().Sub(st)).Info("stopped")
	}()

	config, err := config.LoadConfig("config/pi.conf")

	if err != nil {
		logrus.Fatal("failed load config", err)
	}

	agent, err := agent.NewAgent(&config)

	if err != nil {
		return err
	}

	defer agent.Close()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	<-signals

	return nil
}
