package main

import (
	"context"
	"log"
	"net/http"

	vcclient "github.com/grabx/vcclient"
	"github.com/kardianos/service"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	logger service.Logger
)

// Initialize program to run in background
type program struct{}

// Handle windows service start
func (p *program) Start(s service.Service) error {
	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

// Handle program logic in background
func (p *program) run() {
	http.Handle("/metrics", promhttp.Handler())
	log.Fatal(http.ListenAndServe(":8008", nil))
}

// Handle service stop request
func (p *program) Stop(s service.Service) error {
	log.Println("Stopping agent.")
	// Stop should not block. Return with a few seconds.
	return nil
}

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
			"Displays whether the job is active or not. 1 = Actice, 0 = Inactive",
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
			"Displays the jobs status. 1 = Waiting, 0 = Running",
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

// Collect metrics and send them to the metrics channel of our metrics
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
		ch <- prometheus.MustNewConstMetric(
			collector.jobExitCode,
			prometheus.GaugeValue,
			float64(job.Stats.ExitCode), job.Name, job.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.jobExitCodeResult,
			prometheus.GaugeValue,
			float64(job.Stats.ExitCodeResult), job.Name, job.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.jobMissed,
			prometheus.GaugeValue,
			missed, job.Name, job.ID,
		)
		ch <- prometheus.MustNewConstMetric(
			collector.jobExecutionTime,
			prometheus.GaugeValue,
			job.Stats.ExecutionTime, job.Name, job.ID,
		)
	}
}

// Request job data from visual cron api client
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

// Run interactively or handle being run as windows service
func main() {
	svcConfig := &service.Config{
		Name:        "VCMonitor",
		DisplayName: "VisualCron Monitor",
		Description: "Monitors job execution and states via the REST API",
	}

	prg := &program{}
	s, err := service.New(prg, svcConfig)
	if err != nil {
		log.Fatal(err)
	}
	logger, err = s.Logger(nil)
	if err != nil {
		log.Fatal(err)
	}
	err = s.Run()
	if err != nil {
		logger.Error(err)
	}
}
