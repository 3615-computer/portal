package config

import (
	"flag"
	"fmt"
	"mastodon-services/app/models"
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

type Config struct {
	App     *fiber.App
	Storage Storage
}

type Storage struct {
	Cache    *sqlite3.Storage
	Session  *session.Store
	Database *gorm.DB
}

func InitConfig() {
	flag.Parse()
	log.SetLevel(log.DebugLevel)

	// Create main DB
	db, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_PATH")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot open %s", os.Getenv("DATABASE_PATH")), "err", err)
	}

	// Migrate schemas
	db.AutoMigrate(
		&models.BlogPost{},
		&models.User{},
	)
	// Run migration queries
	db.Exec("UPDATE blog_posts SET visibility = 0 WHERE visibility IS NULL")
}

func GetConfig() Config {
	storageSessions := sqlite3.New(sqlite3.Config{Database: os.Getenv("DATABASE_PATH_SESSION")})
	cache := sqlite3.New(sqlite3.Config{Database: os.Getenv("DATABASE_PATH_CACHE")})
	db, err := gorm.Open(sqlite.Open(os.Getenv("DATABASE_PATH")), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot open %s", os.Getenv("DATABASE_PATH")), "err", err)
	}

	goth.UseProviders(
		mastodon.NewCustomisedURL(
			os.Getenv("OAUTH2_CLIENT_ID"),
			os.Getenv("OAUTH2_CLIENT_SECRET"),
			fmt.Sprintf("%s/auth/mastodon/callback", os.Getenv("APP_BASE_URL")),
			fmt.Sprintf("%s", os.Getenv("MASTODON_URL")),
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
			Cache:    cache,
			Session:  session,
			Database: db,
		},
	}
}
