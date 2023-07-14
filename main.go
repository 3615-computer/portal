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
	_ "github.com/joho/godotenv/autoload"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"
	gf "github.com/shareed2k/goth_fiber"
)

const (
	APP_BASE_URL         = "APP_BASE_URL"
	BIND_ADDRESS         = "BIND_ADDRESS"
	DATABASE_PATH        = "DATABASE_PATH"
	EXAROTON_API_KEY     = "EXAROTON_API_KEY"
	EXAROTON_SERVERS_ID  = "EXAROTON_SERVERS_ID"
	MASTODON_URL         = "MASTODON_URL"
	OAUTH2_CLIENT_ID     = "OAUTH2_CLIENT_ID"
	OAUTH2_CLIENT_SECRET = "OAUTH2_CLIENT_SECRET"
	ORG_NAME             = "ORG_NAME"
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	storage := sqlite3.New(sqlite3.Config{Database: os.Getenv(DATABASE_PATH)}) // From github.com/gofiber/storage/sqlite3
	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
	})
	app.Static("/", "./public")
	store := session.New(
		session.Config{
			Storage: storage,
		},
	)
	app.Use(func(c *fiber.Ctx) error {
		c.Locals("ORG_NAME", os.Getenv(ORG_NAME))
		return c.Next()
	})

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv(OAUTH2_CLIENT_ID),
			os.Getenv(OAUTH2_CLIENT_SECRET),
			fmt.Sprintf("%s/auth/mastodon/callback", os.Getenv(APP_BASE_URL)),
			fmt.Sprintf("%s", os.Getenv(MASTODON_URL)),
			"read:accounts",
		),
	)

	app.Get("/health", func(ctx *fiber.Ctx) error {
		return ctx.SendStatus(fiber.StatusOK)
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		params := fiber.Map{}
		params["AuthUrl"] = fmt.Sprintf("%s/auth/mastodon", ctx.BaseURL())
		if mastodonAccount.AccessToken != "" {
			// Required for logged in pages
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["Avatar"] = mastodonAccount.AvatarURL
			params["ExarotonAddUrl"] = fmt.Sprintf("%s/mojang/", ctx.BaseURL())
			params["LogoutUrl"] = fmt.Sprintf("%s/logout/mastodon/", ctx.BaseURL())
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

		params := fiber.Map{}
		params["AuthUrl"] = fmt.Sprintf("%s/auth/mastodon/", ctx.BaseURL())
		if mastodonAccount.AccessToken != "" {
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["Avatar"] = mastodonAccount.AvatarURL
			params["LogoutUrl"] = fmt.Sprintf("%s/logout/mastodon/", ctx.BaseURL())
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("mojang", params, "layouts/main")
		return nil
	})

	app.Post("/mojang", func(ctx *fiber.Ctx) error {
		mojang := GetUserMojang(ctx.FormValue("username"))
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

		// Get Mojang Name using Mastodon ID
		previousMojangName, err := storage.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
		if err != nil {
			panic(err)
		}

		params := fiber.Map{}

		if mastodonAccount.AccessToken != "" {
			// Required for logged in pages
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["Avatar"] = mastodonAccount.AvatarURL
			params["LogoutUrl"] = fmt.Sprintf("%s/logout/mastodon/", ctx.BaseURL())
			// Specific
			params["PreviousMojangName"] = string(previousMojangName)
			params["MojangId"] = mojangAccount.Id
			params["MojangUsername"] = mojangAccount.Name
			params["MastodonId"] = mastodonAccount.UserID
			params["MastodonUsername"] = mastodonAccount.NickName
		} else {
			ctx.Redirect("/")
		}

		ctx.Render("check", params, "layouts/main")

		return nil
	})

	app.Post("/add", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		mojangAccount := getUserMojangFromSession(store, ctx)

		// Get from the DB the Mojang username using Mastodon account ID
		previousMojangName, err := storage.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
		if err != nil {
			panic(err)
		}

		// Remove the previously used username
		if previousMojangName != nil {
			err = exarotonRemoveUser(string(previousMojangName))
			if err != nil {
				panic(err)
			}
		}

		// Add the user to our Exaroton servers allowlists
		err = exarotonAllowUser(mojangAccount.Name)
		if err != nil {
			ctx.Render("exaroton/add", fiber.Map{"err": err, "currentPath": ctx.Path()}, "layouts/main")
		}

		// Associate Mastodon ID with Mojang Username
		log.Debug("saving username to DB:", "minecraft-%s", mastodonAccount.UserID, mojangAccount.Name)
		storage.Set(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID), []byte(mojangAccount.Name), 0)
		params := fiber.Map{}

		if mastodonAccount.AccessToken != "" {
			// Required for logged in pages
			params["IsSignedIn"] = true
			params["Name"] = mastodonAccount.Name
			params["Avatar"] = mastodonAccount.AvatarURL
			params["LogoutUrl"] = fmt.Sprintf("%s/logout/mastodon/", ctx.BaseURL())
			// Specific
			params["accountName"] = mojangAccount.Name
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("exaroton/add", params, "layouts/main")
		return nil
	})

	if err := app.Listen(os.Getenv(BIND_ADDRESS)); err != nil {
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
