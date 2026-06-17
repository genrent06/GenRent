package main

import (
	"database/sql"
	"flag"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
)

func main() {
	_ = godotenv.Load()
	flag.Parse()
	args := flag.Args()

	if len(args) < 1 {
		log.Fatal("Usage: migrate <up|down|reset|status|version> [GOOSE_DBSTRING]")
	}

	// Register PostgreSQL dialect
	goose.SetDialect("postgres")

	// Get database URL from environment or argument
	dbString := os.Getenv("DATABASE_URL")
	if len(args) > 1 {
		dbString = args[1]
	}
	if dbString == "" {
		log.Fatal("DATABASE_URL not set and not provided as argument")
	}

	// Ensure migrations directory exists
	if err := os.MkdirAll("migrations", 0o755); err != nil {
		log.Fatalf("Failed to create migrations directory: %v", err)
	}

	// Open database connection using pgx driver
	db, err := sql.Open("pgx", dbString)
	if err != nil {
		log.Fatalf("goose: failed to open DB: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Fatalf("goose: failed to close DB: %v", err)
		}
	}()

	// Test connection
	if err := db.Ping(); err != nil {
		log.Fatalf("goose: failed to ping DB: %v", err)
	}

	command := args[0]

	switch command {
	case "up":
		if err := goose.Up(db, "migrations"); err != nil {
			log.Fatalf("goose up failed: %v", err)
		}
		log.Println("✓ Migrations applied successfully")
	case "down":
		if err := goose.Down(db, "migrations"); err != nil {
			log.Fatalf("goose down failed: %v", err)
		}
		log.Println("✓ Migration rolled back successfully")
	case "reset":
		if err := goose.Reset(db, "migrations"); err != nil {
			log.Fatalf("goose reset failed: %v", err)
		}
		log.Println("✓ All migrations reset successfully")
	case "status":
		if err := goose.Status(db, "migrations"); err != nil {
			log.Fatalf("goose status failed: %v", err)
		}
	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			log.Fatalf("goose: failed to get version: %v", err)
		}
		log.Printf("Current migration version: %d\n", version)
	default:
		log.Fatalf("goose: unknown command: %s", command)
	}
}
