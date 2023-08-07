package config

import (
	"flag"
	"fmt"
	"mastodon-to-exaroton-oauth2/app/models"
	"os"

	"github.com/charmbracelet/log"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/session"
	"github.com/gofiber/storage/sqlite3"
	_ "github.com/joho/godotenv/autoload"
	"github.com/markbates/goth"
	"github.com/markbates/goth/providers/mastodon"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const (
	APP_BASE_URL         = "APP_BASE_URL"
	BIND_ADDRESS         = "BIND_ADDRESS"
	DATABASE_CACHE_PATH  = "DATABASE_CACHE_PATH"
	DATABASE_PATH        = "DATABASE_PATH"
	EXAROTON_API_KEY     = "EXAROTON_API_KEY"
	EXAROTON_SERVERS_ID  = "EXAROTON_SERVERS_ID"
	MASTODON_URL         = "MASTODON_URL"
	OAUTH2_CLIENT_ID     = "OAUTH2_CLIENT_ID"
	OAUTH2_CLIENT_SECRET = "OAUTH2_CLIENT_SECRET"
	ORG_NAME             = "ORG_NAME"
)

type Config struct {
	App     *fiber.App
	Storage Storage
}

type Storage struct {
	Cache   *sqlite3.Storage
	Session *session.Store
	Blog    *gorm.DB
}

func InitConfig() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	// Create blog DB
	storageBlog, err := gorm.Open(sqlite.Open("blog.sqlite3"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Cannot open blog.sqlite3", "err", err)
	}

	// Migrate the schema
	storageBlog.AutoMigrate(&models.BlogPost{})
}

func GetConfig() Config {
	storageSessions := sqlite3.New(sqlite3.Config{Database: os.Getenv(DATABASE_PATH)})
	cache := sqlite3.New(sqlite3.Config{Database: os.Getenv(DATABASE_CACHE_PATH)})
	storageBlog, err := gorm.Open(sqlite.Open("blog.sqlite3"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal("Cannot open blog.sqlite3", "err", err)
	}

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv(OAUTH2_CLIENT_ID),
			os.Getenv(OAUTH2_CLIENT_SECRET),
			fmt.Sprintf("%s/auth/mastodon/callback", os.Getenv(APP_BASE_URL)),
			fmt.Sprintf("%s", os.Getenv(MASTODON_URL)),
			"read:accounts",
		),
	)

	session := session.New(
		session.Config{
			Storage: storageSessions,
		},
	)

	return Config{
		Storage: Storage{
			Cache:   cache,
			Session: session,
			Blog:    storageBlog,
		},
	}
}
