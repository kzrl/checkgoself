package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

type Metric struct {
	Name     string
	Type     string
	Target   string
	MaxValue string
	Output   string
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
	f := shellCmd(m.Target)
	f = strings.TrimSuffix(f, "%")
	percentUsed, err := strconv.Atoi(f)
	checkError(err)
	max := strings.TrimSuffix(m.MaxValue, "%")
	percentMax, err := strconv.Atoi(max)
	checkError(err)

	//@TODO log this to http://golang.org/pkg/log/syslog/
	if percentUsed > percentMax {
		log.Printf("ALERT: Disk usage %d%% > %d%%", percentUsed, percentMax)
	} else {
		log.Printf("OK: Disk usage %d%% < %d%%", percentUsed, percentMax)
	}
}

//command executes m.Target using the shell and checks the output == m.Output
func command(m Metric) {
	out := shellCmd(m.Target)
	if out != m.Output {
		//@TODO log this to http://golang.org/pkg/log/syslog/
		log.Printf("ALERT: %s - %s", m.Name, out)
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

	if end > maxDuration {
		log.Printf("ALERT: Request time %s > %s", end, maxDuration)
	} else {
		log.Printf("OK: Request time %s < %s", end, maxDuration)

	}
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
