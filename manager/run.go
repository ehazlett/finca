package manager

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

func (m *Manager) Run() error {

	router := mux.NewRouter()
	router.HandleFunc("/", m.apiIndex)
	router.HandleFunc("/jobs/new", m.apiJobsUpload).Methods("POST")
	router.HandleFunc("/jobs/result", m.apiJobsResult).Methods("POST")
	router.HandleFunc("/workers/heartbeat", m.apiWorkerHeartbeat).Methods("POST")

	logrus.Infof("starting manager on %s", m.config.ListenAddr)

	go m.queueWatcher()

	return http.ListenAndServe(m.config.ListenAddr, router)
}
