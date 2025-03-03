package main

import (
	authorization "go_final_project/http-server/auth"
	"go_final_project/http-server/handlers"
	"go_final_project/storage/sqlite"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	_ "github.com/mattn/go-sqlite3"
)

func main() {
	// use the godotenv library to load the project root environment variables
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalf("Error loading .env file: %w", err)
	}
	var port string = os.Getenv("TODO_PORT")
	var DBFile = os.Getenv("TODO_DBFILE")
	var NumberOfOuptuTasks = os.Getenv("NUMNER_OF_OUTPUT_TASKS")

	storage, err := sqlite.New(DBFile)
	defer storage.Close()

	server := chi.NewRouter()
	// Basic Handler Description. Middleware authorization.CheckToken is used to check the JWT token
	// The token is generated during user authorization.
	server.Handle("/*", http.FileServer(http.Dir("web")))
	server.HandleFunc("/api/nextdate", handlers.ApiNextDate)
	server.Post("/api/task", authorization.CheckToken(handlers.PostTask(storage)))
	server.Get("/api/task", authorization.CheckToken(handlers.GetTask(storage)))
	server.Put("/api/task", authorization.CheckToken(handlers.CorrectTask(storage)))
	server.Post("/api/task/done", authorization.CheckToken(handlers.DoneTask(storage)))
	server.Delete("/api/task", authorization.CheckToken(handlers.DeleteTask(storage)))
	server.Get("/api/tasks", authorization.CheckToken(handlers.GetTasks(storage, NumberOfOuptuTasks)))
	server.Post("/api/signin", authorization.Authorization)

	log.Printf("Starting server on :%s\n", port)
	err = http.ListenAndServe(":"+port, server)
	if err != nil {
		log.Fatalf("server startup error: %w", err)
	}

}
