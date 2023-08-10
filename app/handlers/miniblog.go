package handlers

import (
	"fmt"
	"mastodon-services/app/config"
	"mastodon-services/app/models"
	"net/url"
	"strings"
	"time"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	_ "github.com/joho/godotenv/autoload"

	"gorm.io/gorm"
)

func GetMiniblog(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)

	var user models.User
	var blogPosts []models.BlogPost
	var latestPosts []models.BlogPost

	config.Storage.Blog.First(&user, models.User{ID: mastodonAccount.UserID})
	config.Storage.Blog.Preload("User").Order("created_at desc").Limit(20).Where("user_id = ?", mastodonAccount.UserID).Find(&blogPosts)

	config.Storage.Blog.Preload("User").Order("created_at desc").Limit(20).Where("user_id != ?", mastodonAccount.UserID).Find(&latestPosts)

	params := fiber.Map{}
	if mastodonAccount.UserID != "" {
		params["mastodonAccount"] = mastodonAccount
		params["Title"] = "Miniblog"
		params["Posts"] = blogPosts
		params["AllPosts"] = latestPosts
	} else {
		c.Redirect("/")
	}
	c.Render("miniblog/index", params)
	return nil
}

func GetMiniblogNew(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)

	params := fiber.Map{}
	if mastodonAccount.UserID != "" {
		params["mastodonAccount"] = mastodonAccount
		params["Title"] = "Miniblog"
	}
	c.Render("miniblog/new", params)
	return nil
}

func GetMiniblogByUsername(c *fiber.Ctx) error {
	return c.Redirect(fmt.Sprintf("/miniblog/%s/posts", c.Params("username")))
}

func GetMiniblogByUsernamePosts(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	username := c.Params("username")

	var user models.User
	var blogPosts []models.BlogPost

	if err := config.Storage.Blog.First(&user, models.User{NickNameURL: username}).Error; err != nil {
		// TODO: user not found
		log.Error(err)
	}

	config.Storage.Blog.Preload("User").Order("created_at desc").Limit(20).Where("user_id = ?", user.ID).Find(&blogPosts)

	params := fiber.Map{}

	params["Title"] = fmt.Sprintf("%s (@%s)", user.Name, user.NickName)
	params["Posts"] = blogPosts
	params["mastodonAccount"] = mastodonAccount

	c.Render("miniblog/posts/index", params)
	return nil
}

func GetMiniblogByUsernamePostsPost(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	postId := c.Params("post")

	var blogPost models.BlogPost
	config.Storage.Blog.Preload("User").First(&blogPost, "id = ?", postId)

	params := fiber.Map{}

	params["Title"] = fmt.Sprintf("%s – %s (@%s)", blogPost.Title, blogPost.User.Name, blogPost.User.NickName)
	params["User"] = blogPost.User
	params["Post"] = blogPost
	params["mastodonAccount"] = mastodonAccount

	c.Render("miniblog/posts/show", params)
	return nil
}

func GetMiniblogByUsernamePostsPostEdit(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	postId := c.Params("post")

	var blogPost models.BlogPost
	config.Storage.Blog.Preload("User").First(&blogPost, "id = ?", postId)

	if mastodonAccount.UserID != blogPost.UserID {
		//TODO: handle error
		return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
	}

	params := fiber.Map{}

	params["Title"] = fmt.Sprintf("%s – %s (@%s)", blogPost.Title, blogPost.User.Name, blogPost.User.NickName)
	params["User"] = blogPost.User
	params["Post"] = blogPost
	params["mastodonAccount"] = mastodonAccount

	c.Render("miniblog/posts/update", params)
	return nil
}

func PostMiniblogByUsernamePostsPostEdit(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	postId := c.Params("post")

	var blogPost models.BlogPost
	config.Storage.Blog.Preload("User").First(&blogPost, "id = ?", postId)

	if mastodonAccount.UserID != blogPost.UserID {
		//TODO: handle error
		return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
	}

	blogPost.Body = c.FormValue("body")
	blogPost.Title = c.FormValue("title")

	err := saveBlogPost(config.Storage.Blog, blogPost)
	if err != nil {
		panic(err)
	}

	return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
}

func GetMiniblogByUsernamePostsPostDelete(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	postId := c.Params("post")

	var blogPost models.BlogPost
	config.Storage.Blog.Preload("User").First(&blogPost, "id = ?", postId)

	if mastodonAccount.UserID != blogPost.UserID {
		//TODO: handle error
		return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
	}

	params := fiber.Map{}

	params["Title"] = fmt.Sprintf("%s – %s (@%s)", blogPost.Title, blogPost.User.Name, blogPost.User.NickName)
	params["Post"] = blogPost

	c.Render("miniblog/posts/delete", params)
	return nil
}

func PostMiniblogByUsernamePostsPostDelete(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)
	postId := c.Params("post")

	var blogPost models.BlogPost
	config.Storage.Blog.Preload("User").First(&blogPost, "id = ?", postId)

	if mastodonAccount.UserID != blogPost.UserID {
		//TODO: handle error
		return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
	}

	if err := config.Storage.Blog.Delete(&blogPost).Error; err != nil {
		log.Fatal("Error during blog post delete", "id", blogPost.ID, "user", blogPost.User, "error", err)
	}

	return c.Redirect("/miniblog/")
}

func PostMiniblog(c *fiber.Ctx) error {
	config := config.GetConfig()
	mastodonAccount := models.GetUserMastodonFromSession(config.Storage.Session, c)

	blogPost := models.BlogPost{
		ID: uuid.New().String(),
		User: models.User{
			ID:          mastodonAccount.UserID,
			Name:        mastodonAccount.Name,
			NickName:    mastodonAccount.NickName,
			NickNameURL: url.QueryEscape(strings.ToLower(mastodonAccount.NickName)),
		},
		Title:        c.FormValue("title"),
		Body:         c.FormValue("body"),
		CreationDate: time.Now(),
	}

	err := saveBlogPost(config.Storage.Blog, blogPost)
	if err != nil {
		panic(err)
	}

	return c.Redirect(fmt.Sprintf("/miniblog/%s/posts/%s", blogPost.User.NickNameURL, blogPost.ID))
}

func saveBlogPost(db *gorm.DB, post models.BlogPost) error {
	if err := db.Save(&post).Error; err != nil {
		log.Fatal("Error during blog post save", "id", post.ID, "user", post.User)
	}

	log.Debug("Blog post saved", "id", post.ID, "user", post.User)
	return nil
}
