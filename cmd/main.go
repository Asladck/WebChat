package main

import (
	"github.com/sirupsen/logrus"
	ws "websckt/internal/ws_server"
)

const (
	addr = "`192.168.0.14:9090`"
)

func main() {
	wsSrv := ws.NewWsServer(addr)
	logrus.Info("Started ws server")
	if err := wsSrv.Start(); err != nil {
		logrus.Fatalf("Error with ws server: %v", err)
	}
}
