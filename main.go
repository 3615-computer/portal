package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"
	gf "github.com/shareed2k/goth_fiber"
)

const (
	OAUTH2_CLIENT_ID     = "OAUTH2_CLIENT_ID"
	OAUTH2_CLIENT_SECRET = "OAUTH2_CLIENT_SECRET"
	OAUTH2_REDIRECT_URL  = "OAUTH2_REDIRECT_URL"
	MASTODON_DOMAIN      = "MASTODON_DOMAIN"

	OAUTH2_REDIRECT_URL_DEFAULT = "urn:ietf:wg:oauth:2.0:oob"
)

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	app := fiber.New()

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
		user, err := gf.CompleteUserAuth(ctx, gf.CompleteUserAuthOptions{ShouldLogout: false})
		if err != nil {
			return err
		}
		ctx.JSON(user)

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
		ctx.Render("./public/index.html", fiber.Map{})
		return nil
	})

	app.Post("/mojang", func(ctx *fiber.Ctx) error {
		ctx.JSON(GetUserIdMojang(ctx.FormValue("mojang_username")))
		return nil
	})

	app.Get("/", func(ctx *fiber.Ctx) error {
		ctx.Format("<p><a href='/auth/mastodon'>mastodon</a></p>")
		return nil
	})

	if err := app.Listen("127.0.0.1:3000"); err != nil {
		log.Fatal(err)
	}
}
