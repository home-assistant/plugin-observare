package main

import (
	"log"
	"net/http"
)

func supervisorPing() bool {
	response, err := http.Get("http://supervisor/ping")
	if err != nil {
		log.Printf("Supervisor ping failed with error %s", err)
		return false
	}

	// Check response
	if response.StatusCode < 200 {
		return true
	}

	log.Printf("Supervisor ping failed with %d", response.StatusCode)
	return false
}
