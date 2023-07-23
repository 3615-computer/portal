package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	"github.com/gofiber/template/html/v2"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"

	"github.com/gomarkdown/markdown"
	mdHtml "github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"github.com/microcosm-cc/bluemonday"
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

type Author struct {
	Id   string
	Name string
}

type BlogPost struct {
	Id           string
	Author       Author
	Title        string
	Body         string
	CreationDate time.Time
}

func main() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	// From github.com/gofiber/storage/sqlite3
	storageSessions := sqlite3.New(sqlite3.Config{Database: os.Getenv(DATABASE_PATH)})
	storageBlog := sqlite3.New(sqlite3.Config{Database: os.Getenv(DATABASE_PATH), Table: "blog_posts"})

	// TODO: create a blogpost database with real fields (not kv)

	engine := html.New("./views", ".html")
	app := fiber.New(fiber.Config{
		Views:             engine,
		PassLocalsToViews: true,
		ViewsLayout:       "layouts/main",
	})
	app.Static("/", "./public")
	store := session.New(
		session.Config{
			Storage: storageSessions,
		},
	)
	app.Use(func(ctx *fiber.Ctx) error {
		ctx.Locals(
			"ORG_NAME", os.Getenv(ORG_NAME),
		)
		return ctx.Next()
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
		params["Title"] = "Home"
		params["mastodonAccount"] = mastodonAccount
		ctx.Render("index", params)
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

	app.Get("/minecraft", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		params := fiber.Map{}
		servers, _ := exarotonGetServersList()
		if mastodonAccount.UserID != "" {
			params["mastodonAccount"] = mastodonAccount
			params["Title"] = "Minecraft"
			params["MinecraftServers"] = servers
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("minecraft/index", params)
		return nil
	})

	app.Get("/minecraft/new", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		params := fiber.Map{}
		servers, _ := exarotonGetServersList()
		if mastodonAccount.UserID != "" {
			params["mastodonAccount"] = mastodonAccount
			params["Title"] = "Minecraft"
			params["MinecraftServers"] = servers
		}
		ctx.Render("minecraft/new", params)
		return nil
	})

	app.Post("/minecraft", func(ctx *fiber.Ctx) error {
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

		ctx.Redirect("/minecraft/check")

		return nil
	})

	app.Get("/minecraft/check", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		mojangAccount := getUserMojangFromSession(store, ctx)

		// Get Mojang Name using Mastodon ID
		previousMojangName, err := storageSessions.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
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
			ctx.Redirect("/")
		}

		ctx.Render("minecraft/check", params)

		return nil
	})

	app.Post("/minecraft/create", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		mojangAccount := getUserMojangFromSession(store, ctx)

		// Get from the DB the Mojang username using Mastodon account ID
		previousMojangName, err := storageSessions.Get(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID))
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
			ctx.Render("minecraft/add", fiber.Map{"err": err, "currentPath": ctx.Path()})
		}

		// Associate Mastodon ID with Mojang Username
		log.Debug("saving username to DB:", "minecraft-%s", mastodonAccount.UserID, mojangAccount.Name)
		storageSessions.Set(fmt.Sprintf("minecraft-%s", mastodonAccount.UserID), []byte(mojangAccount.Name), 0)
		params := fiber.Map{}

		if mastodonAccount.UserID != "" {
			// Required for logged in pages
			params["Title"] = "Minecraft"
			params["mastodonAccount"] = mastodonAccount
			// Specific
			params["accountName"] = mojangAccount.Name
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("minecraft/added", params)
		return nil
	})

	app.Get("/miniblog", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		params := fiber.Map{}
		if mastodonAccount.UserID != "" {
			params["mastodonAccount"] = mastodonAccount
			params["Title"] = "Miniblog"
		} else {
			ctx.Redirect("/")
		}
		ctx.Render("miniblog/index", params)
		return nil
	})

	app.Get("/miniblog/new", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		params := fiber.Map{}
		if mastodonAccount.UserID != "" {
			params["mastodonAccount"] = mastodonAccount
			params["Title"] = "Miniblog"
		}
		ctx.Render("miniblog/new", params)
		return nil
	})

	app.Get("/miniblog/:user/posts/:post", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)
		userId := ctx.Params("user")
		postId := ctx.Params("post")

		titleKey := fmt.Sprintf("blog-%s-%s-title", userId, postId)
		bodyKey := fmt.Sprintf("blog-%s-%s-body", userId, postId)

		title, _ := storageBlog.Get(titleKey)
		body, _ := storageBlog.Get(bodyKey)

		params := fiber.Map{}

		params["Title"] = fmt.Sprintf("Post: \"%s\"", title)
		params["PostTitle"] = string(title)
		params["PostBody"] = template.HTML(string(mdToHTML(body)))
		params["mastodonAccount"] = mastodonAccount

		ctx.Render("miniblog/posts/show", params)
		return nil
	})

	app.Post("/miniblog", func(ctx *fiber.Ctx) error {
		mastodonAccount := getUserMastodonFromSession(store, ctx)

		author := Author{
			Id:   mastodonAccount.UserID,
			Name: strings.ToLower(mastodonAccount.Name),
		}
		blogPost := BlogPost{
			Id:           uuid.New().String(),
			Author:       author, // Not saved
			Title:        ctx.FormValue("title"),
			Body:         ctx.FormValue("body"),
			CreationDate: time.Now(), // Not saved
		}

		err := saveBlogPost(storageBlog, blogPost)
		if err != nil {
			panic(err)
		}

		return ctx.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.Author.Id, blogPost.Id))
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

func getUserMastodonFromSId(id string, store *session.Store, ctx *fiber.Ctx) goth.User {
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

func saveBlogPost(storage *sqlite3.Storage, post BlogPost) error {
	titleKey := fmt.Sprintf("blog-%s-%s-title", post.Author.Id, post.Id)
	bodyKey := fmt.Sprintf("blog-%s-%s-body", post.Author.Id, post.Id)

	storage.Set(titleKey, []byte(post.Title), 0)
	log.Debug("Blog post title saved", "id", post.Id, "userId", post.Author.Id, "title", post.Title)
	storage.Set(bodyKey, []byte(post.Body), 0)
	log.Debug("Blog post content saved", "id", post.Id, "userId", post.Author.Id, "content", post.Body)

	log.Debug("Blog post", "id", post.Id, "userId", post.Author.Id)

	return nil
}

func mdToHTML(md []byte) []byte {
	// create markdown parser with extensions
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs | parser.NoEmptyLineBeforeBlock
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse(md)

	// create HTML renderer with extensions
	htmlFlags := mdHtml.CommonFlags | mdHtml.HrefTargetBlank
	opts := mdHtml.RendererOptions{Flags: htmlFlags}
	renderer := mdHtml.NewRenderer(opts)

	return bluemonday.UGCPolicy().SanitizeBytes(markdown.Render(doc, renderer))
}
