package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/neksuhs/flutter-go/foundation/logger"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"go.uber.org/zap"
)

// Package constants
const (
	appName    = "FLUTTER-TESTING"
	appVersion = "0.1.0"
	logFile    = "./logs/flutter.log"
)

var upgrader = websocket.Upgrader{}

// Config stores the apps configuration.
type Config struct {
	Required Required `validate:"required"`
}

// Required stores all the required configuration paramaters for the
type Required struct {
	Dir     string `validate:"required"`
	SpaHost string `validate:"required"`
	WsHost  string `validate:"required"`
}

// spaHandler implements the http.Handler interface
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}

// home returns some text
func home(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "WebSocket home")
}

// flutter upgrades a websocket connection
func flutter(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}
	defer c.Close()
	for {
		mt, message, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		log.Printf("recv: %s", message)
		err = c.WriteMessage(mt, message)
		if err != nil {
			log.Println("write:", err)
			break
		}
	}
}

func main() {
	// construct application logger.
	logger := logger.NewLogger(appName, logFile, true)
	defer logger.Sync()

	// execute run and exit if an error occurs.
	if err := run(logger); err != nil {
		logger.Info("main: run failed, check log file")
		logger.Fatalf("main: run failed: %s \n", err)
	}
}

func run(logger *zap.SugaredLogger) error {
	var (
		config *Config
	)

	// construct and configure Viper
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("toml")
	v.AddConfigPath(".")

	// read configuration file from disck
	if err := v.ReadInConfig(); err != nil {
		return errors.Wrap(err, "fatal error reading config file")
	}

	// set defaults
	v.SetDefault("Required", map[string]interface{}{
		"Dir":     "build",
		"SpaHost": "127.0.0.1:8000",
		"WsHost":  "127.0.0.1:8001",
	})

	// unmarshal config to config struct
	if err := v.Unmarshal(&config); err != nil {
		return errors.Wrap(err, "fatal error unmarshalling config file")
	}

	// validate config struct
	validate := validator.New()
	if err := validate.Struct(config); err != nil {
		return errors.Wrap(err, "fatal error validating config file")
	}

	// construct spaRouter
	spaRouter := mux.NewRouter()
	wsRouter := mux.NewRouter()

	// health check route
	spaRouter.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	spa := spaHandler{staticPath: "build", indexPath: "index.html"}
	spaRouter.PathPrefix("/").Handler(spa)
	wsRouter.HandleFunc("/", home)
	wsRouter.HandleFunc("/flutter", flutter)

	srv := &http.Server{
		Handler:      spaRouter,
		Addr:         config.Required.SpaHost,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	// construct a separate server for the WS connection to avoid conflicts
	// with the SPA routing
	ws := &http.Server{
		Handler:      wsRouter,
		Addr:         config.Required.WsHost,
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  15 * time.Second,
	}

	fmt.Printf("SPA\n")
	fmt.Printf("Flutter app being served at: %v\n", config.Required.SpaHost)
	fmt.Printf("Health check available at: %v/api/health\n", config.Required.SpaHost)
	fmt.Printf("\nWEBSOCKET\n")
	fmt.Printf("WebSocket server being served at: %v/flutter\n", config.Required.WsHost)

	// launch SPA server on its own GoRoutine
	go func() {
		logger.Fatal(srv.ListenAndServe())
	}()
	logger.Fatal(ws.ListenAndServe())
	return nil
}
