package agent

import (
	"encoding/json"
	"fmt"

	"github.com/ehazlett/finca"
)

func (a *Agent) Addr() string {
	return fmt.Sprintf("%s:%d", a.agentIP.String(), a.config.Port)
}

func (a *Agent) UpdateWorkerInfo() error {
	conn := a.pool.Get()
	defer conn.Close()

	w := &finca.Worker{
		Name: a.config.Name,
		Addr: a.Addr(),
	}
	workerKey := finca.GetWorkerKey(a.config.Name)
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}
	if _, err := conn.Do("SET", workerKey, string(data)); err != nil {
		return err
	}
	if _, err := conn.Do("EXPIRE", workerKey, fmt.Sprintf("%d", int(finca.WorkerTimeout.Seconds()))); err != nil {
		conn.Do("DEL", workerKey)
		return err
	}

	return nil
}
