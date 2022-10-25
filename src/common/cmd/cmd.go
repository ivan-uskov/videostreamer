package cmd

import (
	"github.com/kelseyhightower/envconfig"
	log "github.com/sirupsen/logrus"
	"os"
	"os/signal"
	"syscall"
)

func ParseConfig(v interface{}) {
	if err := envconfig.Process("", v); err != nil {
		log.WithError(err).Fatal("config parse error")
	}
}

func SetupLogger(lvl string) {
	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(os.Stdout)

	l, err := log.ParseLevel(lvl)
	if err != nil {
		log.WithError(err).Fatal()
	}

	log.SetLevel(l)
}

func GetKillSignalChan() chan os.Signal {
	osKillSignalChan := make(chan os.Signal, 1)
	signal.Notify(osKillSignalChan, os.Interrupt, syscall.SIGTERM)
	return osKillSignalChan
}

func WaitForKillSignal(ch <-chan os.Signal) {
	sig := <-ch
	switch sig {
	case os.Interrupt:
		log.Info("get SIGINT")
	case syscall.SIGTERM:
		log.Info("got SIGTERM")
	}
}
