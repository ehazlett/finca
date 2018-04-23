package agent

import (
	"net"
	"os/exec"
)

type Config struct {
	Name        string
	Interface   string
	Port        int
	ManagerAddr string
}

type Agent struct {
	config  *Config
	agentIP net.IP
	luxCmd  *exec.Command
}

func NewAgent(c *Config) (*Agent, error) {
	ip, err := getIP(c.Interface)
	if err != nil {
		return nil, err
	}

	return &Agent{
		config:  c,
		agentIP: ip,
	}, nil
}
