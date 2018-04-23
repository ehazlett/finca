package agent

import (
	"fmt"
	"net"
)

func (a *Agent) getUrl(p string) string {
	return a.config.ManagerAddr + p
}

func getIP(ifaceName string) (net.IP, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}
	for _, i := range ifaces {
		if i.Name != ifaceName {
			continue
		}
		addrs, err := i.Addrs()
		if err != nil {
			return nil, err
		}
		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				return v.IP, nil
			default:
				fmt.Println(v)
			}
		}
	}

	return nil, fmt.Errorf("unable to detect IP address for interface %s", ifaceName)
}
