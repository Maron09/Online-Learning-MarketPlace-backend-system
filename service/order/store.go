package order

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/sikozonpc/ecom/types"
)

type Store struct {
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}



func (s *Store) BeginTransaction() (*sql.Tx, error) {
	return s.db.Begin()
}


func (s *Store) CreateOrder(tx *sql.Tx, order *types.Order) (int, error) {
	var orderID int
	query := `
		INSERT INTO orders (user_id, first_name, last_name, email, country, total, order_number, status, created_at, modified_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10) RETURNING id`
	
	err := tx.QueryRow(
		query,
        order.UserID,
        order.FirstName,
        order.LastName,
        order.Email,
        order.Country,
        order.Total,
        order.OrderNumber,
        order.Status,
        time.Now(), 
        time.Now(),
	).Scan(&orderID)
	if err != nil {
        return 0, fmt.Errorf("could not create order: %v", err)
    }

	return orderID, nil
}



func (s *Store) CreateOrderItem(tx *sql.Tx, item *types.OrderItem) error {
	query := `
		INSERT INTO order_items (order_id, course_id, price) 
		VALUES ($1, $2, $3)`
	_, err := tx.Exec(query, item.OrderID, item.CourseID, item.Price)
	if err != nil {
		return fmt.Errorf("could not add item to order: %v", err)
	}
	return nil
}



func (s *Store) GetCoursePrice(courseID int) (float64, error) {
	var price float64
	query := `SELECT price FROM courses WHERE id = $1`

	err := s.db.QueryRow(query, courseID).Scan(&price)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("course not found")
		}
		return 0, fmt.Errorf("error retrieving course price: %v", err)
	}

	return price, nil
}