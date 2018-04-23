package agent

type Config struct {
	Name        string
	Addr        string
	ManagerAddr string
}

type Agent struct {
	config *Config
}

func NewAgent(c *Config) (*Agent, error) {
	return &Agent{
		config: c,
	}, nil
}
