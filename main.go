package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"strconv"
	"strings"
	"time"
)

var usage = `
Take an email message via STDIN and forward it to a Matrix room

USAGE:

  sendmail-to-matrix [OPTIONS...]

CONFIGURATION:

  You must define a server, token, and room either using a config file or via command-line parameters.

  Config file format (JSON):
  {
    "server": "https://matrix.example.org",
    "token": "<access token>"
    "room": "!roomid:homeservername"
    "preface": "Preface to message"
  }

OPTIONS:

`

var printUsage = func() {
	fmt.Fprint(flag.CommandLine.Output(), usage)
	flag.PrintDefaults()
	fmt.Fprint(flag.CommandLine.Output(), "\n")
}

type MatrixRequestBody struct {
	Body    string `json:"body"`
	Msgtype string `json:"msgtype"`
}

var config map[string]string

func parseFlags() {
	// Parse the flags
	configFile := flag.String("config-file", "", "Path to config file")
	server := flag.String("server", "", "Matrix homeserver url")
	token := flag.String("token", "", "Matrix account access token")
	room := flag.String("room", "", "Matrix Room ID")
	preface := flag.String("preface", "", "Preface the matrix message with arbitrary text (optional)")
	flag.Parse()

	// Put them in the config
	config["configFile"] = *configFile
	config["server"] = *server
	config["token"] = *token
	config["room"] = *room
	config["preface"] = *preface
}

func getConfig() {
	config = make(map[string]string)

	// Parse command line arguments. These override config file values.
	// We need it parsed before for the --config-file flag
	parseFlags()

	// Parse config file
	parseConfigFile()
}

func parseConfigFile() {
	var confFile map[string]string

	// No config file set
	if config["configFile"] == "" {
		return
	}

	// Read config file
	content, err := ioutil.ReadFile(config["configFile"])
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now unmarshall the data into `payload`
	err = json.Unmarshal(content, &confFile)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// Set the values only if they weren't already set by flags
	required := []string{"server", "token", "room", "preface"}
	for _, key := range required {
		if config[key] == "" {
			config[key] = confFile[key]
		}
	}
}

func validateConfigOrDie() {
	var missing []string
	required := []string{"server", "token", "room"}
	for _, key := range required {
		if config[key] == "" {
			missing = append(missing, key)
		}
	}
	if len(missing) > 0 {
		log.Fatal("Missing required parameters: ", strings.Join(missing, ", "))
	}
}

func buildPreface() (p string) {
	p = config["preface"]
	if p != "" {
		p = p + "\n"
	}
	return
}

func buildSubject(m *mail.Message) (s string) {
	s = m.Header.Get("Subject")
	if s != "" {
		s = "Subject: " + s + "\n"
	}
	return
}

func buildBody(m *mail.Message) string {
	b, err := io.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	return string(b[:])
}

func buildMessage(email string) (message string) {
	r := strings.NewReader(email)
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}
	message += buildPreface()
	message += buildSubject(m)
	message += buildBody(m)
	return
}

func getEmailFromStdin() (email string) {
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		email += scanner.Text() + "\n"
	}
	return
}

func getTransactionId() string {
	// Get a 10 character random string to use as a transaction ID (nonce)
	rand.Seed(time.Now().Unix())
	n := math.Floor(rand.Float64() * math.Pow(10, 10))
	s := strconv.FormatFloat(n, 'f', 0, 64)
	return s
}

func sendMessage(message string) {
	reqBody, err := json.Marshal(MatrixRequestBody{
		Body:    message,
		Msgtype: "m.text",
	})
	if err != nil {
		log.Fatal(err)
	}
	urlFmt := "%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s"
	url := fmt.Sprintf(urlFmt, config["server"], config["room"], getTransactionId())
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal(err)
	}
	query := req.URL.Query()
	query.Set("access_token", config["token"])
	req.URL.RawQuery = query.Encode()
	sendHttpRequest(req)
}

func sendHttpRequest(req *http.Request) {
	client := &http.Client{}
	_, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	// Custom usage message because flag kind of sucks
	flag.Usage = printUsage

	// Read config file if present, otherwise use flags
	getConfig()

	// Validate that required parameters are present
	validateConfigOrDie()

	email := getEmailFromStdin()

	message := buildMessage(email)

	sendMessage(message)
}
