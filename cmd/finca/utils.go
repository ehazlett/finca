package main

import "os"

func getNodeName() string {
	h, err := os.Hostname()
	if err != nil {
		return "unknown"
	}

	return h
}
