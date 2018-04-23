package manager

import (
	"fmt"
	"strings"

	"github.com/ehazlett/finca"
	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

func (m *Manager) RenderingJobs() ([]*finca.Job, error) {
	conn := m.pool.Get()
	defer conn.Close()

	keys := m.getJobRenderingKey("*")
	jobs := []*finca.Job{}

	jobKeys, err := redis.Strings(conn.Do("KEYS", keys))
	if err != nil {
		return nil, err
	}

	for _, k := range jobKeys {
		n := m.getJobNameFromKey(k)
		s, err := redis.String(conn.Do("GET", k))
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &finca.Job{
			Name:  n,
			Nodes: strings.Split(s, ","),
		})
	}

	return jobs, nil
}

func (m *Manager) ProcessJob(name string) error {
	conn := m.pool.Get()
	defer conn.Close()

	newKey := m.getJobNewKey(name)
	assignedKey := m.getJobRenderingKey(name)

	workers, err := m.Workers()
	if err != nil {
		return err
	}

	if len(workers) == 0 {
		return fmt.Errorf("no workers available")
	}

	renderingJobs, err := m.RenderingJobs()
	if err != nil {
		return err
	}

	if len(renderingJobs) > 0 {
		logrus.Debug("job is currently rendering; waiting")
		return nil
	}

	// TODO: start job on workers

	if _, err := conn.Do("DEL", newKey); err != nil {
		return err
	}

	nodes := []string{}
	for _, worker := range workers {
		nodes = append(nodes, worker.Name)
	}
	if _, err := conn.Do("SET", assignedKey, strings.Join(nodes, ",")); err != nil {
		return err
	}

	return nil
}
