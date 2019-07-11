package client

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/prometheus/common/log"
)

// Client is the elasticsearch task client
type Client interface {
	Tasks() ([]Task, error)
}

type client struct {
	baseURL string
}

// New creates a new elasticsearch client
func New(url string) Client {
	return &client{
		baseURL: url,
	}
}

func (c *client) Tasks() ([]Task, error) {
	log.Debugf("querying %s...", c.baseURL)
	var tasks []Task
	var result Result
	resp, err := http.Get(c.baseURL + "/_tasks?pretty&detailed=true&actions=*reindex")
	if err != nil {
		return tasks, errors.Wrap(err, "failed to get metrics")
	}
	if resp.StatusCode != 200 {
		return tasks, fmt.Errorf("failed to get metrics: %s", resp.Status)
	}
	defer resp.Body.Close()
	bts, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return tasks, errors.Wrap(err, "failed to parse metrics")
	}
	if err := json.Unmarshal(bts, &result); err != nil {
		return tasks, errors.Wrap(err, "failed to parse metrics")
	}
	for _, node := range result.Nodes {
		for _, task := range node.Tasks {
			tasks = append(tasks, task)
		}
	}
	return tasks, nil
}

type Result struct {
	Nodes map[string]struct {
		Tasks map[string]Task `json:"tasks"`
	} `json:"nodes"`
}

type Task struct {
	Status struct {
		Total   float64 `json:"total"`
		Updated float64 `json:"updated"`
		Created float64 `json:"created"`
		Deleted float64 `json:"deleted"`
		Batches float64 `json:"batches"`
	} `json:"status"`
	Description        string `json:"description"`
	StartTimeInMillis  int64  `json:"start_time_in_millis"`
	RunningTimeInNanos int64  `json:"running_time_in_nanos"`
}
