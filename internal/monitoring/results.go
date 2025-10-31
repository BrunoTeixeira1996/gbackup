package monitoring

import (
	"log"

	"github.com/BrunoTeixeira1996/gbackup/internal/targets"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/push"
)

func SendFinalResultsToMonitoring(results []targets.BackupResult) {
	// Collectors slice
	var collectors []prometheus.Collector

	for _, r := range results {
		sizeBefore := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "backup_size_before_mb",
			Help:        "Size before backup in MB",
			ConstLabels: prometheus.Labels{"target": r.TargetName},
		})
		sizeBefore.Set(r.TargetSize.Before)
		collectors = append(collectors, sizeBefore)

		sizeAfter := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "backup_size_after_mb",
			Help:        "Size after backup in MB",
			ConstLabels: prometheus.Labels{"target": r.TargetName},
		})
		sizeAfter.Set(r.TargetSize.After)
		collectors = append(collectors, sizeAfter)

		elapsed := prometheus.NewGauge(prometheus.GaugeOpts{
			Name:        "backup_elapsed_seconds",
			Help:        "Elapsed backup time in seconds",
			ConstLabels: prometheus.Labels{"target": r.TargetName},
		})
		elapsed.Set(r.ElapsedTime.Value)
		collectors = append(collectors, elapsed)
	}

	// Push all metrics at once
	pusher := push.New("http://192.168.30.24:9091", "gbackup_metrics")
	for _, c := range collectors {
		pusher.Collector(c)
	}

	if err := pusher.Push(); err != nil {
		log.Printf("[monitoring error] could not push metrics: %v", err)
	} else {
		log.Printf("[monitoring info] all metrics pushed successfully")
	}
}
