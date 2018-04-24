package agent

import (
	"encoding/json"
	"os"
	"os/exec"
	"os/signal"
	"syscall"
	"time"

	"github.com/ehazlett/finca"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

func (a *Agent) Run() error {
	// TODO: subscribe to channel to stop when needed
	cmd, err := a.startLuxconsole()
	if err != nil {
		return err
	}

	c := a.pool.Get()
	defer c.Close()

	psc := redis.PubSubConn{Conn: c}
	psc.Subscribe(finca.WorkerChannel)

	go func() {
		for {
			switch v := psc.Receive().(type) {
			case redis.Message:
				var msg *finca.WorkerMessage
				if err := json.Unmarshal(v.Data, &msg); err != nil {
					logrus.Errorf("error deserializing worker message: %s", err)
					continue
				}

				if msg.CancelJob {
					if err := a.stopLuxconsole(cmd); err != nil {
						logrus.Errorf("error stopping luxconsole: %s", err)
						continue
					}

					l, err := a.startLuxconsole()
					if err != nil {
						logrus.Errorf("error starting luxconsole: %s", err)
						continue
					}

					cmd = l
				}
			}
		}
	}()

	t := time.NewTicker(finca.WorkerTimeout / 2)
	go func() {
		for range t.C {
			if err := a.UpdateWorkerInfo(); err != nil {
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

func (a *Agent) startLuxconsole() (*exec.Cmd, error) {
	luxPath, err := exec.LookPath("luxconsole")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(luxPath, "-s", "-l")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	return cmd, nil
}

func (a *Agent) stopLuxconsole(cmd *exec.Cmd) error {
	if err := cmd.Process.Kill(); err != nil {
		return err
	}
	cmd.Wait()
	return nil
}
