package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"

	_ "github.com/go-sql-driver/mysql"
)

type User struct {
	FullName string `json:"fullName"`
	UserName string `json:"userName"`
	Password string `json:"password"`
}

var db *sql.DB

func main() {
	var err error
	dsn := "root:root@tcp(127.0.0.1:3306)/gochat"
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		panic(err)
	}

	err = db.Ping()
	if err != nil {
		panic(err)
	}
	fmt.Println("Connected to MySQL")

	// Routes
	http.HandleFunc("/create-account", createAccountHandler)
	http.HandleFunc("/login", loginHandler)

	fmt.Println("Server running at http://127.0.0.1:8080")
	http.ListenAndServe(":8080", nil)
}

func createAccountHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	_, err = db.Exec("INSERT INTO users (full_name, user_name, password) VALUES (?, ?, ?)",
		user.FullName, user.UserName, user.Password)
	if err != nil {
		http.Error(w, "Username already exists or DB error", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Account created successfully",
	})
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var creds User
	err := json.NewDecoder(r.Body).Decode(&creds)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	var storedPassword string
	err = db.QueryRow("SELECT password FROM users WHERE user_name = ?", creds.UserName).Scan(&storedPassword)
	if err == sql.ErrNoRows {
		http.Error(w, "Invalid username", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "DB error", http.StatusInternalServerError)
		return
	}

	if creds.Password != storedPassword {
		http.Error(w, "Invalid password", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Login successful",
	})
}

