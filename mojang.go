package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/charmbracelet/log"
)

type MojangAccount struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetUserMojang(username string) MojangAccount {
	var account MojangAccount
	resp, err := http.Get(fmt.Sprintf("https://api.mojang.com/users/profiles/minecraft/%s", username))
	log.Debug("Mojang API resp: %v", resp)
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
	log.Debugf("Mojang account: %v", account)
	return account
}
