package checkgoself

import (
	"encoding/json"
	"fmt"
	"log"
	"log/syslog"
	"net/http"
	"net/smtp"
	"net/url"
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

type EmailSettings struct {
	SmtpHost     string
	SmtpUsername string
	SmtpPassword string
	SmtpPort     string
	FromEmail    string
}

type Config struct {
	Metrics []Metric
	Email   EmailSettings
}

// Set by flags in checkgoself/main.go
var ConfigFile *string
var SendEmails *bool

//ParseConfig parses config.json, which can be set with a flag
//eg.$ checkgoself -config="../config.json"
func ParseConfig() Config {
	file, _ := os.Open(*ConfigFile)
	decoder := json.NewDecoder(file)
	config := Config{}
	err := decoder.Decode(&config)
	if err != nil {
		log.Fatal("Unable to parse config.json: ", err)
	}
	return config
}

//Check loops over the metrics we're collecting, and does the appropriate check
func Check(metrics []Metric) {

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
		msg = fmt.Sprintf("ALARM: Disk usage %d%% > %d%% ", percentUsed, percentMax)
		alarm(m, msg)
	}
	if msg != "" {
		writeLog(msg)
	}
}

//command executes m.Target using the shell and checks the output == m.Output
func command(m Metric) {
	out := shellCmd(m.Target)
	if out != m.Output {
		msg := fmt.Sprintf("ALARM: %s - %s", m.Name, out)
		writeLog(msg)
		alarm(m, msg)
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
		msg = fmt.Sprintf("ALARM: Request time %s > %s", end, maxDuration)
		alarm(m, msg)
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

//checkError - all errors are fatal for simplicity
func checkError(e error) {
	if e != nil {
		log.Fatal(e)
	}
}

//alarm triggers a HTTP GET or Email when a metric is out of bounds
func alarm(m Metric, msg string) {

	sendEmail(m, msg)

	//Peform HTTP GET
	if m.AlarmGet != "" {

		//Pass some basic parameters along with the GET
		v := url.Values{}
		v.Set("metric", m.Name)
		v.Set("message", msg)

		alarmUrl := m.AlarmGet + "?" + v.Encode()

		_, getErr := http.Get(alarmUrl)
		checkError(getErr)

	}

}

//writeLog to system log
func writeLog(s string) {
	if s == "" {
		return
	}
	l, err := syslog.New(syslog.LOG_ERR, "[checkgoself]")
	defer l.Close()
	if err != nil {
		log.Fatal("error writing syslog!")
	}
	err = l.Warning(s)
	checkError(err)
}

//sendEmail more or less taken from http://golang.org/pkg/net/smtp/#example_PlainAuth
func sendEmail(m Metric, msg string) {

	//if the -email="false" flag is set
	if *SendEmails == false {
		log.Println("NOT SENDING EMAILS")
		return
	}

	//if the alarm email is not specified for this metric
	if m.AlarmEmail == "" {
		return
	}

	config := ParseConfig()
	// Set up authentication information.
	c := config.Email

	auth := smtp.PlainAuth("", c.SmtpUsername, c.SmtpPassword, c.SmtpHost)

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := []string{m.AlarmEmail}

	emailBody := "Subject: [checkgoself] " + msg + "\r\n"

	msgBytes := []byte(emailBody)
	err := smtp.SendMail(c.SmtpHost+":"+c.SmtpPort, auth, c.FromEmail, to, msgBytes)
	if err != nil {
		log.Fatal(err)
	}
}
