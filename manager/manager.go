package manager

import (
	"fmt"
	"time"

	"github.com/gomodule/redigo/redis"
	minio "github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
)

const (
	workerKeyspace           = "/finca-workers"
	jobKeyspace              = "/finca-jobs"
	storageJobsBucketName    = "finca-jobs"
	storageResultsBucketName = "finca-results"
)

var (
	queueWatcherInterval = time.Second * 10
)

type Config struct {
	Name        string
	ListenAddr  string
	RedisAddr   string
	S3Endpoint  string
	S3Region    string
	S3AccessKey string
	S3SecretKey string
	S3UseSSL    bool
}

type Manager struct {
	config *Config
	pool   *redis.Pool
	mc     *minio.Client
}

func NewManager(c *Config) (*Manager, error) {
	mc, err := minio.New(c.S3Endpoint, c.S3AccessKey, c.S3SecretKey, c.S3UseSSL)
	if err != nil {
		return nil, err
	}
	// ensure buckets exist
	for _, b := range []string{storageJobsBucketName, storageResultsBucketName} {
		if err := mc.MakeBucket(b, c.S3Region); err != nil {
			exists, err := mc.BucketExists(b)
			if err != nil {
				return nil, err
			}
			if !exists {
				return nil, fmt.Errorf("unable to find or create bucket %s", b)
			}
		}
	}
	pool := &redis.Pool{
		MaxIdle:     3,
		IdleTimeout: 300 * time.Second,
		Dial:        func() (redis.Conn, error) { return redis.Dial("tcp", c.RedisAddr) },
	}
	logrus.Debugf("s3 endpoint: %s", c.S3Endpoint)
	logrus.Debugf("redis addr: %s", c.RedisAddr)
	return &Manager{
		config: c,
		pool:   pool,
		mc:     mc,
	}, nil
}
