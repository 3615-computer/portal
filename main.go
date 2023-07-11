package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/template/html/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"
	gf "github.com/shareed2k/goth_fiber"
)

const (
	OAUTH2_CLIENT_ID     = "OAUTH2_CLIENT_ID"
	OAUTH2_CLIENT_SECRET = "OAUTH2_CLIENT_SECRET"
	OAUTH2_REDIRECT_URL  = "OAUTH2_REDIRECT_URL"
	MASTODON_DOMAIN      = "MASTODON_DOMAIN"
	EXAROTON_API_KEY     = "EXAROTON_API_KEY"
	EXAROTON_SERVERS_ID  = "EXAROTON_SERVERS_ID"

	OAUTH2_REDIRECT_URL_DEFAULT = "urn:ietf:wg:oauth:2.0:oob"
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")
	store := session.New()

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv(OAUTH2_CLIENT_ID),
			os.Getenv(OAUTH2_CLIENT_SECRET),
			"http://127.0.0.1:3000/auth/mastodon/callback",
			fmt.Sprintf("https://%s/", os.Getenv(MASTODON_DOMAIN)),
			"read:accounts",
		),
	)

	app.Get("/auth/:provider/callback", func(ctx *fiber.Ctx) error {
		mastodon, err := gf.CompleteUserAuth(ctx, gf.CompleteUserAuthOptions{ShouldLogout: false})
		if err != nil {
			return err
		}
		ctx.JSON(mastodon)
		log.Debugf("Mastodon account: %v", mastodon)

		// Store User in a session
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		mastodonJson, err := json.Marshal(mastodon)
		if err != nil {
			panic(err)
		}
		sess.Set("mastodon", string(mastodonJson))
		sess.Save()

		ctx.Redirect("/mojang")

		return nil
	})

	app.Get("/logout/:provider", func(ctx *fiber.Ctx) error {
		gf.Logout(ctx)
		ctx.Redirect("/")
		return nil
	})

	app.Get("/auth/:provider", func(ctx *fiber.Ctx) error {
		if gothicUser, err := gf.CompleteUserAuth(ctx, gf.CompleteUserAuthOptions{ShouldLogout: false}); err == nil {
			ctx.JSON(gothicUser)
		} else {
			gf.BeginAuthHandler(ctx)
		}
		return nil
	})

	app.Get("/mojang", func(ctx *fiber.Ctx) error {
		ctx.Render("mojang", fiber.Map{}, "layouts/main")
		return nil
	})

	app.Post("/mojang", func(ctx *fiber.Ctx) error {
		mojang := GetUserMojang(ctx.FormValue("mojang_username"))
		ctx.JSON(mojang)
		// Store Mojang in a session
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		mojangJson, err := json.Marshal(mojang)
		if err != nil {
			panic(err)
		}
		sess.Set("mojang", string(mojangJson))
		sess.Save()

		ctx.Redirect("/check")

		return nil
	})

	app.Get("/check", func(ctx *fiber.Ctx) error {
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		var mojangAccount MojangAccount
		var mastodonAccount goth.User
		err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mojang"))), &mojangAccount)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mastodon"))), &mastodonAccount)
		if err != nil {
			panic(err)
		}

		ctx.Render("check", fiber.Map{
			"MojangId":         mojangAccount.Id,
			"MojangUsername":   mojangAccount.Name,
			"MastodonId":       mastodonAccount.UserID,
			"MastodonUsername": mastodonAccount.NickName,
		}, "layouts/main")

		return nil
	})

	app.Post("/add", func(ctx *fiber.Ctx) error {
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		var mojangAccount MojangAccount
		var mastodonAccount goth.User
		err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mojang"))), &mojangAccount)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mastodon"))), &mastodonAccount)
		if err != nil {
			panic(err)
		}

		exarotonAllowUser(mojangAccount.Name)

		return nil
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		ctx.Render("index", fiber.Map{
			"auth_url": "http://127.0.0.1:3000/auth/mastodon/",
		}, "layouts/main")
		return nil
	})

	if err := app.Listen("127.0.0.1:3000"); err != nil {
		log.Fatal(err)
	}
}
