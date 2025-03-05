package main

import (
	authorization "go_final_project/http-server/auth"
	"go_final_project/http-server/handlers"
	"go_final_project/storage/sqlite"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// Используем библиотеку godotenv для загрузки переменных окружения
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %w", err)
	}

	port := os.Getenv("TODO_PORT")
	DBFile := os.Getenv("TODO_DBFILE")

	// Используем константу для количества выводимых задач
	const NumberOfOutPutTasks = 10

	storage, err := sqlite.New(DBFile)
	if err != nil {
		log.Fatalf("Error initializing storage: %w", err)
	}
	defer storage.Close()

	server := chi.NewRouter()
	server.Handle("/*", http.FileServer(http.Dir("web")))
	server.HandleFunc("/api/nextdate", handlers.ApiNextDate)
	server.Post("/api/task", authorization.CheckToken(handlers.PostTask(storage)))
	server.Get("/api/task", authorization.CheckToken(handlers.GetTask(storage)))
	server.Put("/api/task", authorization.CheckToken(handlers.CorrectTask(storage)))
	server.Post("/api/task/done", authorization.CheckToken(handlers.DoneTask(storage)))
	server.Delete("/api/task", authorization.CheckToken(handlers.DeleteTask(storage)))

	NumberOfOutPutTasksStr := strconv.Itoa(NumberOfOutPutTasks)
	server.Get("/api/tasks", authorization.CheckToken(handlers.GetTasks(storage, NumberOfOutPutTasksStr)))
	server.Post("/api/signin", authorization.Authorization)

	log.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		log.Fatalf("server startup error: %w", err)
	}
}
