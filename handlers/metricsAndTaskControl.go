package handlers

import (
	"encoding/json"
	"mertics-exporter/domain"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
)

type Handlers struct {
	Storage      domain.Storage
	BoConnection bool
}

var (
	// CurrentTasks = promauto.NewGauge(
	// 	prometheus.GaugeOpts{
	// 		Name: "task_in_process",
	// 		Help: "Current number of tasks in process",
	// 	})

	// CurrentArticles = promauto.NewGauge(
	// 	prometheus.GaugeOpts{
	// 		Name: "articles_in_process",
	// 		Help: "Current number of pending notification",
	// 	})

	// MatchingTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_matching_total",
	// 		Help: "Total number of matching tasks",
	// 	})

	// PreloadTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_preload_total",
	// 		Help: "Total number of slow updates",
	// 	})

	// ImageTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_image_total",
	// 		Help: "Total number of fast updates",
	// 	})

	// ReplacedTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_replaced",
	// 		Help: "Total number of replaced tasks",
	// 	})

	// FailedTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_failed",
	// 		Help: "Total number of failed tasks",
	// 	})

	// SucceededTasks = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "task_succeeded",
	// 		Help: "Total number of failed tasks",
	// 	})

	// ProcessCounter = promauto.NewCounter(
	// 	prometheus.CounterOpts{
	// 		Name: "process_counter",
	// 		Help: "Total number of failed tasks",
	// 	})

	// TaskDurationHistogram = prometheus.NewHistogramVec(
	// 	prometheus.HistogramOpts{
	// 		Name:    "task_duration_seconds",
	// 		Help:    "Tasks duration distribution",
	// 		Buckets: []float64{15, 30, 60, 90, 120, 240, 480},
	// 	},
	// 	[]string{"job_type"},
	// )

	EslLabelsTotalMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_labels_total",
		})

	EslLabelsOfflineMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_labels_offline",
		})

	EslUpdatesFailedMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_updates_failed",
		})

	EslUpdatesOnlineFailedMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_updates_failed_online",
		})

	EslUpdatesGoodFailedMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_updates_failed_good",
		})

	EslUpdatesWaitingMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_updates_waiting",
		})

	EslLabelsBadBatteryMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_labels_bad_battery",
		})

	EslApTotalMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_ap_total",
		})

	EslApOnlineMetric = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "esl_ap_online",
		})
)

func (h *Handlers) GetTasksForPlu(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := strings.Split(parsedUrl.Path, "/")

	pattern := regexp.MustCompile(`^\d{3,8}`)
	if len(input) == 5 && pattern.MatchString(input[4]) {
		result := h.Storage.GetTasksForPLU(input[4])
		data, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) GetTasksForLabel(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := strings.Split(parsedUrl.Path, "/")

	labelIdPattern := regexp.MustCompile(`\D[\d\D]{7}`)

	if len(input) == 5 && labelIdPattern.MatchString(input[4]) {
		result := h.Storage.GetTasksForLabel(input[4])
		data, _ := json.Marshal(result)
		w.WriteHeader(http.StatusOK)
		w.Write(data)

	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) GetTasksForID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	parsedUrl, err := url.Parse(r.URL.RequestURI())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	input := strings.Split(parsedUrl.Path, "/")

	if len(input) == 5 {
		taskId, err := uuid.Parse(input[4])
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if task, ok := h.Storage.GetTaskById(taskId); ok {
			data, _ := json.Marshal(task)
			w.WriteHeader(http.StatusOK)
			w.Write(data)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	} else {
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (h *Handlers) GetAllTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	data, _ := json.Marshal(h.Storage.GetAllTasks())
	w.WriteHeader(http.StatusOK)
	w.Write(data)

}

func (h *Handlers) StopProcess(w http.ResponseWriter, _ *http.Request) {
	h.Storage.ClearProcess()
	h.Storage.DumpAllTasks()
	w.WriteHeader(http.StatusOK)
	defer os.Exit(1)
}
