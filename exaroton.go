package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"

	"github.com/charmbracelet/log"
)

func exarotonRequestV1(verb string, path string, bodyReq io.Reader) error {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	url := fmt.Sprintf("https://api.exaroton.com/v1%s", path)

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + os.Getenv(EXAROTON_API_KEY)

	// Create a new request using http
	req, err := http.NewRequest(verb, url, bodyReq)
	if err != nil {
		panic(err)
	}

	// add authorization header to the req
	req.Header.Add("Authorization", bearer)
	req.Header.Add("Content-Type", "application/json")

	// Send req using http Client
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatalf("Error on response %s", err)
	}
	defer resp.Body.Close()

	bodyResp, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error while reading the response bytes: %s", err)
	}
	log.Debug(string([]byte(bodyResp)))
	return err
}

func exarotonManagePlayersList(verb string, playerName string) error {
	serversId := strings.Split(os.Getenv(EXAROTON_SERVERS_ID), ",")

	body := map[string]interface{}{
		"entries": []string{
			playerName,
		},
	}

	log.Debug("Body struct", "body", body)

	out, err := json.Marshal(body)
	log.Debug("JSON Marshal out", "out", bytes.NewBuffer(out))
	log.Debug("Json Marshal out", "string", string(out))
	if err != nil {
		log.Fatal(err)
	}

	for _, serverId := range serversId {
		err := exarotonRequestV1(verb, fmt.Sprintf("/servers/%s/playerlists/%s/", serverId, "whitelist"), bytes.NewBuffer(out))
		if err != nil {
			log.Debug(out)
			return err
		}
	}
	return nil
}

func exarotonAllowUser(playerName string) error {
	return exarotonManagePlayersList("PUT", playerName)
}

func exarotonRemoveUser(playerName string) error {
	return exarotonManagePlayersList("DELETE", playerName)
}
