package manager

import "path"

func (m *Manager) getWorkerKey(name string) string {
	return path.Join(workerKeyspace, name)
}

func (m *Manager) getJobNewKey(name string) string {
	return path.Join(jobKeyspace, "new", name)
}

func (m *Manager) getJobRenderingKey(name string) string {
	return path.Join(jobKeyspace, "rendering", name)
}

func (m *Manager) getJobCompleteKey(name string) string {
	return path.Join(jobKeyspace, "complete", name)
}

func (m *Manager) getJobNameFromKey(key string) string {
	return path.Base(key)
}
