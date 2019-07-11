package collector

import (
	"regexp"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/common/log"
	"github.com/totvslabs/elasticsearch-reindex-exporter/client"
)

const namespace = "elasticsearch"
const subsystem = "reindex"

var labels = []string{"index"}

// NewCollector collector
func NewCollector(client client.Client) prometheus.Collector {
	return &collector{
		client: client,

		// default metrics
		up: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "up"),
			"SGW admin API is responding",
			nil,
			nil,
		),
		scrapeDuration: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "scrape_duration_seconds"),
			"Scrape duration in seconds",
			nil,
			nil,
		),
		total: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "total"),
			"Total docs to reindex",
			labels,
			nil,
		),
		updated: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "updated"),
			"Docs updated",
			labels,
			nil,
		),
		deleted: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "deleted"),
			"Docs deleted",
			labels,
			nil,
		),
		created: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "created"),
			"Docs reindexed",
			labels,
			nil,
		),
		running: prometheus.NewDesc(
			prometheus.BuildFQName(namespace, subsystem, "running_seconds"),
			"Time running in seconds",
			labels,
			nil,
		),
	}
}

type collector struct {
	mutex  sync.Mutex
	client client.Client

	up             *prometheus.Desc
	scrapeDuration *prometheus.Desc
	total          *prometheus.Desc
	updated        *prometheus.Desc
	created        *prometheus.Desc
	deleted        *prometheus.Desc
	running        *prometheus.Desc
}

// Describe all metrics
func (c *collector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.up
	ch <- c.scrapeDuration
	ch <- c.total
	ch <- c.updated
	ch <- c.created
	ch <- c.deleted
	ch <- c.running
}

// Collect all metrics
func (c *collector) Collect(ch chan<- prometheus.Metric) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	start := time.Now()
	defer func() {
		ch <- prometheus.MustNewConstMetric(c.scrapeDuration, prometheus.GaugeValue, time.Since(start).Seconds())
	}()

	log.Info("Collecting ES Reindex metrics...")
	tasks, err := c.client.Tasks()
	if err != nil {
		ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
		log.With("error", err).Error("failed to scrape ES")
		return
	}

	for _, task := range tasks {
		var match = indexNameRE.FindStringSubmatch(task.Description)
		if len(match) < 1 {
			ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 0)
			log.Errorf("couldn't extract index name from %s", task.Description)
			return
		}
		var index = match[1]
		ch <- prometheus.MustNewConstMetric(c.total, prometheus.GaugeValue, task.Status.Total, index)
		ch <- prometheus.MustNewConstMetric(c.updated, prometheus.GaugeValue, task.Status.Updated, index)
		ch <- prometheus.MustNewConstMetric(c.created, prometheus.GaugeValue, task.Status.Created, index)
		ch <- prometheus.MustNewConstMetric(c.deleted, prometheus.GaugeValue, task.Status.Deleted, index)
		ch <- prometheus.MustNewConstMetric(c.running, prometheus.GaugeValue, time.Duration(task.RunningTimeInNanos).Seconds(), index)
	}

	ch <- prometheus.MustNewConstMetric(c.up, prometheus.GaugeValue, 1)
}

// reindex from [staging-index-6d064080943311e9b3cc42010a840084] to [staging-index-6d064080943311e9b3cc42010a840084-new1]
var indexNameRE = regexp.MustCompile("reindex from \\[(.*)\\] to .*")
