package main

import (
	"log"
	"net/http"
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"helm.sh/helm/v3/pkg/action"
	"helm.sh/helm/v3/pkg/cli"
)

type HelmCollector struct {
	releaseStatus *prometheus.Desc
}

func NewHelmCollector() *HelmCollector {
	return &HelmCollector{
		releaseStatus: prometheus.NewDesc(
			"helm_release_status",
			"Status of Helm releases (1 for deployed, 0 otherwise)",
			[]string{"release_name", "namespace", "chart", "app_version", "chart_version"},
			nil,
		),
	}
}

func (collector *HelmCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.releaseStatus
}

func (collector *HelmCollector) Collect(ch chan<- prometheus.Metric) {
	settings := cli.New()
	actionConfig := new(action.Configuration)
	if err := actionConfig.Init(settings.RESTClientGetter(), settings.Namespace(), os.Getenv("HELM_DRIVER"), log.Printf); err != nil {
		log.Fatalf("Error initializing Helm action configuration: %v", err)
	}

	client := action.NewList(actionConfig)
	client.AllNamespaces = true
	releases, err := client.Run()
	if err != nil {
		log.Fatalf("Error fetching Helm releases: %v", err)
	}

	for _, release := range releases {
		status := float64(0)
		if release.Info.Status == "deployed" {
			status = 1
		}

		ch <- prometheus.MustNewConstMetric(
			collector.releaseStatus,
			prometheus.GaugeValue,
			status,
			release.Name,
			release.Namespace,
			release.Chart.Metadata.Name,
			release.Chart.Metadata.AppVersion,
			release.Chart.Metadata.Version,
		)
	}
}

func main() {
	collector := NewHelmCollector()
	prometheus.MustRegister(collector)

	http.Handle("/metrics", promhttp.Handler())
	log.Println("Starting server on :2112")
	log.Fatal(http.ListenAndServe(":2112", nil))
}
