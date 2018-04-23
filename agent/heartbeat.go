package agent

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/ehazlett/finca"
)

func (a *Agent) SendHeartbeat() error {
	w := finca.Worker{
		Name: a.config.Name,
		Addr: fmt.Sprintf("%s:%d", a.agentIP.String(), a.config.Port),
	}

	p := a.getUrl("/workers/heartbeat")
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}

	buf := bytes.NewBufferString(string(data))

	if _, err := http.Post(p, "application/octet-stream", buf); err != nil {
		return err
	}

	return nil
}
