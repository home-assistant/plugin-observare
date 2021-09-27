package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/pkg/stdcopy"
)

type SupervisorResponse struct {
	Result  string                 `json:"result"`
	Message string                 `json:"message,omitempty"`
	Data    map[string]interface{} `json:"data,omitempty"`
}

type SupervisorInfo struct {
	Healthy   bool `json:"healthy"`
	Supported bool `json:"supported"`
}

func supervisorApiProxy(path string) (SupervisorResponse, error) {
	var jsonResponse SupervisorResponse
	request, _ := http.NewRequest("GET", fmt.Sprintf("http://supervisor/%s", path), nil)
	request.Header = http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", os.Getenv("SUPERVISOR_TOKEN"))},
	}

	response, err := httpClient.Do(request)
	if err != nil {
		log.Printf("Supervisor API call failed with error %s", err)
		return jsonResponse, err
	}

	if response.StatusCode >= 300 {
		log.Printf("Supervisor API call failed with status code %v", response.StatusCode)
		return jsonResponse, err
	}

	bodyBytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return jsonResponse, err
	}

	defer response.Body.Close()

	json.Unmarshal([]byte(bodyBytes), &jsonResponse)
	return jsonResponse, err
}

func supervisorPing() bool {
	_, err := supervisorApiProxy("supervisor/ping")
	if err != nil {
		log.Printf("Supervisor ping failed with error %s", err)
		return false
	}
	return true
}

func getSupervisorInfo() (SupervisorInfo, error) {
	var supervisorInfo SupervisorInfo
	response, err := supervisorApiProxy("supervisor/info")
	if err != nil {
		log.Printf("Supervisor API call failed with error %s", err)
		return supervisorInfo, err
	}

	jsonData, _ := json.Marshal(response.Data)
	json.Unmarshal(jsonData, &supervisorInfo)

	return supervisorInfo, err
}

func supervisorLogs(w io.Writer) error {
	// Read logs from container
	reader, err := cli.ContainerLogs(context.Background(), "hassio_supervisor", types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
		Timestamps: false,
		Tail:       "all",
	})

	if err != nil {
		log.Printf("Can't get supervisor logs %s", err)
		return err
	}
	defer reader.Close()

	// Return the content
	_, err = stdcopy.StdCopy(w, w, reader)
	return err
}

func supervisorRestart() error {
	// Read logs from container
	err := cli.ContainerStop(context.Background(), "hassio_supervisor", nil)
	if err != nil {
		log.Printf("Can't stop supervisor: %s", err)
		return err
	}

	return nil
}
