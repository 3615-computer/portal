package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

type MojangAccount struct {
	Id string `json:"id"`
}

func GetUserIdMojang(username string) string {
	var account MojangAccount
	resp, err := http.Get(fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username))
	if err != nil {
		log.Fatal(resp)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(resp)
	}

	err = json.Unmarshal(body, &account)
	if err != nil {
		log.Fatal(resp)
	}
	log.Debugf("Mojang ID: %s", account.Id)
	return account.Id
}
