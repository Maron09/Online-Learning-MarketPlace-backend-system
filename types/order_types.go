package types

import (
	"database/sql"
	"time"

)




type OrderStore interface {
	BeginTransaction() (*sql.Tx, error)
	CreateOrder(tx *sql.Tx, order *Order) (int, error)
	CreateOrderItem(tx *sql.Tx, item *OrderItem) error
	GetCoursePrice(courseID int) (float64, error)
	// GetOrdersByUserID(userID int) ([]Order, error)
	// GetOrderItemsByOrderID(orderID int) ([]OrderItem, error)
}

type Order struct {
	ID          int       `json:"id"`
	UserID      int       `json:"user_id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	Email       string    `json:"email"`
	Country     string    `json:"country"`
	Total       float64   `json:"total"`
	OrderNumber string    `json:"order_number"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt  time.Time `json:"modified_at"`
}


type OrderItem struct {
	OrderID  int     `json:"order_id"`
	CourseID int     `json:"course_id"`
	Price    float64 `json:"price"`
}


type CreateOrderPayload struct {
	FirstName string       `json:"first_name"`
	LastName  string       `json:"last_name"`
	Email     string       `json:"email"`
	Country   string       `json:"country"`
}