package main

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os/user"
	"path/filepath"
	"runtime"
)

const schema = `CREATE TABLE IF NOT EXISTS cache (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
	cookie_value TEXT,
	cookie_name TEXT,
	browser_path TEXT,
	os TEXT,
	linux_distro TEXT,
	folder_path TEXT
);`

// cache for general app.
type Cache struct {
	id           int
	cookie       http.Cookie
	browser_path string
	os           string
	linux_distro string
	folder_path  string
}

// Return a new DB to the user with WAL config.
func initDB(path string) (*sql.DB, error) {
	var config string = "?_journal_mode=WAL&_busy_timeout=5000&_synchronous=NORMAL&_cache_size=-64000&_foreign_keys=ON"
	var dest string = path + config
	db, err := sql.Open("sqlite3", dest)
	if err != nil {
		log.Printf("There was an error with opening the DB - dest: %s\n ERROR: %v\n", dest, err)
		return nil, err
	}
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	if err := db.Ping(); err != nil {
		log.Printf("There was an issue with pinging the DB. %v\n", err)
		return nil, err
	}
	return db, nil
}
func initSchema(db *sql.DB) error {
	_, err := db.Exec(schema)
	return err
}
func initCache(db *sql.DB) error {
	cookie_name, cookie_val := GetSessionCookie().Name, GetSessionCookie().Value
	// check for linux machine, if so determine distro.
	linux_distro := ""
	if runtime.GOOS == "linux" {
		linux_distro = GetDistro()
	}
	browser_path := FireFoxPath()
	user, _ := user.Current()
	homeDir := user.HomeDir
	folder_path := filepath.Join(homeDir, "assignments")
	_, err := db.Exec("INSERT into cache(cookie_name, cookie_value, browser_path, os, linux_distro, folder_path) VALUES(?, ?, ?, ?, ?, ?)", cookie_name, cookie_val, browser_path, runtime.GOOS, linux_distro, folder_path)
	if err != nil {
		return err
	}
	return nil
}
func GetCache(cache *Cache, db *sql.DB, id int) error {
	var query string = "SELECT * from cache where id = ?"
	row := db.QueryRow(query, id)
	err := row.Scan(&cache.id, &cache.cookie.Name, &cache.cookie.Value, &cache.browser_path, &cache.os, &cache.linux_distro, &cache.folder_path)
	if err != nil {
		return err
	}
	return nil
}
