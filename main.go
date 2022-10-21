package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"unicode"

	_ "github.com/lib/pq"
)

const (
	dbUser              = "tester"
	dbPassword          = "testdb"
	dbName              = "tester"
	dbSslMode           = "disable"
	emailMaxLenght      = 256
	emailRequiredSymbol = "@"
	passwordMaxLenght   = 256
	passwordMinLenght   = 8
	fullnameMinLenght   = 2
)

type User struct {
	Email    string `json:"email"`
	Fullname string `json:"fullname"` //according to the task in the request fullname is one "field", but in the future it is
	//maybe better to divide it into first name and last name, so as not to update the entire field if you need to change, for
	//example, only the first name
	Password string `json:"password"`
}

func (u userHandler) isEmailUnique(n User) bool {
	var emailIsUnique bool
	if err := u.db.QueryRow("SELECT Email FROM users WHERE Email=$1", n.Email).Scan(&emailIsUnique); err != nil {
		if err == sql.ErrNoRows {
			return true
		}
		log.Println("There's an error with the database", err)
	}
	return false
}

func isASCII(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] > unicode.MaxASCII {
			return false
		}
	}

	return true
}

func (u User) isValid() bool {
	if len(u.Email) > emailMaxLenght || !strings.Contains(u.Email, "@") {
		return false
	}

	if len(u.Password) > passwordMaxLenght || len(u.Password) < passwordMinLenght {
		return false
	}

	if !isASCII(u.Password) {
		return false
	}
	if len(u.Fullname) < fullnameMinLenght {
		return false
	}

	return true
}

type userHandler struct {
	db *sql.DB
}

func newUserHandler(db *sql.DB) userHandler {
	return userHandler{db: db}
}

func (u userHandler) handleCreateUser(writer http.ResponseWriter, request *http.Request) {
	if request.Method != "POST" {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad Request"))
		log.Println("Wrong method in request")
		return
	}
	var newUser User
	if err := json.NewDecoder(request.Body).Decode(&newUser); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad Request"))
		log.Println("There was an error decoding the request body into the struct")
		return
	}

	if !newUser.isValid() {
		writer.WriteHeader(http.StatusBadRequest)
		writer.Write([]byte("Bad Request"))
		log.Println("Email or password or fullname do not meet the requirements")
		return
	}

	if !u.isEmailUnique(newUser) {
		writer.WriteHeader(http.StatusConflict)
		writer.Write([]byte("Conflict"))
		log.Println("Email is not unique")
		return
	}

	const insertQuery = "INSERT INTO users(email, fullname,password) VALUES($1, $2, $3);"
	row := u.db.QueryRow(insertQuery, newUser.Email, newUser.Fullname, newUser.Password)
	if row.Err() != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write([]byte("Internal Server Error"))
		log.Println("Database error", row.Err())
		return
	}

	writer.Header().Set("Content-Type", "application/json")
	writer.WriteHeader(http.StatusOK)
}

func main() {
	dbinfo := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=%s", dbUser, dbPassword, dbName, dbSslMode)
	db, err := sql.Open("postgres", dbinfo)
	if err != nil {
		log.Fatalln("There's an error with the database", err)
	}

	uHandler := newUserHandler(db)

	http.HandleFunc("/users", uHandler.handleCreateUser)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalln("There's an error with the server,", err)
	}
}
