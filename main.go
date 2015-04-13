package main

import (
	"encoding/json"
	"log"
	"os"
	"os/exec"
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
	for i, metric := range metrics {
		log.Printf("%v => %+v", i, metric.Name)

		switch metric.Type {
		case "command":
			log.Println(metric.Target)
			cmd := "/bin/sh"
			output, err := exec.Command(cmd, metric.Target).Output()
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Output is %v\n", output)

		case "httpreq":
			log.Println("Do HTTP Things")

		}

	}
}

func diskusage() {

}
