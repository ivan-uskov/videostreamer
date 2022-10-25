package main

import (
	"videostreamer/src/common/cmd"
	"videostreamer/src/common/transport"
)

type config struct {
	SID      string `envconfig:"sid"`
	SfuAddr  string `envconfig:"sfu_addr"`
	LogLevel string `envconfig:"log_level"`
	transport.MediaDevicesConfig
	transport.PeerConfig
}

func main() {
	var conf config
	cmd.ParseConfig(&conf)
	cmd.SetupLogger(conf.LogLevel)

	killSignalChan := cmd.GetKillSignalChan()

	client := newSfuClient(conf)
	defer client.Close()

	cmd.WaitForKillSignal(killSignalChan)
}
