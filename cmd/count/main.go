package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	_ "github.com/lib/pq"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "mydb"
)

type Handlers struct {
	dbProvider DatabaseProvider
}

type DatabaseProvider struct {
	db *sql.DB
}

func (dp *DatabaseProvider) GetCounter() (int, error) {
	var counter int
	row := dp.db.QueryRow("SELECT value FROM counter_table LIMIT 1")
	err := row.Scan(&counter)
	if err != nil {
		return 0, err
	}
	return counter, nil
}

func (dp *DatabaseProvider) UpdateCounter(value int) error {
	_, err := dp.db.Exec("UPDATE counter_table SET value = value + $1", value)
	return err
}

func (h *Handlers) getCount(c echo.Context) error {
	counter, err := h.dbProvider.GetCounter()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.String(http.StatusOK, fmt.Sprintf("%d", counter))
}

func (h *Handlers) postCount(c echo.Context) error {
	countParam := c.QueryParam("count")
	count, err := strconv.Atoi(countParam)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "это не число"})
	}

	err = h.dbProvider.UpdateCounter(count)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return c.JSON(http.StatusOK, map[string]string{"message": "Success"})
}

func main() {
	address := flag.String("address", "127.0.0.1:3332", "адрес для запуска сервера")
	flag.Parse()

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	dp := DatabaseProvider{db: db}
	h := Handlers{dbProvider: dp}

	e := echo.New()

	e.Logger.SetLevel(2)

	e.GET("/count", h.getCount)
	e.POST("/count", h.postCount)

	fmt.Println("Сервер запущен на порту :3332")
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
