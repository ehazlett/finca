package main

import (
	"io/ioutil"
	"os"
	"strings"

	"github.com/ehazlett/finca/manager"
	"github.com/urfave/cli"
)

var managerCommand = cli.Command{
	Name:  "manager",
	Usage: "run finca manager",
	Flags: []cli.Flag{
		cli.StringFlag{
			Name:  "name, n",
			Usage: "manager name",
			Value: getNodeName(),
		},
		cli.StringFlag{
			Name:  "listen-addr, l",
			Usage: "listen address for manager",
			Value: ":8080",
		},
		cli.StringFlag{
			Name:  "redis-addr, r",
			Usage: "address to redis",
			Value: "127.0.0.1:6379",
		},
		cli.StringFlag{
			Name:  "s3-endpoint, e",
			Usage: "s3 endpoint",
			Value: "127.0.0.1:9000",
		},
		cli.StringFlag{
			Name:  "s3-access-key, k",
			Usage: "s3 access key",
			Value: "",
		},
		cli.StringFlag{
			Name:  "s3-secret-key, s",
			Usage: "s3 secret key",
			Value: "",
		},
		cli.StringFlag{
			Name:  "s3-region",
			Usage: "s3 region",
			Value: "us-east-1",
		},
		cli.BoolFlag{
			Name:  "s3-use-ssl",
			Usage: "enable SSL for s3 connections",
		},
	},
	Action: managerAction,
}

func getManagerConfig(c *cli.Context) (*manager.Config, error) {
	cfg := &manager.Config{
		Name:        c.String("name"),
		ListenAddr:  c.String("listen-addr"),
		RedisAddr:   c.String("redis-addr"),
		S3Endpoint:  c.String("s3-endpoint"),
		S3AccessKey: c.String("s3-access-key"),
		S3SecretKey: c.String("s3-secret-key"),
		S3UseSSL:    c.Bool("s3-use-ssl"),
	}
	// check for swarm secrets
	accessKeyPath := "/run/secrets/access_key"
	secretKeyPath := "/run/secrets/secret_key"
	if aInfo, _ := os.Stat(accessKeyPath); aInfo != nil {
		data, err := ioutil.ReadFile(accessKeyPath)
		if err != nil {
			return nil, err
		}

		cfg.S3AccessKey = strings.TrimSpace(string(data))
	}
	if sInfo, _ := os.Stat(secretKeyPath); sInfo != nil {
		data, err := ioutil.ReadFile(secretKeyPath)
		if err != nil {
			return nil, err
		}

		cfg.S3SecretKey = strings.TrimSpace(string(data))
	}

	return cfg, nil
}

func managerAction(c *cli.Context) error {
	cfg, err := getManagerConfig(c)
	if err != nil {
		return err
	}

	m, err := manager.NewManager(cfg)
	if err != nil {
		return err
	}

	return m.Run()
}
