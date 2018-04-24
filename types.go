package finca

type Job struct {
	Name  string
	Nodes []string
}

type Worker struct {
	Name string
	Addr string
}

type WorkerMessage struct {
	CancelJob bool
}
