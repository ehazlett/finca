package finca

import (
	"path"
	"time"
)

var (
	WorkerTimeout  = time.Second * 30
	WorkerChannel  = "finca-workers"
	WorkerKeyspace = "/finca-workers"
)

func GetWorkerKey(name string) string {
	return path.Join(WorkerKeyspace, name)
}
