package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"net/http"

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

func (h *Handlers) HandleCount(c echo.Context) error {
	switch c.Request().Method {
	case http.MethodGet:
		counter, err := h.dbProvider.GetCounter()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		err = h.dbProvider.UpdateCounter(1)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}

		return c.String(http.StatusOK, fmt.Sprintf("%d", counter+1)) // Увеличиваем на 1 для ответа

	case http.MethodPost:
		var requestBody struct {
			Count int `json:"count"`
		}

		if err := c.Bind(&requestBody); err != nil {
			return c.JSON(http.StatusBadRequest, map[string]string{"error": "это не число"})
		}

		err := h.dbProvider.UpdateCounter(requestBody.Count)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"message": "Success"})

	default:
		return c.JSON(http.StatusMethodNotAllowed, map[string]string{"error": "Неизвестный метод"})
	}
}

func main() {
	address := flag.String("address", "127.0.0.1:3333", "адрес для запуска сервера")
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

	e.GET("/count", h.HandleCount)
	e.POST("/count", h.HandleCount)

	fmt.Println("Сервер запущен на порту :3333")
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
