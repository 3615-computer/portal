package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/charmbracelet/log"

	"golang.org/x/oauth2"
)

const (
	OAUTH2_CLIENT_ID     = "OAUTH2_CLIENT_ID"
	OAUTH2_CLIENT_SECRET = "OAUTH2_CLIENT_SECRET"
	OAUTH2_REDIRECT_URL  = "OAUTH2_REDIRECT_URL"
	MASTODON_DOMAIN      = "MASTODON_DOMAIN"

	OAUTH2_REDIRECT_URL_DEFAULT = "urn:ietf:wg:oauth:2.0:oob"
)

type MastodonAccount struct {
	Id string `json:"id"`
}

type MojangAccount struct {
	Id string `json:"id"`
}

func main() {
	ctx := context.Background()
	flag.Parse()

	log.SetLevel(log.DebugLevel)

	var (
		Oauth2ClientId     string
		Oauth2ClientSecret string
		Oauth2RedirectUrl  string
		MastodonBaseUrl    string
		ok                 bool
	)

	// OAUTH2_CLIENT_ID
	Oauth2ClientId, ok = os.LookupEnv(OAUTH2_CLIENT_ID)
	if !ok {
		log.Fatalf("Value not found : %s", OAUTH2_CLIENT_ID)
	}
	// OAUTH2_CLIENT_SECRET
	Oauth2ClientSecret, ok = os.LookupEnv(OAUTH2_CLIENT_SECRET)
	if !ok {
		log.Fatalf("Value not found : %s", OAUTH2_CLIENT_SECRET)
	}
	// OAUTH2_REDIRECT_URL
	_, ok = os.LookupEnv(OAUTH2_REDIRECT_URL)
	if !ok {
		os.Setenv(OAUTH2_REDIRECT_URL, OAUTH2_REDIRECT_URL_DEFAULT)
		Oauth2RedirectUrl = OAUTH2_REDIRECT_URL_DEFAULT
		log.Infof("Value not found : %s, using %s", OAUTH2_REDIRECT_URL, OAUTH2_REDIRECT_URL_DEFAULT)
	} else {
		Oauth2RedirectUrl = os.Getenv(OAUTH2_REDIRECT_URL)
	}
	// MASTODON_DOMAIN
	MastodonBaseUrl, ok = os.LookupEnv(MASTODON_DOMAIN)
	if !ok {
		log.Fatalf("Value not found : %s", MASTODON_DOMAIN)
	}

	log.Debugf("OAUTH2_REDIRECT_URL : %s", Oauth2RedirectUrl)
	conf := &oauth2.Config{
		ClientID:     Oauth2ClientId,
		ClientSecret: Oauth2ClientSecret,
		Scopes:       []string{"read:accounts"},
		RedirectURL:  Oauth2RedirectUrl,
		Endpoint: oauth2.Endpoint{
			AuthURL:  fmt.Sprintf("https://%s/oauth/authorize", MastodonBaseUrl),
			TokenURL: fmt.Sprintf("https://%s/oauth/token", MastodonBaseUrl),
		},
	}

	// Redirect user to consent page to ask for permission
	// for the scopes specified above.
	url := conf.AuthCodeURL("state", oauth2.AccessTypeOffline)
	log.Infof("Visit the URL for the auth dialog: %v", url)

	// Use the authorization code that is pushed to the redirect
	// URL. Exchange will do the handshake to retrieve the
	// initial access token. The HTTP Client returned by
	// conf.Client will refresh the token as necessary.
	var code string
	log.Info("Enter your OAuth2 token:")
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatal(err)
	}
	tok, err := conf.Exchange(ctx, code)
	if err != nil {
		log.Fatal(err)
	}

	var mojangUsername string
	log.Info("Enter your Mojang username:")
	if _, err := fmt.Scan(&mojangUsername); err != nil {
		log.Fatal(err)
	}

	client := conf.Client(ctx, tok)
	GetUserIdMastodon(*client)
	GetUserIdMojang(mojangUsername)
}

func GetUserIdMastodon(client http.Client) string {
	var account MastodonAccount
	resp, err := client.Get(fmt.Sprintf("https://%s/api/v1/accounts/verify_credentials", os.Getenv(MASTODON_DOMAIN)))
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
	log.Debugf("Mastodon ID: %s", account.Id)
	return account.Id
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
