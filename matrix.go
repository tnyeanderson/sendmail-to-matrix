package main

import (
	"bytes"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math"
	"math/big"
	"net/http"
)

type MatrixRequestBody struct {
	Body    string `json:"body"`
	Msgtype string `json:"msgtype"`
}

func getToken(server, user, password string) (string, error) {
	uri := fmt.Sprintf("%s/_matrix/client/r0/login", server)
	bodyfmt := `{"type":"m.login.password", "user": "%s", "password":"%s"}`
	body := fmt.Sprintf(bodyfmt, user, password)
	client := &http.Client{}
	res, err := client.Post(uri, "application/json; charset=UTF-8", bytes.NewBuffer([]byte(body)))
	if err != nil {
		return "", err
	}
	b, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}
	data := make(map[string]interface{})
	err = json.Unmarshal(b, &data)
	if err != nil {
		return "", err
	}
	if token, ok := data["access_token"].(string); ok {
		return token, nil
	}
	return "", fmt.Errorf("Failed to unmarshal access_token from response")
}

func getTransactionId() string {
	// Get a 10 character random string to use as a transaction ID (nonce)
	length := float64(10)
	min := int64(math.Pow(10, length-1))
	max := int64(math.Pow(10, length) - 1)
	r, err := rand.Int(rand.Reader, big.NewInt(max))
	if err != nil {
		log.Fatal(err)
	}
	v := big.NewInt(0).Add(r, big.NewInt(min))
	return v.String()
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
