package manager

import (
	"encoding/json"

	"github.com/ehazlett/finca"
	"github.com/gomodule/redigo/redis"
)

func (m *Manager) Workers() ([]*finca.Worker, error) {
	conn := m.pool.Get()
	defer conn.Close()

	workerKeys, err := redis.Strings(conn.Do("KEYS", finca.GetWorkerKey("*")))
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
