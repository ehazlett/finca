package manager

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/sirupsen/logrus"
)

func (m *Manager) QueueJob(name string) error {
	conn := m.pool.Get()
	defer conn.Close()

	jobKey := m.getJobNewKey(name)
	if _, err := conn.Do("SET", jobKey, m.config.Name); err != nil {
		return err
	}
	return nil
}

func (m *Manager) queueWatcher() {
	ticker := time.NewTicker(queueWatcherInterval)
	for range ticker.C {
		conn := m.pool.Get()
		key := m.getJobNewKey("*")
		jobs, err := redis.Strings(conn.Do("KEYS", key))
		if err != nil {
			logrus.Error(err)
			conn.Close()
			continue
		}

		if len(jobs) == 0 {
			conn.Close()
			continue
		}

		logrus.Debugf("jobs queued: %d", len(jobs))
		jobName := m.getJobNameFromKey(jobs[0])

		if err := m.ProcessJob(jobName); err != nil {
			logrus.Error(err)
			conn.Close()
			continue
		}

		conn.Close()
	}
}
