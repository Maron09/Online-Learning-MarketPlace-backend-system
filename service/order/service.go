package order

import (
	"fmt"

	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

func (h *Handler) CreateOrder(userID int, Items []types.Cart, firstName, lastName, email, country string) (int, float64, error) {
	courseMap := make(map[int]float64)

	for _, item := range Items {
		price, err := h.order.GetCoursePrice(item.CourseID)
		if err!= nil {
            return 0, 0, fmt.Errorf("could not fetch course price: %v", err)
        }
		courseMap[item.CourseID] = price 
	}

	totalPrice := calculateTotalPrice(Items, courseMap)

	tx, err := h.order.BeginTransaction()
	if err!= nil {
        return 0, 0, fmt.Errorf("could not start transaction: %v", err)
    }
	defer tx.Rollback()

	order_number := utils.GenerateOrderNumber()

	order := types.Order{
		UserID: userID,
		FirstName: firstName,
		LastName: lastName,
        Email: email,
        Country: country,
        Total: totalPrice,
		OrderNumber: order_number,
		Status: "pending",
	}

	orderID, err := h.order.CreateOrder(tx, &order)
	if err!= nil {
        return 0, 0, fmt.Errorf("could not create order: %v", err)
    }

	for _, item := range Items {
		err = h.order.CreateOrderItem(tx, &types.OrderItem{
			OrderID: orderID,
			CourseID: item.CourseID,
			Price: courseMap[item.CourseID],
		})
		if err!= nil {
            return 0, 0, fmt.Errorf("could not create order item: %v", err)
        }
	}

	if err := tx.Commit(); err != nil {
		return 0, 0, fmt.Errorf("could not commit transaction: %v", err)
	}
	return orderID, totalPrice, nil
}







func calculateTotalPrice(items []types.Cart, courseMap map[int]float64) float64 {
    total := 0.0
    for _, item := range items {
        total += courseMap[item.CourseID]
    }
    return total
}


func (h *Handler) EnrollStudentAfterPayment(userID int, courseIDs []int) error {
	for _, courseID := range courseIDs {
		enrollment := &types.Enrollment{
			Student: userID,
			CourseID: courseID,
		}
		err := h.order.CreateEnrollment(enrollment)
		if err!= nil {
            return fmt.Errorf("could not enroll student: %v", err)
        }
	}
	return nil
}