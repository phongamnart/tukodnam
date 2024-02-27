package main

import (
	"database/sql"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    _ "github.com/lib/pq"
)

var db *sql.DB

const (
    host     = "localhost"
    port     = 5432
    user     = "postgres"
    password = "tamarin17"
    dbname   = "tukodnam"
)

type Menu struct {
	Product 	string `json:"product"`
	Price		string `json:"price"`
	Quantity	string `json:"quantity"`
}

func connectToDB() error {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)
	var err error
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		return err
	}
	if err = db.Ping(); err != nil {
		return err
	}
	return nil
}

func handleGet (w http.ResponseWriter, r *http.Request) {
	rows , err := db.Query("SELECT * FROM menu")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer rows.Close()
	var menu []Menu
	for rows.Next() {
		var mn Menu
		if err := rows.Scan(&mn.Product, &mn.Price, &mn.Quantity); err != nil {
			log.Println(err)
			continue
		}
		menu = append(menu, mn)
	}
	if err := rows.Err(); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(menu)
}

func main() {
	if err := connectToDB(); err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	http.HandleFunc("/menu", handleGet)

	log.Fatal(http.ListenAndServe(":9000", nil))
}