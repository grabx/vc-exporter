package main

import (
	"context"
	"log"
	"net/http"
	"time"

	vcclient "github.com/grabx/vcclient"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

/*
Struct contains the mandatory descriptors
*/
type jobCollector struct {
	jobActiveMetric   *prometheus.Desc
	jobExitCode       *prometheus.Desc
	jobExitCodeResult *prometheus.Desc
	jobMissed         *prometheus.Desc
	jobExecutionTime  *prometheus.Desc
	jobStatus         *prometheus.Desc
}

// You must create a constructor for you collector that
// initializes every descriptor and returns a pointer to the collector
func newjobCollector() *jobCollector {
	return &jobCollector{
		jobActiveMetric: prometheus.NewDesc(
			"job_active",
			"Displays whether the job is active or not. 1 = Actice, 2 = Inactive",
			[]string{"jobName", "jobId"}, nil),
		jobExitCode: prometheus.NewDesc(
			"job_exit_code",
			"Displays the job's last exit code. Can be any value since it is customizable on a per script basis.",
			[]string{"jobName", "jobId"}, nil),
		jobExitCodeResult: prometheus.NewDesc(
			"job_exit_code_result",
			"Displays the job's last exit code result. 1 = Success, 2 = Fail/Unknown/Currently Running, 3 = Never Ran.",
			[]string{"jobName", "jobId"}, nil),
		jobMissed: prometheus.NewDesc(
			"job_missed",
			"Displays if the job has missed",
			[]string{"jobName", "jobId"}, nil),
		jobExecutionTime: prometheus.NewDesc(
			"job_execution_time",
			"Displays the jobs execution time",
			[]string{"jobName", "jobId"}, nil),
		jobStatus: prometheus.NewDesc(
			"job_status",
			"Displays the jobs statu. 1 = Waiting, 0 = Running",
			[]string{"jobName", "jobId"}, nil),
	}

}

// Each and every collector must implement the Describe function.
// It essentially writes all descriptors to the prometheus desc channel.
func (collector *jobCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- collector.jobActiveMetric
	ch <- collector.jobExitCode
	ch <- collector.jobMissed
	ch <- collector.jobExitCodeResult
	ch <- collector.jobExecutionTime
	ch <- collector.jobStatus
}

func (collector *jobCollector) Collect(ch chan<- prometheus.Metric) {
	jobs := getJobData()
	for _, job := range *jobs {
		var active float64
		if job.Stats.Active {
			active = 1
		} else {
			active = 0
		}
		var missed float64
		if job.Missed {
			missed = 1
		} else {
			missed = 0
		}
		var lastExecution time.Time
		if lastExec, err := time.Parse(time.RFC3339, job.Stats.DateLastExecution); err != nil {
			lastExecution = time.Time{}
		} else {
			lastExecution = lastExec
		}
		var lastMissed time.Time
		if lastMiss, err := time.Parse(time.RFC3339, job.MissedDate); err != nil {
			lastMissed = time.Time{}
		} else {
			lastMissed = lastMiss
		}
		ch <- prometheus.MustNewConstMetric(
			collector.jobActiveMetric,
			prometheus.GaugeValue,
			active, job.Name, job.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.jobStatus,
			prometheus.GaugeValue,
			float64(job.Stats.Status), job.Name, job.ID,
		)
		ch <- prometheus.NewMetricWithTimestamp(lastExecution, prometheus.MustNewConstMetric(
			collector.jobExitCode,
			prometheus.GaugeValue,
			float64(job.Stats.ExitCode), job.Name, job.ID,
		))
		ch <- prometheus.NewMetricWithTimestamp(lastExecution, prometheus.MustNewConstMetric(
			collector.jobExitCodeResult,
			prometheus.GaugeValue,
			float64(job.Stats.ExitCodeResult), job.Name, job.ID,
		))
		ch <- prometheus.NewMetricWithTimestamp(lastMissed, prometheus.MustNewConstMetric(
			collector.jobMissed,
			prometheus.GaugeValue,
			missed, job.Name, job.ID,
		))
		ch <- prometheus.NewMetricWithTimestamp(lastExecution, prometheus.MustNewConstMetric(
			collector.jobExecutionTime,
			prometheus.GaugeValue,
			job.Stats.ExecutionTime, job.Name, job.ID,
		))
	}
}

func getJobData() *[]vcclient.Jobs {
	ctx := context.TODO()
	c := vcclient.NewClient(
		"",
		"",
	)
	jobs, err := c.GetJobs(ctx)
	if err != nil {
		log.Fatalln(err)
	}
	return jobs
}

func init() {
	collector := newjobCollector()
	prometheus.MustRegister(collector)
}

func main() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8008", nil))
}
