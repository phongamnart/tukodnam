package main

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/gorilla/handlers"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Menu struct {
	ID       uint `gorm:"primaryKey"`
	Product  string
	Price    float64
	Quantity int
}

type Cart struct {
	ID       uint `gorm:"primaryKey"`
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
    db.AutoMigrate(&Cart{})

	http.HandleFunc("/menus", getMenus)
	http.HandleFunc("/add-to-cart", addToCartHandler)
	http.HandleFunc("/remove-from-cart", removeFromCartHandler)
	http.HandleFunc("/get-cart", getCartHandler)

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

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	var product Menu
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var existingProduct Cart
	if err := db.Where("product = ?", product.Product).First(&existingProduct).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			cartProduct := Cart{
				Product:  product.Product,
				Price:    product.Price,
				Quantity: 1,
			}
			db.Create(&cartProduct)
			w.WriteHeader(http.StatusCreated)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	existingProduct.Quantity++
	db.Save(&existingProduct)
	w.WriteHeader(http.StatusOK)
}

func removeFromCartHandler(w http.ResponseWriter, r *http.Request) {
	var product Menu
	if err := json.NewDecoder(r.Body).Decode(&product); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var cartProduct Cart
	if err := db.Where("product = ?", product.Product).First(&cartProduct).Error; err != nil {
		http.Error(w, "Product not found in cart", http.StatusNotFound)
		return
	}

	if cartProduct.Quantity > 1 {
		cartProduct.Quantity--
		db.Save(&cartProduct)
	} else {
		db.Delete(&cartProduct)
	}
	w.WriteHeader(http.StatusOK)
}

func getCartHandler(w http.ResponseWriter, r *http.Request) {
	var cartItems []Cart
	db.Find(&cartItems)
	json.NewEncoder(w).Encode(cartItems)
}
