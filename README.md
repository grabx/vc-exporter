# Visual Cron Exporter (vc-exporter)

Visual Cron Exporter is a [Prometheus Exporter](https://prometheus.io/docs/instrumenting/exporters/) that collects job metrics from VisualCron using the [Rest API Client](https://github.com/grabx/vcclient)

The goal is to make the Visual Cron API readable for [Prometheus](https://prometheus.io/). 

## Metrics

The exporter collects the following metrics:

```sh
# Displays whether the job is active or not. 1 = Actice, 0 = Inactive
job_active
```
```sh
# Displays the job's last exit code. Can be any value since it is customizable on a per script basis.
job_exit_code
```
```sh
# Displays the job's last exit code result. 1 = Success, 2 = Fail/Unknown/Currently Running, 3 = Never Ran.
job_exit_code_result
```
```sh
# Displays if the job has missed 1 = Missed, 0 = Not Missed
job_missed
```
```sh
# Displays the jobs execution time
job_execution_time
```
```sh
# Displays the jobs status. 1 = Waiting, 0 = Running
job_status
```

For each job a metric is generated. Each metric has the labels ```jobName``` and ```jobId```. With these labels it is easier to see which job created the metric and it also helps you to create rules by correlating the jobIds and/or names between metrics.

## Basic Prometheus Configuration

Add the following block to the ```scrape_configs``` of your prometheus.yml config file:
```yaml
scrape_configs:
  - job_name: vc_exporter
    static_configs:
    - targets: ['<<VISUAL-CRON-SERVER-IP-OR-HOSTNAME>>:8008']
```

## Configuration of Visual Cron Exporter

The Visual Cron Exporter can be configured directly in the source code. Things such as the listening port and username and password need to be set manually before compiling. By default the listening port is **8008**. 
To configure the username and password, set the parameters of ```vcclient.NewClient()``` manually inside the code. Refer to the documentation [here](https://pkg.go.dev/github.com/grabx/vcclient#NewClient) if you are unsure how to set it.

### Compiling

You need to have Go 1.19 intalled on your system. To build a working agent you need to either build it on a Windows host by using ```go build``` inside the working directory.

 If you are building the agent on any other operating system use ```GOOS=windows go build``` to build a Windows binary. In the end the build should result in a ```vc-exporter.exe``` file.

### Configuring a Windows Service

To create a Windows Service that will run the agent in the background please move the file to a location like ```C:\Program Files``` or any other directory where you wish to store it safely. *The following example uses ```C:\vcmonitor\vc-exporter.exe``` as the binary path.*

Next, create a Windows service by running the following commands:

```powershell
New-Service -Name VCMonitor -BinaryPathName C:\vcmonitor\vc-exporter.exe -StartupType Automatic -DisplayName "VisualCron Monitor"

Start-Service VCMonitor
```