package manager

import (
	"encoding/json"
	"fmt"

	"github.com/ehazlett/finca"
	"github.com/gomodule/redigo/redis"
)

func (m *Manager) UpdateWorker(w *finca.Worker) error {
	conn := m.pool.Get()
	defer conn.Close()

	workerKey := m.getWorkerKey(w.Name)
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

func (m *Manager) Workers() ([]*finca.Worker, error) {
	conn := m.pool.Get()
	defer conn.Close()

	workerKeys, err := redis.Strings(conn.Do("KEYS", m.getWorkerKey("*")))
	if err != nil {
		return nil, err
	}

	workers := []*finca.Worker{}
	for _, k := range workerKeys {
		data, err := redis.Bytes(conn.Do("GET", k))
		if err != nil {
			return nil, err
		}

		var w *finca.Worker
		if err := json.Unmarshal(data, &w); err != nil {
			return nil, err
		}

		workers = append(workers, w)
	}

	return workers, nil
}
