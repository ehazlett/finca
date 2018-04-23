package agent

import (
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/finca"
	"github.com/sirupsen/logrus"
)

func (a *Agent) Run() error {
	if err := a.startLuxconsole(); err != nil {
		return err
	}

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

	logrus.Infof("agent started on %s", a.agentIP.String())

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

func (a *Agent) startLuxconsole() error {
	luxPath, err := exec.LookPath("luxconsole")
	if err != nil {
		return err
	}

	cmd := exec.Command(luxPath, "-s")
	cmd.Stdout = os.Stdout

	a.luxCmd = cmd

	return cmd.Start()
}
