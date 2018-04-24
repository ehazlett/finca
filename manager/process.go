package manager

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/ehazlett/finca"
	"github.com/gomodule/redigo/redis"
	"github.com/mholt/archiver"
	minio "github.com/minio/minio-go"
	"github.com/sirupsen/logrus"
)

func (m *Manager) RenderingJobs() ([]*finca.Job, error) {
	conn := m.pool.Get()
	defer conn.Close()

	keys := m.getJobRenderingKey("*")
	jobs := []*finca.Job{}

	jobKeys, err := redis.Strings(conn.Do("KEYS", keys))
	if err != nil {
		return nil, err
	}

	for _, k := range jobKeys {
		n := m.getJobNameFromKey(k)
		s, err := redis.String(conn.Do("GET", k))
		if err != nil {
			return nil, err
		}
		jobs = append(jobs, &finca.Job{
			Name:  n,
			Nodes: strings.Split(s, ","),
		})
	}

	return jobs, nil
}

func (m *Manager) setJobRendering(name string, nodes []string) error {
	conn := m.pool.Get()
	defer conn.Close()

	newKey := m.getJobNewKey(name)
	assignedKey := m.getJobRenderingKey(name)

	if _, err := conn.Do("DEL", newKey); err != nil {
		return err
	}

	if _, err := conn.Do("SET", assignedKey, strings.Join(nodes, ",")); err != nil {
		return err
	}

	return nil
}

func (m *Manager) setJobComplete(name string) error {
	conn := m.pool.Get()
	defer conn.Close()

	assignedKey := m.getJobRenderingKey(name)
	completeKey := m.getJobCompleteKey(name)

	if _, err := conn.Do("DEL", assignedKey); err != nil {
		return err
	}

	if _, err := conn.Do("SET", completeKey, time.Now().Format(time.RFC3339)); err != nil {
		return err
	}

	return nil
}

func (m *Manager) ProcessJob(name string) error {
	workers, err := m.Workers()
	if err != nil {
		return err
	}

	if len(workers) == 0 {
		return fmt.Errorf("no workers available")
	}

	renderingJobs, err := m.RenderingJobs()
	if err != nil {
		return err
	}

	if len(renderingJobs) > 0 {
		logrus.Debug("a job is currently rendering; waiting until complete")
		return nil
	}

	nodes := []string{}
	nodeAddrs := []string{}
	for _, worker := range workers {
		nodes = append(nodes, worker.Name)
		nodeAddrs = append(nodeAddrs, worker.Addr)
	}

	if err := m.setJobRendering(name, nodeAddrs); err != nil {
		return err
	}

	tmpFile, err := ioutil.TempFile("", "finca-")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	// TODO: download job and unzip to temp
	obj, err := m.mc.GetObject(storageJobsBucketName, name, minio.GetObjectOptions{})
	if err != nil {
		return err
	}

	if _, err := io.Copy(tmpFile, obj); err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "finca-render-")
	if err != nil {
		return err
	}

	if err := archiver.Zip.Open(tmpFile.Name(), tmpDir); err != nil {
		return err
	}

	if err := m.startLuxconsole(name, tmpDir, nodeAddrs); err != nil {
		return err
	}

	if err := m.setJobComplete(name); err != nil {
		return err
	}

	logrus.Infof("render job %s complete", name)

	return nil
}

func (m *Manager) startLuxconsole(jobName, workingDir string, nodes []string) error {
	// iterate over files in working dir to find .lxs
	files, err := ioutil.ReadDir(workingDir)
	if err != nil {
		return err
	}
	defer os.RemoveAll(workingDir)

	projectFile := ""
	for _, f := range files {
		if filepath.Ext(f.Name()) == ".lxs" {
			projectFile = f.Name()
			break
		}
	}

	if projectFile == "" {
		return fmt.Errorf("unable to find luxrender scene")
	}

	luxPath, err := exec.LookPath("luxconsole")
	if err != nil {
		return err
	}

	// TODO: specify output filename for upload
	opts := []string{
		"-i30",
	}

	for _, node := range nodes {
		opts = append(opts, fmt.Sprintf("-u%s", node))
	}

	opts = append(opts, projectFile)

	logrus.Infof("starting render for %s on %s", projectFile, strings.Join(nodes, ","))

	cmd := exec.Command(luxPath, opts...)
	cmd.Dir = workingDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Start(); err != nil {
		return err
	}

	go func() {
		for {
			select {
			case <-time.After(time.Second * 10):
				if err := m.uploadRenderResults(jobName, workingDir); err != nil {
					logrus.Errorf("error uploading results: %s", err)
				}
			case <-m.cancelRenderCh:
				logrus.Warnf("signal received; cancelling job %s", jobName)
				if err := m.stopRender(cmd, jobName, workingDir); err != nil {
					logrus.Errorf("error stopping render: %s", err)
				}
				return
			case <-time.After(m.config.RenderTimeout):
				logrus.Warnf("timeout reached (%s) while rendering; stopping", m.config.RenderTimeout.String())
				if err := m.stopRender(cmd, jobName, workingDir); err != nil {
					logrus.Errorf("error stopping render: %s", err)
				}
				return
			}
		}
	}()

	logrus.Debug("waiting for rendering to complete")
	cmd.Wait()

	return nil
}

func (m *Manager) stopRender(cmd *exec.Cmd, jobName, workingDir string) error {
	// send message to redis for workers to stop rendering
	conn := m.pool.Get()
	defer conn.Close()

	msg := finca.WorkerMessage{
		CancelJob: true,
	}

	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := conn.Do("PUBLISH", finca.WorkerChannel, string(data)); err != nil {
		return err
	}

	if err := m.uploadRenderResults(jobName, workingDir); err != nil {
		return err
	}
	if err := cmd.Process.Kill(); err != nil {
		return err
	}
	return nil
}

func (m *Manager) uploadRenderResults(jobName, workingDir string) error {
	// iterate over files in working dir to find .png
	files, err := ioutil.ReadDir(workingDir)
	if err != nil {
		return err
	}

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".png" {
			src := filepath.Join(workingDir, f.Name())
			dest := path.Join(jobName, f.Name())
			logrus.Debugf("uploading result image %s", f.Name())
			if err := m.uploadRender(src, dest); err != nil {
				return err
			}
		}
	}

	return nil
}
