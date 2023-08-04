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
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/storage/sqlite3"
)

type ExarotonBody struct {
	Success bool            `json:"success"`
	Error   string          `json:"error"`
	Data    json.RawMessage `json:"data"`
}

type ExarotonServer struct {
	Id       string                 `json:"id"`
	Name     string                 `json:"name"`
	Address  string                 `json:"address"`
	Motd     string                 `json:"motd"`
	Status   int                    `json:"status"`
	Host     string                 `json:"host"`
	Port     int                    `json:"port"`
	Software ExarotonServerSoftware `json:"software"`
}

type ExarotonServerSoftware struct {
	Id      string `json:"id"`
	Name    string `json:"name"`
	Version string `json:"version"`
}

func exarotonRequestV1(verb string, path string, bodyReq io.Reader) (response []byte, err error) {
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
	return bodyResp, err
}

func exarotonManagePlayersList(verb string, playerName string) (response []byte, err error) {
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
		resp, err := exarotonRequestV1(verb, fmt.Sprintf("/servers/%s/playerlists/%s/", serverId, "whitelist"), bytes.NewBuffer(out))
		if err != nil {
			log.Debug(out)
			return resp, err
		}
	}
	return nil, nil
}

func exarotonGetServersList(cache *sqlite3.Storage) (response []ExarotonServer, err error) {
	var respJson []byte
	var resp ExarotonBody
	var servers []ExarotonServer

	// Get exaroton JSON servers list from cache
	respJson, err = cache.Get("exaroton-servers-list-json")
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// If cache is empty, get the JSON list and save it
	if respJson == nil {
		log.Debug("exaroton-servers-list-json cache is empty.")
		respJson, err = exarotonRequestV1("GET", fmt.Sprintf("/servers/"), &bytes.Buffer{})
		if err != nil {
			log.Error(err)
			return nil, err
		}
		cache.Set("exaroton-servers-list-json", respJson, 15*time.Minute)
		log.Debug("exaroton-servers-list-json cache saved.")
	}
	err = json.Unmarshal(respJson, &resp)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	err = json.Unmarshal(resp.Data, &servers)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Extract only IDs we want
	serverIDs := strings.Split(os.Getenv(EXAROTON_SERVERS_ID), ",")
	filteredServers := make([]ExarotonServer, 0)
	for _, server := range servers {
		for _, targetID := range serverIDs {
			if server.Id == targetID {
				filteredServers = append(filteredServers, server)
			}
		}
	}
	return filteredServers, nil
}

func exarotonAllowUser(playerName string) (response []byte, err error) {
	return exarotonManagePlayersList("PUT", playerName)
}

func exarotonRemoveUser(playerName string) (response []byte, err error) {
	return exarotonManagePlayersList("DELETE", playerName)
}
