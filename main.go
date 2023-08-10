package main

import (
	"fmt"
	"mastodon-services/app/config"
	"mastodon-services/app/handlers"
	"mastodon-services/app/models"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/template/html/v2"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"
)

func main() {
	config.InitConfig()
	config := config.GetConfig()

	engine := html.New("./app/views", ".html")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
		ViewsLayout:       "layouts/main",
	})
	app.Static("/", "./app/public")
	app.Use(func(c *fiber.Ctx) error {
		c.Locals(
			"ORG_NAME", os.Getenv("ORG_NAME"),
		)
		return c.Next()
	})

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv("OAUTH2_CLIENT_ID"),
			os.Getenv("OAUTH2_CLIENT_SECRET"),
			fmt.Sprintf("%s/auth/mastodon/callback", os.Getenv("APP_BASE_URL")),
			fmt.Sprintf("%s", os.Getenv("MASTODON_URL")),
			"read:accounts",
		),
	)

	app.Get("/health", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusOK)
	})

	app.Get("/", func(c *fiber.Ctx) error {
		mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)

		params := fiber.Map{}
		params["Title"] = "Home"
		params["mastodonAccount"] = mastodonAccount
		c.Render("index", params)
		return nil
	})

	authentication := app.Group("/")
	minecraft := app.Group("/minecraft")
	miniblog := app.Group("/miniblog")

	authentication.Get("/auth/:provider/callback", handlers.GetAuthProviderCallback)
	authentication.Get("/logout/:provider", handlers.GetLogoutProvider)
	authentication.Get("/auth/:provider", handlers.GetAuthProvider)

	minecraft.Get("/", handlers.GetMinecraft)
	minecraft.Get("/new", handlers.GetMinecraftNew)
	minecraft.Post("/", handlers.PostMinecraft)
	minecraft.Get("/check", handlers.GetMinecraftCheck)
	minecraft.Post("/create", handlers.PostMinecraftCreate)

	miniblog.Get("/", handlers.GetMiniblog)
	miniblog.Post("/", handlers.PostMiniblog)
	miniblog.Get("/new", handlers.GetMiniblogNew)
	miniblog.Get("/:username/", handlers.GetMiniblogByUsername)
	miniblog.Get("/:username/posts/", handlers.GetMiniblogByUsernamePosts)
	miniblog.Get("/:username/posts/:post", handlers.GetMiniblogByUsernamePostsPost)
	miniblog.Get("/:username/posts/:post/edit", handlers.GetMiniblogByUsernamePostsPostEdit)
	miniblog.Post("/:username/posts/:post/edit", handlers.PostMiniblogByUsernamePostsPostEdit)
	miniblog.Get("/:username/posts/:post/delete", handlers.GetMiniblogByUsernamePostsPostDelete)
	miniblog.Post("/:username/posts/:post/delete", handlers.PostMiniblogByUsernamePostsPostDelete)

	if err := app.Listen(os.Getenv("BIND_ADDRESS")); err != nil {
		log.Fatal(err)
	}
}
