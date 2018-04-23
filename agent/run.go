package agent

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/finca"
	"github.com/sirupsen/logrus"
)

func (a *Agent) Run() error {
	t := time.NewTicker(finca.WorkerTimeout / 2)
	go func() {
		for range t.C {
			logrus.Debug("heartbeat")
			if err := a.SendHeartbeat(); err != nil {
				logrus.Error(err)
				continue
			}
		}
	}()

	signals := make(chan os.Signal, 32)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)

	for {
		select {
		case s := <-signals:
			switch s {
			case syscall.SIGINT:
				return nil
			}
		}
	}

	return nil
}
