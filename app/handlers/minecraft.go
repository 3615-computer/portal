package handlers

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"mastodon-to-exaroton-oauth2/app/config"
	"mastodon-to-exaroton-oauth2/app/models"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
)

type MojangAccount struct {
	Id   string `json:"id"`
	Name string `json:"name"`
}

func GetMinecraft(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	params := fiber.Map{}
	servers, _ := exarotonGetServersList(config.Storage.Cache)
	if mastodonAccount.UserID != "" {
		params["mastodonAccount"] = mastodonAccount
		params["Title"] = "Minecraft"
		params["MinecraftServers"] = servers
	} else {
		c.Redirect("/")
	}
	c.Render("minecraft/index", params)
	return nil
}

func GetMinecraftNew(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)

	params := fiber.Map{}
	servers, _ := exarotonGetServersList(config.Storage.Cache)
	if mastodonAccount.UserID != "" {
		params["mastodonAccount"] = mastodonAccount
		params["Title"] = "Minecraft"
		params["MinecraftServers"] = servers
	}
	c.Render("minecraft/new", params)
	return nil
}

func PostMinecraft(c *fiber.Ctx) error {
	config := config.GetConfig()
	mojang := getUserMojang(c.FormValue("username"))
	c.JSON(mojang)
	// Store Mojang in a session
	sess, err := config.Storage.Session.Get(c)
	if err != nil {
		panic(err)
	}
	mojangJson, err := json.Marshal(mojang)
	if err != nil {
		panic(err)
	}
	sess.Set("mojang", string(mojangJson))
	sess.Save()

	c.Redirect("/minecraft/check")

	return nil
}

func GetMinecraftCheck(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	mojangAccount := getUserMojangFromSession(config.Storage.Session, c)

	// Get Mojang Name using Mastodon ID
	previousMojangName, err := config.Storage.Session.Storage.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
	if err != nil {
		panic(err)
	}

	params := fiber.Map{}

	if mastodonAccount.UserID != "" {
		// Required for logged in pages
		params["mastodonAccount"] = mastodonAccount
		params["Title"] = "Minecraft"
		// Specific
		params["PreviousMojangName"] = string(previousMojangName)
		params["MojangId"] = mojangAccount.Id
		params["MojangUsername"] = mojangAccount.Name
		params["MastodonId"] = mastodonAccount.UserID
		params["MastodonUsername"] = mastodonAccount.NickName
	} else {
		c.Redirect("/")
	}

	c.Render("minecraft/check", params)

	return nil
}

func PostMinecraftCreate(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	mojangAccount := getUserMojangFromSession(config.Storage.Session, c)

	// Get from the DB the Mojang username using Mastodon account ID
	previousMojangName, err := config.Storage.Session.Storage.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
	if err != nil {
		panic(err)
	}

	// Remove the previously used username
	if previousMojangName != nil {
		_, err = exarotonRemoveUser(string(previousMojangName))
		if err != nil {
			panic(err)
		}
	}

	// Add the user to our Exaroton servers allowlists
	_, err = exarotonAllowUser(mojangAccount.Name)
	if err != nil {
		c.Render("minecraft/add", fiber.Map{"err": err, "currentPath": c.Path()})
	}

	// Associate Mastodon ID with Mojang Username
	log.Debug("saving username to DB:", "minecraft-%s", mastodonAccount.UserID, mojangAccount.Name)
	config.Storage.Session.Storage.Set(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID), []byte(mojangAccount.Name), 0)
	params := fiber.Map{}

	if mastodonAccount.UserID != "" {
		// Required for logged in pages
		params["Title"] = "Minecraft"
		params["mastodonAccount"] = mastodonAccount
		// Specific
		params["accountName"] = mojangAccount.Name
	} else {
		c.Redirect("/")
	}
	c.Render("minecraft/added", params)
	return nil
}

func getUserMojang(username string) MojangAccount {
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

func exarotonRequestV1(verb string, path string, bodyReq io.Reader) (response []byte, err error) {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	url := fmt.Sprintf("https://api.exaroton.com/v1%s", path)

	// Create a Bearer string by appending string access token
	var bearer = "Bearer " + os.Getenv("EXAROTON_API_KEY")

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
	serversId := strings.Split(os.Getenv("EXAROTON_SERVERS_ID"), ",")

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

func exarotonGetServersList(cache *sqlite3.Storage) (response []models.ExarotonServer, err error) {
	var respJson []byte
	var resp models.ExarotonBody
	var servers []models.ExarotonServer

	// Get exaroton JSON servers list from cache
	respJson, err = cache.Get("exaroton-servers-list-json")
	log.Debug(string(respJson))
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
		json.Unmarshal(respJson, &resp)
		if resp.Success == false {
			log.Error("Error during Exaroton API call", "err", resp.Error, "data", string(resp.Data))
			return
		}
		cache.Set("exaroton-servers-list-json", respJson, 15*time.Minute)
		log.Debug("exaroton-servers-list-json cache saved.")
	}

	err = json.Unmarshal(respJson, &resp)
	if err != nil {
		log.Error(err)
		return nil, err
	}

	if resp.Success == false {
		log.Error("Error during Exaroton API call", "err", resp.Error, "data", string(resp.Data))
	}

	err = json.Unmarshal(resp.Data, &servers)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	// Extract only IDs we want
	serverIDs := strings.Split(os.Getenv("EXAROTON_SERVERS_ID"), ",")
	filteredServers := make([]models.ExarotonServer, 0)
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

func getUserMojangFromSession(store *session.Store, ctx *fiber.Ctx) MojangAccount {
	sess, err := store.Get(ctx)
	if err != nil {
		panic(err)
	}
	var mojangAccount MojangAccount
	err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mojang"))), &mojangAccount)
	if err != nil {
		panic(err)
	}
	return mojangAccount
}
