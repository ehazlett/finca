package manager

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ehazlett/finca"
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

func (m *Manager) apiJobsResult(w http.ResponseWriter, r *http.Request) {

}

func (m *Manager) apiWorkerHeartbeat(w http.ResponseWriter, r *http.Request) {
	var worker *finca.Worker
	if err := json.NewDecoder(r.Body).Decode(&worker); err != nil {
		http.Error(w, "invalid worker config", http.StatusBadRequest)
		return
	}

	if err := m.UpdateWorker(worker); err != nil {
		errStr := fmt.Sprintf("error updating worker: %s", err)
		http.Error(w, errStr, http.StatusInternalServerError)
		logrus.Error(errStr)
		return
	}
	logrus.Debugf("worker ping: %s", worker.Name)
}
