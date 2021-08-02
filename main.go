package main

import (
	"fmt"
	"net/http"
	"os"
	"strconv"

	"github.com/go-redis/redis"
	"github.com/gorilla/mux"
	"github.com/sirupsen/logrus"
)

const (
	appPort         int    = 8080
	redisPort       int    = 6379
	redisServerName string = "redis-server"
	visitsDBKey     string = "visits"
)

var rdb *redis.Client

func main() {
	r := mux.NewRouter().SkipClean(true)
	r.HandleFunc("/", visits).Methods(http.MethodGet, http.MethodOptions)
	r.HandleFunc("/crash", crashContainerTest).Methods(http.MethodGet, http.MethodOptions)

	logrus.Info("Stating http server.")
	if err := http.ListenAndServe(fmt.Sprintf(":%d", appPort), r); err != nil {
		logrus.WithError(err).Error("Exitin http server.")
	}
}

func visits(writer http.ResponseWriter, _ *http.Request) {
	rdbRes := rdb.Get(visitsDBKey)
	if rdbRes.Err() != nil {
		errorResponse(writer, http.StatusInternalServerError, "Failed to retrieve data from the database.")

		return
	}

	val, err := rdbRes.Result()
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, "Failed to read data from the database.")

		return
	}

	visits, err := visitsIncrementor(val)
	if err != nil {
		errorResponse(writer, http.StatusInternalServerError, "Failed to convert visits as a number.")

		return
	}

	successReponse(writer, http.StatusOK, fmt.Sprintf("Number of visits: %s", visits))
}

func crashContainerTest(writer http.ResponseWriter, _ *http.Request) {
	os.Exit(0)
}

func visitsIncrementor(visits string) (string, error) {
	var setRes *redis.StatusCmd

	num, err := strconv.Atoi(visits)
	if err != nil {
		return "", err
	}

	num++
	if setRes = rdb.Set(visitsDBKey, num, 0); setRes.Err() != nil {
		return "", setRes.Err()
	}

	return strconv.Itoa(num), nil
}

func errorResponse(writer http.ResponseWriter, resposeStatus int, responseMsg string) {
	writer.WriteHeader(resposeStatus)
	if _, err := writer.Write([]byte(responseMsg)); err != nil {
		logrus.WithError(err).Error("Failed write error response.")
	}
}

func successReponse(writer http.ResponseWriter, resposeStatus int, responseMsg string) {
	writer.WriteHeader(resposeStatus)
	if _, err := writer.Write([]byte(responseMsg)); err != nil {
		logrus.WithError(err).Error("Failed write success response.")
	}
}

func init() {
	var res *redis.StatusCmd

	rdb = redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%d", redisServerName, redisPort),
	})

	if res = rdb.Set(visitsDBKey, 0, 0); res.Err() != nil {
		logrus.WithError(res.Err()).Error("Failed to establish database connection.")
		os.Exit(1)
	}

	logrus.Info("Database is up and running.")
}
