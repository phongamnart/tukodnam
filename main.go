package main

import (
    "encoding/json"
    "log"
    "net/http"

    "gorm.io/driver/postgres"
    "gorm.io/gorm"
    "github.com/gorilla/handlers"
)

type Menu struct {
    ID       uint   `gorm:"primaryKey"`
    Product  string
    Price    float64
    Quantity int
}

var db *gorm.DB


func main() {
    var err error
    dsn := "host=localhost user=postgres password=tamarin17 dbname=tukodnam port=5432 sslmode=disable"
    db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
    if err != nil {
        log.Fatal(err)
    }
    db.AutoMigrate(&Menu{})

    http.HandleFunc("/menus", getMenus)

    allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
    allowedOrigins := handlers.AllowedOrigins([]string{"*"})
    allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
    cors := handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)

    log.Println("Server started at :9000")
    log.Fatal(http.ListenAndServe(":9000", cors(http.DefaultServeMux)))
}

func getMenus(w http.ResponseWriter, r *http.Request) {
    var menus []Menu
    db.Find(&menus)
    json.NewEncoder(w).Encode(menus)
}
