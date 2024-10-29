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



func (s *Store) GetLatestOrderByUserID(userID int) (*types.Order, error) {
	var order types.Order

    // Assuming `orders` is the table name and `created_at` determines the order creation time
    query := `
        SELECT id, user_id, first_name, last_name, email, country, total, order_number, status, created_at
        FROM orders
        WHERE user_id = $1
        ORDER BY created_at DESC
        LIMIT 1;
    `

	rows := s.db.QueryRow(query, userID)
	err := rows.Scan(
        &order.ID,
        &order.UserID,
        &order.FirstName,
        &order.LastName,
		&order.Email,
        &order.Country,
        &order.Total,
        &order.OrderNumber,
        &order.Status,
		&order.CreatedAt,
	)
	if err!= nil {
		if err == sql.ErrNoRows {
            return nil, fmt.Errorf("no orders found for user ID: %d", userID)
        }
        return nil, fmt.Errorf("error retrieving latest order: %v", err)
	}
	return &order, nil
}


func (s *Store) UpdateOrderStatus(orderID int, status string) error {
    query := `UPDATE orders SET status=$1 WHERE id=$2`
    result, err := s.db.Exec(query, status, orderID)
    if err != nil {
        return fmt.Errorf("error updating order status: %v", err)
    }

    rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("could not check affected rows: %v", err)
    }
    if rowsAffected == 0 {
        return fmt.Errorf("no order found with ID %d", orderID)
    }

    return nil
}



func (s *Store) GetOrderItemsByOrderID(orderID int) ([]types.OrderItem, error) {
	query := `
	SELECT id, order_id, course_id, price
	FROM order_items
	WHERE order_id = $1
	`

	rows, err := s.db.Query(query, orderID)
	if err!= nil {
        return nil, fmt.Errorf("could not get order items: %v", err)
    }
	defer rows.Close()
	
    var items []types.OrderItem
	for rows.Next() {
		var item types.OrderItem
        err := rows.Scan(
            &item.ID,
            &item.OrderID,
            &item.CourseID,
            &item.Price,
        )
        if err!= nil {
            return nil, err
        }
        items = append(items, item)
	}
	return items, nil
}


func (s *Store) CreateEnrollment(enrollment *types.Enrollment) error {
	query := `INSERT INTO enrollments (student_id, course_id, enrolled_at)
		VALUES ($1, $2, NOW())`

	_, err := s.db.Exec(query, enrollment.Student, enrollment.CourseID)
	if err!= nil {
        return fmt.Errorf("could not create enrollment: %v", err)
    }
	return nil
}
