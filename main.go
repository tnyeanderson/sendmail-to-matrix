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

  sendmail-to-matrix [-h] [-f CONFIG_FILE] [-s SERVER] [-t TOKEN] [-r ROOM] [-p PREFACE]

OPTIONS:

  -h, --help            show this help message and exit

  -f CONFIG_FILE, --config-file CONFIG_FILE
                        Path to config file

  -s SERVER, --server SERVER
                        The matrix homeserver url

  -t TOKEN, --token TOKEN
                        Matrix account access token

  -r ROOM, --room ROOM  The matrix Room ID

  -p PREFACE, --preface PREFACE
                        Preface the matrix message with arbitrary text (optional)

CONFIGURATION:

  You must define a server, token, and room either using a config file or via command-line parameters.

  Config file format (JSON):
  {
    "server": "https://matrix.example.org",
    "token": "<access token>"
    "room": "!roomid:homeservername"
    "preface": "Preface to message"
  }

`

var printUsage = func() {
	fmt.Print(usage)
}

type MatrixRequestBody struct {
	Body    string `json:"body"`
	Msgtype string `json:"msgtype"`
}

type Config struct {
	configFile string
	server     string
	token      string
	room       string
	preface    string
}

var config Config

func parseFlags() {
	// -f, --config-file
	flag.StringVar(&config.configFile, "f", "", "")
	flag.StringVar(&config.configFile, "config-file", "", "")

	// -s, --server
	flag.StringVar(&config.server, "s", "", "")
	flag.StringVar(&config.server, "server", "", "")

	// -t, --token
	flag.StringVar(&config.token, "t", "", "")
	flag.StringVar(&config.token, "token", "", "")

	// -r, --room
	flag.StringVar(&config.room, "r", "", "")
	flag.StringVar(&config.room, "room", "", "")

	// -p, --preface
	flag.StringVar(&config.preface, "p", "", "")
	flag.StringVar(&config.preface, "preface", "", "")

	flag.Parse()
}

func getConfig() {
	// Parse command line arguments. These override config file values.
	// We need it parsed before for the --config-file flag
	parseFlags()

	// Parse config file
	parseConfigFile()
}

func parseConfigFile() {
	var jsonConfig map[string]string

	// No config file set
	if config.configFile == "" {
		return
	}

	// Read config file
	content, err := ioutil.ReadFile(config.configFile)
	if err != nil {
		log.Fatal("Error when opening file: ", err)
	}

	// Now let's unmarshall the data into `payload`
	err = json.Unmarshal(content, &jsonConfig)
	if err != nil {
		log.Fatal("Error during Unmarshal(): ", err)
	}

	// Set the values only if they weren't already set by flags
	if config.server == "" {
		config.server = jsonConfig["server"]
	}
	if config.token == "" {
		config.token = jsonConfig["token"]
	}
	if config.room == "" {
		config.room = jsonConfig["room"]
	}
	if config.preface == "" {
		config.preface = jsonConfig["preface"]
	}
}

func validateConfigOrDie() {
	var missing []string
	if config.server == "" {
		missing = append(missing, "server")
	}
	if config.token == "" {
		missing = append(missing, "token")
	}
	if config.room == "" {
		missing = append(missing, "room")
	}
	if len(missing) > 0 {
		log.Fatal("Missing required parameters: ", strings.Join(missing, ", "))
	}
}

func buildMessage(email string) (message string) {
	r := strings.NewReader(email)
	m, err := mail.ReadMessage(r)
	if err != nil {
		log.Fatal(err)
	}
	preface := config.preface
	if preface != "" {
		message += preface + "\n"
	}
	subject := m.Header.Get("Subject")
	if subject != "" {
		message += "Subject: " + subject + "\n"
	}
	body, err := io.ReadAll(m.Body)
	if err != nil {
		log.Fatal(err)
	}
	message += string(body[:])
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
	txnId := getTransactionId()
	url := fmt.Sprintf("%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s", config.server, config.room, txnId)
	req, err := http.NewRequest(http.MethodPut, url, bytes.NewBuffer(reqBody))
	if err != nil {
		log.Fatal(err)
	}
	query := req.URL.Query()
	query.Set("access_token", config.token)
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
