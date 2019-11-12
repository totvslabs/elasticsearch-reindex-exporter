package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"github.com/totvslabs/elasticsearch-reindex-exporter/client"
	"github.com/totvslabs/elasticsearch-reindex-exporter/collector"
	kingpin "gopkg.in/alecthomas/kingpin.v2"
)

// nolint: gochecknoglobals,lll
var (
	version       = "dev"
	app           = kingpin.New("elasticsearch-reindex-exporter", "exports elasticsearch reindex task metrics in prometheus format")
	listenAddress = app.Flag("web.listen-address", "Address to listen on for web interface and telemetry").Default(":9421").String()
	metricsPath   = app.Flag("web.telemetry-path", "Path under which to expose metrics").Default("/metrics").String()
	esURL         = app.Flag("es.url", "ElasticSearch URL to scrape").Default("http://localhost:9200").String()
)

func main() {
	app.Version(version)
	app.HelpFlag.Short('h')
	log.AddFlags(app)
	kingpin.MustParse(app.Parse(os.Args[1:]))

	log.Infof("starting elasticsearch-reindex-exporter %s on %s...\n", version, *esURL)

	var client = client.New(*esURL)

	prometheus.MustRegister(collector.NewCollector(client))
	http.Handle(*metricsPath, promhttp.Handler())

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w,
			`
			<html>
			<head><title>ElasticSearch Reindex Exporter</title></head>
			<body>
				<h1>ElasticSearch Reindex Exporter</h1>
				<p><a href="`+*metricsPath+`">Metrics</a></p>
			</body>
			</html>
			`)
	})

	log.Infof("server listening on %s", *listenAddress)
	if err := http.ListenAndServe(*listenAddress, nil); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
