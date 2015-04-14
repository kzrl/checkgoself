package main

import (
	"encoding/json"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metric struct {
	Name       string
	Type       string
	Target     string
	MaxValue   string
	Output     string
	AlarmEmail string
	AlarmGet   string
}

type Config struct {
	Metrics []Metric
}

func main() {
	config := parseConfig()

	check(config.Metrics)
}

func parseConfig() Config {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to parse config.json: ", err)
	}
	return config
}

func check(metrics []Metric) {

	//Loop over the metrics we're collecting. Check them
	for _, metric := range metrics {

		switch metric.Type {
		case "command":
			command(metric)

		case "httpreq":
			httpreq(metric)

		case "freespace":
			diskUsage(metric)
		}

	}
}

//diskUsage checks the current usage is less than m.MaxValue
func diskUsage(m Metric) {

	cmd := fmt.Sprintf("df -h %s | awk 'NR>1{print $5}'", m.Target)

	f := shellCmd(cmd)
	f = strings.TrimSuffix(f, "%")
	percentUsed, err := strconv.Atoi(f)
	checkError(err)
	max := strings.TrimSuffix(m.MaxValue, "%")
	percentMax, err := strconv.Atoi(max)
	checkError(err)

	msg := ""

	if percentUsed > percentMax {
		msg = fmt.Sprintf("ALERT: Disk usage %d%% > %d%% ", percentUsed, percentMax)
	} else {
		msg = fmt.Sprintf("OK: Disk usage %d%% < %d%%", percentUsed, percentMax)
	}
	if msg != "" {
		writeLog(msg)
	}
}

//command executes m.Target using the shell and checks the output == m.Output
func command(m Metric) {
	out := shellCmd(m.Target)
	if out != m.Output {
		//@TODO log this to http://golang.org/pkg/log/syslog/
		writeLog(fmt.Sprintf("ALERT: %s - %s", m.Name, out))
	}
}

//httpreq makes a GET request on m.Target and checks the response time < m.MaxValue
func httpreq(m Metric) {

	start := time.Now()

	_, err := http.Get(m.Target)
	checkError(err)

	end := time.Since(start)

	maxDuration, err := time.ParseDuration(m.MaxValue)
	checkError(err)
	msg := ""
	if end > maxDuration {
		msg = fmt.Sprintf("ALERT: Request time %s > %s", end, maxDuration)
	} else {
		msg = fmt.Sprintf("OK: Request time %s < %s", end, maxDuration)
	}
	writeLog(msg)
}

//shellCmd runs a command using the shell
func shellCmd(target string) string {

	cmd := "/bin/sh"
	output, err := exec.Command(cmd, "-c", target).Output()
	if err != nil {
		log.Fatal(err)
	}

	s := string(output[:])
	s = strings.TrimSuffix(s, "\n")

	return s
}

//checkError
func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

func alarm(m Metric) {

}

//writeLog
func writeLog(s string) {
	log.Println(s)
	l, err := syslog.New(syslog.LOG_ERR, "[checkgoself]")
	defer l.Close()
	if err != nil {
		log.Fatal("error writing syslog!")
	}
	err = l.Warning(s)
	checkError(err)
}
