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

type Order struct {
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
	db.AutoMigrate(&Order{})

	http.HandleFunc("/menus", getMenus)
	http.HandleFunc("/add-to-cart", addToCartHandler)
	http.HandleFunc("/remove-from-cart", removeFromCartHandler)
	http.HandleFunc("/get-cart", getCartHandler)
	http.HandleFunc("/clear-cart", clearCartHandler)
	http.HandleFunc("/place-order", placeOrderHandler)
	http.HandleFunc("/pay", payHandler)
	http.HandleFunc("/sold", getSoldHandler)

	allowedHeaders := handlers.AllowedHeaders([]string{"X-Requested-With", "Content-Type", "Authorization"})
	allowedOrigins := handlers.AllowedOrigins([]string{"*"})
	allowedMethods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	cors := handlers.CORS(allowedHeaders, allowedOrigins, allowedMethods)

	log.Println("Server started at :9000")
	log.Fatal(http.ListenAndServe(":9000", cors(http.DefaultServeMux)))
}

// get menu form db
func getMenus(w http.ResponseWriter, r *http.Request) {
	var menus []Menu
	//select * from table
	db.Find(&menus)
	json.NewEncoder(w).Encode(menus)
}

// add product to cart
func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	var product Menu
	//error read request body
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
	//insert into db
	db.Save(&existingProduct)
	w.WriteHeader(http.StatusOK)
}

// remove product to cart
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

// get cart to db
func getCartHandler(w http.ResponseWriter, r *http.Request) {
	var cartItems []Cart
	db.Find(&cartItems)
	json.NewEncoder(w).Encode(cartItems)
}

// clear cart
func clearCartHandler(w http.ResponseWriter, r *http.Request) {
	db.Exec("DELETE FROM carts")
	w.WriteHeader(http.StatusOK)
}

// place order
func placeOrderHandler(w http.ResponseWriter, r *http.Request) {
	var total float64
	var cartItems []Cart
	db.Find(&cartItems)

	// Calculate total amount
	for _, item := range cartItems {
		total += item.Price * float64(item.Quantity)
	}

	// Clear the cart
	db.Delete(&Cart{})

	summary := map[string]interface{}{
		"total": total,
		"items": cartItems,
	}

	json.NewEncoder(w).Encode(summary)
}

func payHandler(w http.ResponseWriter, r *http.Request) {
	var cartItems []Cart
	db.Find(&cartItems)

	for _, item := range cartItems {
		order := Order{
			Product:  item.Product,
			Price:    item.Price,
			Quantity: item.Quantity,
		}
		db.Create(&order)
	}

	db.Delete(&cartItems)

	for _, item := range cartItems {
		var menu Menu
		db.Where("product = ?", item.Product).First(&menu)
		menu.Quantity -= item.Quantity
		db.Save(&menu)
	}

	w.WriteHeader(http.StatusOK)
}

func getSoldHandler(w http.ResponseWriter, r *http.Request) {
	var order []Order
	db.Find(&order)
	json.NewEncoder(w).Encode(order)
}
