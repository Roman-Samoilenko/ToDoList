package main

import (
	"database/sql"
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

type Task struct {
	ID   int    `json:"id_todos"`
	Item string `json:"item"`
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	_, err = db.Exec("CREATE TABLE IF NOT EXISTS todos (id_todos SERIAL PRIMARY KEY, item TEXT)")
	if err != nil {
		log.Fatal(err)
	}

	router := mux.NewRouter()
	router.HandleFunc("/todos", getTodos(db)).Methods("GET")
	router.HandleFunc("/todos/{id}", getTodo(db)).Methods("GET")
	router.HandleFunc("/todos", createTodo(db)).Methods("POST")
	router.HandleFunc("/todos/{id}", updateTodo(db)).Methods("PUT")
	router.HandleFunc("/todos/{id}", deleteTodo(db)).Methods("DELETE")

	log.Println("Server starting on :8002")
	log.Fatal(http.ListenAndServe(":8002", middleware(router)))
}

func middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

func getTodos(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM todos")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		var todos = []Task{}
		for rows.Next() {
			var t Task
			err := rows.Scan(&t.ID, &t.Item)
			if err != nil {
				log.Fatal(err)
			}
			todos = append(todos, t)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(todos)

	}

}

func getTodo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		var t Task
		err := db.QueryRow("SELECT * FROM todos WHERE id_todos = $1", id).Scan(&t.ID, &t.Item)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(t)
	}
}

func createTodo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		err := db.QueryRow("INSERT INTO todos (item) VALUES ($1) RETURNING id_todos", t.Item).Scan(&t.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(t)
	}
}

func updateTodo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		params := mux.Vars(r)
		id := params["id"]

		_, err := db.Exec("UPDATE todos SET item = $1 WHERE id_todos = $2", t.Item, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(t)
	}
}

func deleteTodo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		params := mux.Vars(r)
		id := params["id"]

		var t Task
		err := db.QueryRow("SELECT * FROM todos WHERE id_todos = $1", id).Scan(&t.ID, &t.Item)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM todos WHERE id_todos = $1", id)
			if err != nil {
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("Todo deleted")
		}
	}

}
