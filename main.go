package main

import (
	"fmt"
	"mertics-exporter/config"
	"mertics-exporter/domain"
	"mertics-exporter/handlers"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Handlers struct {
	Storage      domain.Storage
	BoConnection bool
}

func main() {

	config.ParseOptions()
	fmt.Println(config.Options.Login, config.Options.Password)
	Handlers := handlers.Handlers{}
	prometheus.MustRegister(handlers.EslApOnlineMetric)
	prometheus.MustRegister(handlers.EslApTotalMetric)
	prometheus.MustRegister(handlers.EslLabelsBadBatteryMetric)
	prometheus.MustRegister(handlers.EslLabelsOfflineMetric)
	prometheus.MustRegister(handlers.EslLabelsTotalMetric)
	prometheus.MustRegister(handlers.EslUpdatesFailedMetric)
	prometheus.MustRegister(handlers.EslUpdatesGoodFailedMetric)
	prometheus.MustRegister(handlers.EslUpdatesWaitingMetric)
	go func() {
		totalLabels, apTotal, waitingTasks := Handlers.Storage.ESLConnector.EslGetServiceStatus()
		offlineLabels, failedUpdates, failedOnlineLabels, failedGoodSignalLabels, badBatteryLabels := Handlers.Storage.ESLConnector.EslGetLabelsStatus()
		apOnline := Handlers.Storage.ESLConnector.EslGetApStatus(apTotal)
		handlers.EslLabelsTotalMetric.Set(float64(totalLabels))
		handlers.EslApTotalMetric.Set(float64(apTotal))
		handlers.EslApOnlineMetric.Set(float64(apOnline))
		handlers.EslLabelsOfflineMetric.Set(float64(offlineLabels))
		handlers.EslUpdatesFailedMetric.Set(float64(failedUpdates))
		handlers.EslUpdatesWaitingMetric.Set(float64(waitingTasks))
		handlers.EslLabelsBadBatteryMetric.Set(float64(badBatteryLabels))
		handlers.EslUpdatesGoodFailedMetric.Set(float64(failedGoodSignalLabels))
		handlers.EslUpdatesOnlineFailedMetric.Set(float64(failedOnlineLabels))
		time.Sleep(30 * time.Second)
	}()

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	fmt.Println("Metrics location: http://localhost:9116/metrics")
	http.ListenAndServe(":9116", mux)
}
