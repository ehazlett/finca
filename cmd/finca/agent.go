package main

import (
	"github.com/ehazlett/finca/agent"
	"github.com/urfave/cli"
)

var agentCommand = cli.Command{
	Name:  "agent",
	Usage: "run finca agent",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "agent name",
			Value: getNodeName(),
		},
		cli.StringFlag{
			Name:  "addr, a",
			Usage: "address for render agent",
			Value: "0.0.0.0:18018",
		},
		cli.StringFlag{
			Name:  "manager-url, m",
			Usage: "manager url",
			Value: "http://127.0.0.1:8080",
		},
	},
	Action: agentAction,
}

func agentAction(c *cli.Context) error {
	cfg := &agent.Config{
		Name:        c.String("name"),
		Addr:        c.String("addr"),
		ManagerAddr: c.String("manager-url"),
	}

	a, err := agent.NewAgent(cfg)
	if err != nil {
		return err
	}

	return a.Run()
}