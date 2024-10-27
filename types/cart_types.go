package types

import "time"


type CartStore interface {
	GetCart(userID int) ([]Cart, error)
	AddToCart(cart *Cart) error
	DeleteFromCart(cartID, userID int) error
	CheckIfCourseInCart(userID, courseID int) (bool, error)
	GetCartItemsByUserID(userID int) ([]Cart, error)
}


type Cart struct {
	ID        	int `json:"id"`
	UserID    	int `json:"user_id"`
	CourseID  	int `json:"course_id"`
	CreatedAt 	time.Time `json:"created_at"`
	ModifiedAt 	time.Time `json:"modified_at"`
}


type AddToCartPayload struct {
	CourseID int `json:"course_id"`
}