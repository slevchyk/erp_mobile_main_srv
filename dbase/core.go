package dbase

import (
	"database/sql"
	"fmt"
	"github.com/slevchyk/worker_srv/models"
	"log"
)

func ConnectDB(cfg models.DBConfig) (*sql.DB, error) {

	dbConnection := fmt.Sprintf("postgres://%v:%v@localhost/%v?sslmode=disable", cfg.User, cfg.Password, cfg.Name)
	db, err := sql.Open("postgres", dbConnection)

	return db, err
}

func InitDB(db *sql.DB) {

	_, err := db.Exec(`
		CREATE TABLE IF NOT EXISTS cloud_settings (
			id SERIAL PRIMARY KEY,
			alias TEXT,
			srv_ip TEXT,			
			srv_user TEXT,
			srv_password TEXT);`)
	if err != nil {
		log.Fatal(err)
	}

	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS cloud_users (
			id SERIAL PRIMARY KEY,
			id_settings INT REFERENCES cloud_settings(id),
			phone TEXT,			
			pin INT DEFAULT 0);`)
	if err != nil {
		log.Fatal(err)
	}


}
