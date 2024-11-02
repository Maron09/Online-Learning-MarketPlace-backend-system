package cart

import (
	"database/sql"
	"fmt"

	"github.com/sikozonpc/ecom/types"
)

type Store struct {
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}


func (s *Store) CheckIfCourseInCart(userID, courseID int) (bool, error) {
	query := `
	SELECT EXISTS(SELECT 1 FROM cart WHERE user_id = $1 AND course_id = $2)
	`
	var exists bool
	err := s.db.QueryRow(query, userID, courseID).Scan(&exists)
	return exists, err
}




func (s *Store) AddToCart(cart *types.Cart) error {
	query := `
    INSERT INTO cart (user_id, course_id, created_at, modified_at)
    VALUES ($1, $2, $3, $4)
    `
    _, err := s.db.Exec(query, cart.UserID, cart.CourseID, cart.CreatedAt, cart.ModifiedAt)
    return err
}


func (s *Store) DeleteFromCart(cartID, userID int) error {
	query := `
    DELETE FROM cart WHERE id = $1 AND user_id = $2
    `
    _, err := s.db.Exec(query, cartID, userID)
    return err
}


func (s *Store) GetCartItemsByUserID(userID int) ([]types.Cart, error) {
	query := `
	SELECT c.id, c.user_id, c.course_id, courses.name, c.created_at, c.modified_at
	FROM cart AS c
	JOIN courses ON c.course_id = courses.id
	WHERE c.user_id = $1
	`
	rows, err := s.db.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch cart items: %v", err)
	}
	defer rows.Close()

	var cartItems []types.Cart
	for rows.Next() {
		var item types.Cart
        err := rows.Scan(
            &item.ID,
            &item.UserID,
            &item.CourseID,
			&item.CourseName,
            &item.CreatedAt,
            &item.ModifiedAt,
        )
        if err!= nil {
            return nil, fmt.Errorf("could not scan cart item row: %v", err)
        }
        cartItems = append(cartItems, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}
	
	return cartItems, nil
}

func (o *Store) DeleteCartItems(userID int) error {
    query := "DELETE FROM cart WHERE user_id = $1"
    _, err := o.db.Exec(query, userID)
    if err != nil {
        return fmt.Errorf("could not delete cart items: %w", err)
    }
    return nil
}