package order

import (
	"fmt"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	order types.OrderStore
	store types.UserStore
	cart types.CartStore
}


func NewHandler(order types.OrderStore, store types.UserStore, cart types.CartStore) *Handler {
    return &Handler{
		order: order,
	    store: store,
		cart: cart,
    }
}



func (h *Handler) OrderRoutes(router *mux.Router) {

	usersOnly := []types.UserRole{types.ADMIN, types.STUDENT}

	router.HandleFunc("/checkout", auth.WithJWTAuth(h.createOrderHandler, h.store, usersOnly)).Methods(http.MethodPost)
}


func (h *Handler) createOrderHandler(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetStudentIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
		return
	}

	var payload types.CreateOrderPayload
	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Validate payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	// Fetch user's cart items
	items, err := h.cart.GetCartItemsByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not retrieve cart items: %v", err))
		return
	}

	// Check if there are items in the cart
	if len(items) == 0 {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("cart is empty, add items to the cart before creating an order"))
		return
	}

	// Call CreateOrder with fetched cart items
	orderID, totalPrice, err := h.CreateOrder(userID, items, payload.FirstName, payload.LastName, payload.Email, payload.Country)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, err)
		return
	}

	// Respond with order details
	response := map[string]interface{}{
		"message":     "Order created successfully",
		"order_id":    orderID,
		"total_price": totalPrice,
		"status":      "pending",
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}
