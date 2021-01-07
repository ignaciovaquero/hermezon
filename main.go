package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/bamzi/jobrunner"
	"github.com/igvaquero18/hermezon/boltdb"
	"github.com/igvaquero18/hermezon/scraper"
	"github.com/igvaquero18/hermezon/telegram"
	"github.com/igvaquero18/hermezon/twilio"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"go.uber.org/zap"
)

const (
	sidEnv                  = "HERMEZON_TWILIO_ACCOUNT_SID"
	tokenEnv                = "HERMEZON_TWILIO_ACCOUNT_TOKEN"
	expectedStatusCodeEnv   = "HERMEZON_EXPECTED_STATUS_CODE"
	telegramTokenEnv        = "HERMEZON_TELEGRAM_TOKEN"
	maxRetriesEnv           = "HERMEZON_MAX_RETRIES"
	retrySecondsEnv         = "HERMEZON_RETRY_SECONDS"
	verboseEnv              = "HERMEZON_VERBOSE"
	portEnv                 = "HERMEZON_LISTEN_PORT"
	priceScheduleEnv        = "HERMEZON_PRICE_SCHEDULE_FREQUENCY"
	availabilityScheduleEnv = "HERMEZON_AVAILABILITY_SCHEDULE_FREQUENCY"
	jwtSecretEnv            = "HERMEZON_JWT_SECRET"
	databaseFilePathEnv     = "HERMEZON_DB_FILE_PATH"
	twilioPhoneEnv          = "HERMEZON_TWILIO_PHONE"
	apiVersion              = "/v1"
)

var (
	twilioSID             = os.Getenv(sidEnv)
	twilioToken           = os.Getenv(tokenEnv)
	twilioPhone           = os.Getenv(twilioPhoneEnv)
	telegramToken         = os.Getenv(telegramTokenEnv)
	v                     = getOrElse(verboseEnv, "false")
	jwtSecret             = getOrElse(jwtSecretEnv, "secret")
	listenPort            = getOrElse(portEnv, "8080")
	databaseFilePath      = getOrElse(databaseFilePathEnv, "hermezon.db")
	priceFrequency        = getOrElse(priceScheduleEnv, "1h")
	availabilityFrequency = getOrElse(availabilityScheduleEnv, "1m")
	maxRetries            int8
	retrySeconds          int8
	expectedStatusCode    int
	sugar                 *zap.SugaredLogger
	verbose               bool
	messenger             *Messenger
	e                     *echo.Echo
	p                     *prometheus.Prometheus
	db                    KeyValueStorage
	messagingClient       Messenger
)

func getOrElse(envVar, defaultValue string) string {
	value := os.Getenv(envVar)
	if value == "" {
		return defaultValue
	}
	return value
}

func init() {
	var err error

	verbose, err = strconv.ParseBool(v)
	if err != nil {
		log.Printf("Incorrect value for %s: %s. Defaulting to false", verboseEnv, v)
		verbose = false
	}

	var zl *zap.Logger
	cfg := zap.Config{
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    zap.NewProductionEncoderConfig(),
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}
	if verbose {
		cfg.Level = zap.NewAtomicLevelAt(zap.DebugLevel)
	} else {
		cfg.Level = zap.NewAtomicLevelAt(zap.InfoLevel)
	}
	zl, err = cfg.Build()

	if err != nil {
		log.Fatalf("Error when initializing logger: %v", err)
	}

	sugar = zl.Sugar()
	sugar.Debug("Logger initialization successful")

	expectedStatusCode = http.StatusOK
	if statusCode := os.Getenv(expectedStatusCodeEnv); statusCode != "" {
		expectedStatusCode, err = strconv.Atoi(statusCode)
		if err != nil {
			sugar.Errorw("error when setting expected status code. Defaulting to 200", "msg", err.Error(), "status_code", statusCode)
			expectedStatusCode = http.StatusOK
		}
	}

	maxRetries = scraper.DefaultMaxRetries
	if retries := os.Getenv(maxRetriesEnv); retries != "" {
		ret, err := strconv.Atoi(retries)
		if err != nil {
			sugar.Errorw("error when setting max retries. Taking default value...", "msg", err.Error(), "retries", retries)
			maxRetries = scraper.DefaultMaxRetries
		} else {
			maxRetries = int8(ret)
		}
	}

	retrySeconds = scraper.DefaultRetrySeconds
	if seconds := os.Getenv(retrySecondsEnv); seconds != "" {
		ret, err := strconv.Atoi(seconds)
		if err != nil {
			sugar.Errorw("error when setting retry seconds. Taking default value...", "msg", err.Error(), "seconds", seconds)
			retrySeconds = scraper.DefaultRetrySeconds
		} else {
			retrySeconds = int8(ret)
		}
	}

	if (twilioSID == "" || twilioToken == "" || twilioPhone == "") && telegramToken == "" {
		sugar.Fatal("at least one of twilio or telegram configurations is required")
	}

	// Creating Messaging client
	if twilioSID != "" && twilioToken != "" && twilioPhone != "" {
		messagingClient = twilio.NewClient(twilioSID, twilioToken, twilio.SetLogger(sugar))
	} else {
		messagingClient, err = telegram.NewClient(telegramToken, sugar)
		if err != nil {
			sugar.Fatalw("error when creating the telegram client", "msg", err.Error())
		}
	}

	// Setting routes in echo router and securing them with JWT
	e = echo.New()
	e.Use(middleware.Recover())
	r := e.Group(fmt.Sprintf("%s/actions", apiVersion))
	r.Use(middleware.JWT([]byte(jwtSecret)))
	r.POST("", postActions)

	// Enabling Prometheus metrics
	p = prometheus.NewPrometheus("echo", nil)
	p.Use(e)

}

func main() {
	var err error

	// Setting up the database
	db, err = boltdb.NewClient(databaseFilePath, sugar)
	if err != nil {
		sugar.Fatalw("error when opening the database", "msg", err.Error())
	}
	defer db.Close()

	jobrunner.Start()
	if err = jobrunner.Schedule(fmt.Sprintf("@every %s", availabilityFrequency), availability{}); err != nil {
		sugar.Fatalw("error when scheduling availability jobs", "msg", err.Error())
	}
	if err = jobrunner.Schedule(fmt.Sprintf("@every %s", priceFrequency), price{}); err != nil {
		sugar.Fatalw("error when scheduling price down jobs", "msg", err.Error())
	}
	sugar.Fatal(e.Start(fmt.Sprintf(":%s", listenPort)))
}
