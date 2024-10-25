package types

import "time"


type CartStore interface {
	AddToCart(cart *Cart) error
	DeleteFromCart(id int) error
}


type Cart struct {
	ID        	int `json:"id"`
	UserID    	int `json:"user_id"`
	CourseID  	int `json:"course_id"`
	CreatedAt 	time.Time `json:"created_at"`
	ModifiedAt 	time.Time `json:"modified_at"`
}