package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ehazlett/finca/version"
	"github.com/sirupsen/logrus"
)

func (m *Manager) apiIndex(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(version.FullVersion() + "\n"))
}

func (m *Manager) apiJobsUpload(w http.ResponseWriter, r *http.Request) {
	r.ParseMultipartForm(64 << 20)
	file, handler, err := r.FormFile("file")
	if err != nil {
		http.Error(w, fmt.Sprintf("invalid job: %s", err), http.StatusBadRequest)
		return
	}
	defer file.Close()

	jobFilename, err := m.uploadJob(handler.Filename, handler.Size, file)
	if err != nil {
		errStr := fmt.Sprintf("error saving job: %s", err)
		http.Error(w, errStr, http.StatusInternalServerError)
		logrus.Error(errStr)
		return
	}

	logrus.Debugf("job received: %s", handler.Filename)
	if err := m.QueueJob(jobFilename); err != nil {
		errStr := fmt.Sprintf("error queueing job: %s", err)
		http.Error(w, errStr, http.StatusInternalServerError)
		logrus.Error(errStr)
		return
	}
	logrus.Infof("job queued: %s", jobFilename)
	w.Write([]byte(fmt.Sprintf("job %s queued\n", handler.Filename)))
}

func (m *Manager) apiJobsCancel(w http.ResponseWriter, r *http.Request) {
	m.cancelRenderCh <- struct{}{}
	w.WriteHeader(http.StatusNoContent)
}

func (m *Manager) apiWorkers(w http.ResponseWriter, r *http.Request) {
	workers, err := m.Workers()
	if err != nil {
		errStr := fmt.Sprintf("error getting workers: %s", err)
		http.Error(w, errStr, http.StatusInternalServerError)
		logrus.Error(errStr)
		return
	}

	if err := json.NewEncoder(w).Encode(workers); err != nil {
		errStr := fmt.Sprintf("error serializing workers: %s", err)
		http.Error(w, errStr, http.StatusInternalServerError)
		logrus.Error(errStr)
		return
	}
}
