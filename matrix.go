package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

type MatrixRequestBody struct {
	Body    string `json:"body"`
	Msgtype string `json:"msgtype"`
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

func buildRequestBody(message string) *bytes.Buffer {
	buf := bytes.NewBuffer([]byte{})
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	err := enc.Encode(MatrixRequestBody{
		Body:    message,
		Msgtype: "m.text",
	})
	if err != nil {
		log.Fatal(err)
	}
	return buf
}

func sendMessage(message string) {
	url := getUrl(config["server"], config["room"])
	body := buildRequestBody(message)
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
