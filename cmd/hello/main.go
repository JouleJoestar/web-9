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
func (h *Handlers) GetHello(c echo.Context) error {
	msg, err := h.dbProvider.SelectHello()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.String(http.StatusOK, msg)
}

func (h *Handlers) PostHello(c echo.Context) error {
	var input struct {
		Msg string `json:"msg"`
	}

	if err := c.Bind(&input); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Неверный формат запроса"})
	}

	err := h.dbProvider.InsertHello(input.Msg)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}

	return c.JSON(http.StatusCreated, map[string]string{"message": "Сообщение успешно добавлено"})
}

// Методы для работы с базой данных
func (dp *DatabaseProvider) SelectHello() (string, error) {
	var msg string

	// Получаем одно сообщение из таблицы hello, отсортированной в случайном порядке
	row := dp.db.QueryRow("SELECT message FROM hello ORDER BY RANDOM() LIMIT 1")
	err := row.Scan(&msg)
	if err != nil {
		return "", err
	}

	return msg, nil
}

func (dp *DatabaseProvider) InsertHello(msg string) error {
	_, err := dp.db.Exec("INSERT INTO hello (message) VALUES ($1)", msg)
	return err
}

func main() {
	// Считываем аргументы командной строки
	address := flag.String("address", "127.0.0.1:8081", "адрес для запуска сервера")
	flag.Parse()

	// Формирование строки подключения для postgres
	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	// Создание соединения с сервером postgres
	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Создаем провайдер для БД с набором методов
	dp := DatabaseProvider{db: db}
	// Создаем экземпляр структуры с набором обработчиков
	h := Handlers{dbProvider: dp}

	// Создаем экземпляр Echo
	e := echo.New()

	// Регистрируем обработчики
	e.GET("/get", h.GetHello)
	e.POST("/post", h.PostHello)

	// Запускаем веб-сервер на указанном адресе
	fmt.Printf("Сервер запущен на %s\n", *address)
	if err := e.Start(*address); err != nil {
		log.Fatal(err)
	}
}
