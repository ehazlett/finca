package agent

func (a *Agent) getUrl(p string) string {
	return a.config.ManagerAddr + p
}
