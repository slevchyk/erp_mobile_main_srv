package main

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/debug"
	"golang.org/x/sys/windows/svc/eventlog"
	"golang.org/x/sys/windows/svc/mgr"

	_ "github.com/lib/pq"
	"github.com/slevchyk/erp_mobile_main_srv/models"
	"github.com/slevchyk/erp_mobile_main_srv/dbase"
)

var cfg models.Config
var db *sql.DB
var eLog debug.Log

type myService struct{}

func init() {
	var err error
	var dir string

	for k, v := range os.Args {
		if v == "-dir" && len(os.Args) > k {
			dir = os.Args[k+1]
			dir, _ = strconv.Unquote(dir)
			dir += "/"
		}
	}

	cfg, err = loadConfiguration(fmt.Sprintf("%sconfig.json", dir))
	if err != nil {
		log.Fatal("Can't load configuration file config.json", err.Error())
	}

	db, _ = dbase.ConnectDB(cfg.DB)
	dbase.InitDB(db)
}

func main() {

	var err error

	isIntSess, err := svc.IsAnInteractiveSession()
	if err != nil {
		log.Fatalf("failed to determine if we are running in an interactive session: %v", err)
	}
	if !isIntSess {
		runService(cfg.WinService.Name, false)
		return
	}

	if len(os.Args) == 2 {

		cmd := strings.ToLower(os.Args[1])
		switch cmd {
		case "debug":
			runService(cfg.WinService.Name, true)
			return
		case "install":
			err = installService(cfg.WinService.Name, cfg.WinService.LongName, cfg.WinService.Description)
		case "remove":
			err = removeService(cfg.WinService.Name)
		case "start":
			err = startService(cfg.WinService.Name)
		case "stop":
			err = controlService(cfg.WinService.Name, svc.Stop, svc.Stopped)
		case "pause":
			err = controlService(cfg.WinService.Name, svc.Pause, svc.Paused)
		case "continue":
			err = controlService(cfg.WinService.Name, svc.Continue, svc.Running)
		default:
			log.Fatalf("unknown command %s", cmd)
		}

		if err != nil {
			log.Fatalf("failed to %s %s: %v", cmd, cfg.WinService.Name, err)
		}

		return
	}

	webApp()
}

func webApp()  {
	defer db.Close()

	http.Handle("/favicon.ico", http.NotFoundHandler())
	http.HandleFunc("/api/getdbsettings", basicAuth(settingsHandler))

	err := http.ListenAndServe(":8811", nil)
	if err != nil {
		panic(err)
	}
}

func exePath() (string, error) {
	prog := os.Args[0]
	p, err := filepath.Abs(prog)
	if err != nil {
		return "", err
	}
	fi, err := os.Stat(p)
	if err == nil {
		if !fi.Mode().IsDir() {
			return p, nil
		}
		err = fmt.Errorf("%s is directory", p)
	}
	if filepath.Ext(p) == "" {
		p += ".exe"
		fi, err := os.Stat(p)
		if err == nil {
			if !fi.Mode().IsDir() {
				return p, nil
			}
			err = fmt.Errorf("%s is directory", p)
		}
	}
	return "", err
}

func installService(name, lname, desc string) error {
	exepath, err := exePath()
	if err != nil {
		return err
	}
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err == nil {
		s.Close()
		return fmt.Errorf("service %s already exists", name)
	}

	wd, err := os.Getwd()
	if err != nil {
		s.Close()
		return err
	}
	wd = strconv.Quote(wd)
	log.Println(wd)

	s, err = m.CreateService(name, exepath, mgr.Config{DisplayName: lname, Description: desc}, "-dir", wd, "is", "auto-started")
	if err != nil {
		return err
	}
	defer s.Close()
	err = eventlog.InstallAsEventCreate(name, eventlog.Error|eventlog.Warning|eventlog.Info)
	if err != nil {
		s.Delete()
		return fmt.Errorf("SetupEventLogSource() failed: %s", err)
	}
	return nil
}

func removeService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("service %s is not installed", name)
	}
	defer s.Close()
	err = s.Delete()
	if err != nil {
		return err
	}
	err = eventlog.Remove(name)
	if err != nil {
		return fmt.Errorf("RemoveEventLogSource() failed: %s", err)
	}
	return nil
}
func startService(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	err = s.Start("is", "manual-started")
	if err != nil {
		return fmt.Errorf("could not start service: %v", err)
	}
	return nil
}

func controlService(name string, c svc.Cmd, to svc.State) error {
	m, err := mgr.Connect()
	if err != nil {
		return err
	}
	defer m.Disconnect()
	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("could not access service: %v", err)
	}
	defer s.Close()
	status, err := s.Control(c)
	if err != nil {
		return fmt.Errorf("could not send control=%d: %v", c, err)
	}
	timeout := time.Now().Add(10 * time.Second)
	for status.State != to {
		if timeout.Before(time.Now()) {
			return fmt.Errorf("timeout waiting for service to go to state=%d", to)
		}
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("could not retrieve service status: %v", err)
		}
	}
	return nil
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		eLog = debug.New(name)
	} else {
		eLog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer eLog.Close()

	eLog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myService{})
	if err != nil {
		log.Printf("%s service failed: %v", name, err)
		eLog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	eLog.Info(1, fmt.Sprintf("%s service stopped", name))
}

func (m *myService) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue

	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	go webApp()

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// Testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
			default:
				eLog.Info(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}

	changes <- svc.Status{State: svc.StopPending}
	return
}

func loadConfiguration(file string) (models.Config, error) {
	var config models.Config

	cfgFile, err := ioutil.ReadFile(file)
	if err != nil {
		log.Println(err)
		return config, err
	}

	err = json.Unmarshal(cfgFile, &config)
	if err != nil {
		log.Println(err)
		return config, err
	}

	return config, nil
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
