package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"strings"

	_ "github.com/lib/pq"
)

type Todo struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var db *sql.DB

func main() {
	connStr := "postgres://postgres:ArSeN001@localhost:5432/tododb?sslmode=disable"
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	db.Exec(`CREATE TABLE IF NOT EXISTS todos (
		id SERIAL PRIMARY KEY,
		title TEXT NOT NULL,
		done BOOLEAN NOT NULL DEFAULT false
	)`)

	http.HandleFunc("/todos", todosHandler)
	http.HandleFunc("/todos/", todoHandler)

	log.Println("Server running at http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func todosHandler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		rows, _ := db.Query("SELECT id, title, done FROM todos")
		defer rows.Close()
		var todos []Todo
		for rows.Next() {
			var t Todo
			rows.Scan(&t.ID, &t.Title, &t.Done)
			todos = append(todos, t)
		}
		json.NewEncoder(w).Encode(todos)

	case http.MethodPost:
		var t Todo
		json.NewDecoder(r.Body).Decode(&t)
		db.QueryRow("INSERT INTO todos (title, done) VALUES ($1, $2) RETURNING id", t.Title, t.Done).Scan(&t.ID)
		json.NewEncoder(w).Encode(t)
	}
}

func todoHandler(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(strings.TrimPrefix(r.URL.Path, "/todos/"))

	switch r.Method {
	case http.MethodPut:
		var t Todo
		json.NewDecoder(r.Body).Decode(&t)
		db.Exec("UPDATE todos SET title=$1, done=$2 WHERE id=$3", t.Title, t.Done, id)
		w.WriteHeader(http.StatusOK)

	case http.MethodDelete:
		db.Exec("DELETE FROM todos WHERE id=$1", id)
		w.WriteHeader(http.StatusNoContent)
	}
}
