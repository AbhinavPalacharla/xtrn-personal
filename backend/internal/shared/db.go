package shared

import (
	"database/sql"
	"os"

	db "github.com/AbhinavPalacharla/xtrn-personal/internal/db/sqlc"
	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB
var Q *db.Queries

func init() {
	if ok, err := LoadEnv(); !ok {
		panic(err)
	}

	dbURL := os.Getenv("DB_URL")

	if dbURL == "" {
		panic("Failed to load `DB_URL` from env")
	}

	conn, err := sql.Open("sqlite3", dbURL)
	if err != nil {
		StdErrLogger.Fatalf("Failed to connect to DB at %s - %v\n", dbURL, err)
	}

	DB = conn
	Q = db.New(conn)

	// fmt.Print("✅ DB Connection initialized\n")
}
