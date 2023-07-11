package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
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

	storage := sqlite3.New() // From github.com/gofiber/storage/sqlite3
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views: engine,
	})
	app.Static("/", "./public")
	store := session.New(
		session.Config{
			Storage: storage,
		},
	)

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv(OAUTH2_CLIENT_ID),
			os.Getenv(OAUTH2_CLIENT_SECRET),
			"http://localhost:3000/auth/mastodon/callback",
			fmt.Sprintf("https://%s/", os.Getenv(MASTODON_DOMAIN)),
			"read:accounts",
		),
	)

	app.Get("/", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		params := fiber.Map{"AuthUrl": "http://localhost:3000/auth/mastodon/"}

		if mastodonAccount.AccessToken != "" {
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["ExarotonAddUrl"] = "http://localhost:3000/mojang/"
			params["LogoutUrl"] = "http://localhost:3000/logout/mastodon/"
		}
		ctx.Render("index", params, "layouts/main")
		return nil
	})

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

		ctx.Redirect("/")

		return nil
	})

	app.Get("/logout/:provider", func(ctx *fiber.Ctx) error {
		gf.Logout(ctx)
		sess, err := store.Get(ctx)
		if err != nil {
			panic(err)
		}
		sess.Destroy()
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
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		params := fiber.Map{"AuthUrl": "http://localhost:3000/auth/mastodon/"}
		if mastodonAccount.AccessToken != "" {
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["LogoutUrl"] = "http://localhost:3000/logout/mastodon/"
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("mojang", params, "layouts/main")
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
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		mojangAccount := getUserMojangFromSession(store, ctx)

		ctx.Render("check", fiber.Map{
			"MojangId":         mojangAccount.Id,
			"MojangUsername":   mojangAccount.Name,
			"MastodonId":       mastodonAccount.UserID,
			"MastodonUsername": mastodonAccount.NickName,
		}, "layouts/main")

		return nil
	})

	app.Post("/add", func(ctx *fiber.Ctx) error {
		err := exarotonAllowUser(getUserMojangFromSession(store, ctx).Name)
		if err != nil {
			ctx.Render("exaroton/add", fiber.Map{"err": err}, "layouts/main")
		}
		params := fiber.Map{"accountName": getUserMojangFromSession(store, ctx).Name}
		ctx.Render("exaroton/add", params, "layouts/main")
		return nil
	})

	if err := app.Listen("localhost:3000"); err != nil {
		log.Fatal(err)
	}
}

func getUserMastodonFromSession(store *session.Store, ctx *fiber.Ctx) goth.User {
	sess, err := store.Get(ctx)
	if err != nil {
		panic(err)
	}
	var mastodonAccount goth.User
	err = json.Unmarshal([]byte(fmt.Sprint(sess.Get("mastodon"))), &mastodonAccount)
	if err != nil {
		return goth.User{}
	}
	return mastodonAccount
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
