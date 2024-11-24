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

// Обработчики HTTP-запросов
func (h *Handlers) GetUser(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name parameter is required"})
	}

	user, err := h.dbProvider.SelectUser(name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.String(http.StatusOK, "Hello, "+user+"!")
}

func (h *Handlers) PostUser(c echo.Context) error {
	var input struct {
		Name string `json:"name"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	err := h.dbProvider.InsertUser(input.Name)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Запись добавлена!"})
}

func (dp *DatabaseProvider) SelectUser(name string) (string, error) {
	var user string

	row := dp.db.QueryRow("SELECT name FROM mytable WHERE name = $1", name)
	err := row.Scan(&user)
	if err != nil {
		return "", err
	}

	return user, nil
}

func (dp *DatabaseProvider) InsertUser(name string) error {
	_, err := dp.db.Exec("INSERT INTO mytable (name) VALUES ($1)", name)
	return err
}

func main() {
	address := flag.String("address", "127.0.0.1:8080", "адрес для запуска сервера")
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

	e.GET("/api/user", h.GetUser)
	e.POST("/api/user/create", h.PostUser)

	fmt.Printf("Сервер запущен на %s\n", *address)
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
