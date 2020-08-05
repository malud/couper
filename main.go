package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/sirupsen/logrus"

	"go.avenga.cloud/couper/gateway/command"
	"go.avenga.cloud/couper/gateway/config"
	"go.avenga.cloud/couper/gateway/server"
)

var (
	configFile = flag.String("f", "example.hcl", "-f ./couper.conf")
	listenPort = flag.Int("p", config.DefaultHTTP.ListenPort, "-p 8080")
)

func main() {
	// TODO: command / args
	if !flag.Parsed() {
		flag.Parse()
	}

	logger := newLogger()

	if *listenPort < 0 || *listenPort > 65535 {
		logger.Fatalf("Invalid listen port given: %d", *listenPort)
	}

	exampleConf := config.LoadFile(*configFile, logger)
	exampleConf.Addr = fmt.Sprintf(":%d", *listenPort)

	ctx := command.ContextWithSignal(context.Background())
	srv := server.New(ctx, logger, exampleConf)
	srv.Listen()
	<-ctx.Done() // TODO: shutdown deadline
}

func newLogger() *logrus.Entry {
	logger := logrus.New()
	logger.Out = os.Stdout
	logger.Formatter = &logrus.JSONFormatter{FieldMap: logrus.FieldMap{
		logrus.FieldKeyTime: "timestamp",
		logrus.FieldKeyMsg:  "message",
	}}
	logger.Level = logrus.DebugLevel
	return logger.WithField("type", "couper")
}
