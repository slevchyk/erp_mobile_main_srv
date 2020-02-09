package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	_ "github.com/lib/pq"
	"github.com/slevchyk/erp_mobile_main_srv/dbase"
	"github.com/slevchyk/erp_mobile_main_srv/models"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

var db *sql.DB
var cfg models.Config

func init() {

	cfg = models.Config{
		Auth: models.AuthConfig{
			User:     "mobile",
			Password: "Dq4fS^J&^nqQ(fg4",
		},
		DB:   models.DBConfig{
			Name:     "worker",
			User:     "postgres",
			Password: "sierra",
		},
	}

	db, _ = dbase.ConnectDB(cfg.DB)
	dbase.InitDB(db)
}

func main()  {
	defer db.Close()

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/api/getdbsettings", basicAuth(settingsHandler))

	err := http.ListenAndServe(":7132", nil)
	if err != nil {
		panic(err)
	}
}

func basicAuth(pass func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !validate(pair[0], pair[1]) {
			http.Error(w, "authorization failed", http.StatusUnauthorized)
			return
		}

		pass(w, r)
	}
}

func validate(username, password string) bool {
	if username == cfg.Auth.User && password == cfg.Auth.Password {
		return true
	}
	return false
}

func settingsHandler(w http.ResponseWriter, r *http.Request) {

	var responseMsg string
	var err error

	bs, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var requestParams map[string]string

	err = json.Unmarshal(bs, &requestParams)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	phone := requestParams["phone"]
	pinStr := requestParams["pin"]

	if phone == "" {
		responseMsg += fmt.Sprintln("phone is not specified")
	}

	if pinStr == "" {
		responseMsg += fmt.Sprintln("pin is not specified")
	}

	var pin int
	if responseMsg == "" {
		pin, err = strconv.Atoi(pinStr)
		if err != nil {
			responseMsg += fmt.Sprintln(err)
		}
	}

	if responseMsg != "" {
		http.Error(w, responseMsg, http.StatusBadRequest)
		return
	}

	rows, err := dbase.SelectCloudSettingsByPhonePin(db, phone, pin)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var cs models.CloudDBSettings

	if rows.Next() {
		err = dbase.ScanCloudDBSettings(rows, &cs)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	result, err := json.Marshal(cs)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

