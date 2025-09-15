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

// Message struct
type Message struct {
    ID       int    `json:"id"`
    Sender   string `json:"sender"`
    Receiver string `json:"receiver"`
    Content  string `json:"content"`
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
	http.HandleFunc("/send-message", sendMessageHandler)
	http.HandleFunc("/get-messages", getMessagesHandler)
	http.HandleFunc("/users", getUsersHandler)

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

// Add this handler
func getUsersHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    rows, err := db.Query("SELECT id, full_name, user_name FROM users")
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    type UserResponse struct {
        ID       int    `json:"id"`
        FullName string `json:"fullName"`
        UserName string `json:"userName"`
    }

    var users []UserResponse
    for rows.Next() {
        var u UserResponse
        if err := rows.Scan(&u.ID, &u.FullName, &u.UserName); err != nil {
            http.Error(w, "DB scan error", http.StatusInternalServerError)
            return
        }
        users = append(users, u)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(users)
}


// Send a message
func sendMessageHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    var msg Message
    if err := json.NewDecoder(r.Body).Decode(&msg); err != nil {
        http.Error(w, "Invalid JSON", http.StatusBadRequest)
        return
    }

    _, err := db.Exec("INSERT INTO messages (sender, receiver, content) VALUES (?, ?, ?)",
        msg.Sender, msg.Receiver, msg.Content)
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }

    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{"message": "Message sent"})
}

// Get messages between current user and receiver
func getMessagesHandler(w http.ResponseWriter, r *http.Request) {
    if r.Method != http.MethodGet {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }

    sender := r.URL.Query().Get("sender")
    receiver := r.URL.Query().Get("receiver")
    if sender == "" || receiver == "" {
        http.Error(w, "sender and receiver required", http.StatusBadRequest)
        return
    }

    rows, err := db.Query(`SELECT sender, receiver, content 
                           FROM messages 
                           WHERE (sender=? AND receiver=?) OR (sender=? AND receiver=?)
                           ORDER BY id ASC`,
        sender, receiver, receiver, sender)
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var msgs []Message
    for rows.Next() {
        var m Message
        if err := rows.Scan(&m.Sender, &m.Receiver, &m.Content); err != nil {
            http.Error(w, "DB scan error", http.StatusInternalServerError)
            return
        }
        msgs = append(msgs, m)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(msgs)
}


