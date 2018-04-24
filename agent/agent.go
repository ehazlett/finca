package agent

import (
	"net"
	"time"

	"github.com/gomodule/redigo/redis"
)

type Config struct {
	Name      string
	Interface string
	Port      int
	RedisAddr string
}

type Agent struct {
	config  *Config
	agentIP net.IP
	pool    *redis.Pool
}

func NewAgent(c *Config) (*Agent, error) {
	ip, err := getIP(c.Interface)
	if err != nil {
		return nil, err
	}
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", c.RedisAddr) },
	}
	return &Agent{
		config:  c,
		agentIP: ip,
		pool:    pool,
	}, nil
}
