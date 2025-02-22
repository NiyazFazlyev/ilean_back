package main

import (
	"errors"
	"github.com/sirupsen/logrus"
	"iLean/server"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	if err := run(); err != nil {
		logrus.Fatal(err)
	}
}

func run() error {

	serverPort := os.Getenv("SERVER_WEBSOCKET_PORT")
	if serverPort == "" {
		return errors.New("websocket port must be specified")
	}
	server, err := server.NewServer(":4000", "key")

	if err != nil {
		return err
	}

	defer server.Stop()

	signals := make(chan os.Signal)
	signal.Notify(signals, os.Interrupt, syscall.SIGTERM)

	logrus.Infof("captured %v signal, stopping", <-signals)

	return nil
}
