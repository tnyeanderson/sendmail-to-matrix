package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/mail"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/microcosm-cc/bluemonday"
)

var Version = "v0.0.1"

var usage = `
Read an email message from STDIN and forward it to a Matrix room

USAGE:

  sendmail-to-matrix [OPTIONS...]

CONFIGURATION:

  A server, token, and room must be set either using a config file or via command-line parameters.

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

// I know, "this should be a struct!"
// But flags need to be parsed first, applying conf file values if not set by flags
// Check the parseConfigFile and validateConfigOrDie functions first
// Then try to convince me to make it a struct for no benefit :)
type Config map[string]string

var config Config

func parseFlags() {
	// Parse the flags
	version := flag.Bool("version", false, "Print the application version and exit")
	configFile := flag.String("config-file", "", "Path to config file")
	server := flag.String("server", "", "Matrix homeserver url")
	token := flag.String("token", "", "Matrix account access token")
	room := flag.String("room", "", "Matrix Room ID")
	preface := flag.String("preface", "", "Preface the matrix message with arbitrary text (optional)")
	flag.Parse()

	// If the version flag is set, print the version and exit
	if *version {
		fmt.Println(Version)
		os.Exit(0)
	}

	// Put them in the config
	config["configFile"] = *configFile
	config["server"] = *server
	config["token"] = *token
	config["room"] = *room
	config["preface"] = *preface
}

func getConfig() {
	config = make(Config)

	// Parse command line arguments. These override config file values.
	// We need it parsed before for the --config-file flag
	parseFlags()

	// Parse config file
	parseConfigFile()
}

func parseConfigFile() {
	var confFile Config

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

func buildMessage(email io.Reader) (message string) {
	m, err := mail.ReadMessage(email)
	if err != nil {
		log.Fatal(err)
	}
	message += buildPreface()
	message += buildSubject(m)
	message += buildBody(m)
	return
}

func getTransactionId() string {
	// Get a 10 character random string to use as a transaction ID (nonce)
	rand.Seed(time.Now().Unix())
	n := math.Floor(rand.Float64() * math.Pow(10, 10))
	s := strconv.FormatFloat(n, 'f', 0, 64)
	return s
}

func getUrl(server string, room string) string {
	urlFmt := "%s/_matrix/client/v3/rooms/%s/send/m.room.message/%s"
	url := fmt.Sprintf(urlFmt, server, room, getTransactionId())
	return url
}

func addAccessToken(req *http.Request, token string) {
	query := req.URL.Query()
	query.Set("access_token", token)
	req.URL.RawQuery = query.Encode()
}

func removeHtmlTags(message string) (s string) {
	policy := bluemonday.StrictPolicy()
	s = policy.Sanitize(message)
	s = html.UnescapeString(s)
	s = fixWhitespace(s)
	return
}

func fixWhitespace(message string) string {
	re := regexp.MustCompile("\n\n+")
	return re.ReplaceAllLiteralString(message, "\n")
}

func getRequestBody(message string) *bytes.Buffer {
	b := removeHtmlTags(message)
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(MatrixRequestBody{
		Body:    b,
		Msgtype: "m.text",
	})
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

func sendMessage(message string) {
	url := getUrl(config["server"], config["room"])
	body := getRequestBody(message)
	req, err := http.NewRequest(http.MethodPut, url, body)
	if err != nil {
		log.Fatal(err)
	}
	addAccessToken(req, config["token"])
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
	flag.Usage = printUsage
	getConfig()
	validateConfigOrDie()
	message := buildMessage(os.Stdin)
	sendMessage(message)
}
