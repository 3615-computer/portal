package handlers

import (
	"encoding/json"
	"html/template"
	"mastodon-services/app/config"
	"mastodon-services/app/models"
	"net/url"
	"strings"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/markbates/goth"
	gf "github.com/shareed2k/goth_fiber"
)

func GetAuthProviderCallback(c *fiber.Ctx) error {
	store := getSessionStore()
	mastodonUser, err := gf.CompleteUserAuth(c, gf.CompleteUserAuthOptions{ShouldLogout: false})
	if err != nil {
		return err
	}
	c.JSON(mastodonUser)
	log.Debugf("Mastodon account: %v", mastodonUser)

	// Store User in a session
	sess, err := store.Get(c)
	if err != nil {
		log.Fatal(err)
	}
	mastodonUserJson, err := json.Marshal(mastodonUser)
	if err != nil {
		log.Fatal(err)
	}
	sess.Set("mastodon", string(mastodonUserJson))
	sess.Save()

	if err := saveUser(mastodonUser); err != nil {
		log.Fatal(err)
	}

	c.Redirect("/")

	return nil
}

func GetLogoutProvider(c *fiber.Ctx) error {
	store := getSessionStore()
	gf.Logout(c)
	sess, err := store.Get(c)
	if err != nil {
		panic(err)
	}
	sess.Destroy()
	c.Redirect("/")
	return nil
}

func GetAuthProvider(c *fiber.Ctx) error {
	if gothicUser, err := gf.CompleteUserAuth(c, gf.CompleteUserAuthOptions{ShouldLogout: false}); err == nil {
		c.JSON(gothicUser)
	} else {
		gf.BeginAuthHandler(c)
	}
	return nil
}

func getSessionStore() *session.Store {
	config := config.GetConfig()

	return session.New(
		session.Config{
			Storage: config.Storage.Session.Storage,
		},
	)
}

func saveUser(mUser goth.User) error {
	config := config.GetConfig()

	user := models.User{
		ID:          mUser.UserID,
		AvatarURL:   mUser.AvatarURL,
		Description: template.HTML(mUser.RawData["note"].(string)),
		FirstName:   mUser.FirstName,
		LastName:    mUser.LastName,
		Name:        mUser.Name,
		NickName:    mUser.NickName,
		NickNameURL: url.QueryEscape(strings.ToLower(mUser.NickName)),
		Provider:    mUser.Provider,
		UserID:      mUser.UserID,
	}

	if err := config.Storage.Database.Save(&user).Error; err != nil {
		log.Fatal("Error while creating user", "id", mUser.UserID, "name", mUser.Name, "nickname", mUser.NickName)
	}

	log.Debug("User saved", "id", mUser.UserID, "name", mUser.Name, "nickname", mUser.NickName)
	return nil
}
