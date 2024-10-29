package order

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	// "os"

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



var clientID = os.Getenv("CLIENT_ID")
var clientSecret = os.Getenv("CLIENT_SECRET")

func NewHandler(order types.OrderStore, store types.UserStore, cart types.CartStore) *Handler {
    return &Handler{
		order: order,
	    store: store,
		cart: cart,
    }
}


func CreatePayPalClient() *http.Client {
    return &http.Client{}
}


func (h *Handler) OrderRoutes(router *mux.Router) {

	usersOnly := []types.UserRole{types.ADMIN, types.STUDENT}

	router.HandleFunc("/checkout", auth.WithJWTAuth(h.createOrderHandler, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/payments/create-paypal", auth.WithJWTAuth(h.CreatePayPalPayment, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/payments/paypal-success", auth.WithJWTAuth(h.HandlePayPalSuccess, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/payments/paypal-cancel", auth.WithJWTAuth(h.HandlePayPalCancel, h.store, usersOnly)).Methods(http.MethodGet)
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

func (h *Handler) CreatePayPalPayment(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetStudentIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
		return
	}

	order, err := h.order.GetLatestOrderByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not retrieve latest order: %v", err))
		return
	}

	// PayPal Payment Payload
	paymentPayload := map[string]interface{}{
		"intent": "sale",
		"payer": map[string]string{
			"payment_method": "paypal",
		},
		"transactions": []map[string]interface{}{
			{
				"amount": map[string]string{
					"total":    fmt.Sprintf("%.2f", order.Total),
					"currency": "USD",
				},
				"description": "Order payment",
			},
		},
		"redirect_urls": map[string]string{
			"return_url": "https://localhost:8000/api/v1/payments/paypal-success",
			"cancel_url": "https://localhost:8000/api/v1/payments/paypal-cancel",
		},
	}

	// Convert payload to JSON
	paymentData, err := json.Marshal(paymentPayload)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to marshal payment data: %v", err))
		return
	}

	client := CreatePayPalClient()
	req, err := http.NewRequest("POST", "https://api.sandbox.paypal.com/v1/payments/payment", bytes.NewBuffer(paymentData))
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error creating request: %v", err))
		return
	}

	// Set Headers
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := client.Do(req)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error executing request: %v", err))
		return
	}
	defer resp.Body.Close()

	// Read and decode the response
	var paymentResponse map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&paymentResponse); err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to decode response: %v", err))
		return
	}

	// Extract the approval URL
	var approvalURL string
	// Extract the approval URL
	links, ok := paymentResponse["links"].([]interface{})
	if !ok || len(links) == 0 {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("approval URL not found"))
		return
	}

	log.Printf("Payment response: %+v", paymentResponse)

	for _, link := range links {
		linkMap, ok := link.(map[string]interface{})
		if !ok {
			continue
		}
		if linkMap["rel"] == "approval_url" {
			approvalURL, _ = linkMap["href"].(string)
			break
		}
	}

	if approvalURL == "" {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("approval URL not found"))
		return
	}


	response := map[string]string{
		"approval_url": approvalURL,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}





func (h *Handler) HandlePayPalSuccess(writer http.ResponseWriter, request *http.Request) {
	paymentID := request.URL.Query().Get("paymentId")
	payerID := request.URL.Query().Get("PayerID")

	log.Printf("Received PaymentID: %s, PayerID: %s", paymentID, payerID) 

	if paymentID == "" || payerID == "" {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("payment ID or payer ID not provided"))
		return
	}

	client := CreatePayPalClient()

	// Execute payment via PayPal's sandbox API endpoint
	executeURL := fmt.Sprintf("https://api.sandbox.paypal.com/v1/payments/payment/%s/execute", paymentID)
	reqBody := map[string]string{"payer_id": payerID}
	reqBodyJSON, _ := json.Marshal(reqBody)

	req, err := http.NewRequest("POST", executeURL, bytes.NewBuffer(reqBodyJSON))
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error creating request: %v", err))
		return
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(clientID, clientSecret)

	resp, err := client.Do(req)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to execute payment: %v", err))
		return
	}
	defer resp.Body.Close()

	// Validate successful payment execution
	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		var errorResponse map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&errorResponse)
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("payment execution failed: %v", errorResponse))
		return
	}

	// Extract user ID from token
	userID, err := auth.GetStudentIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
		return
	}

	// Retrieve latest order by user ID
	order, err := h.order.GetLatestOrderByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not retrieve latest order: %v", err))
		return
	}

	// Retrieve order items by order ID
	orderItems, err := h.order.GetOrderItemsByOrderID(order.ID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not retrieve order items: %v", err))
		return
	}

	// Collect course IDs from the order items
	var courseIDs []int
	for _, item := range orderItems {
		courseIDs = append(courseIDs, item.CourseID)
	}

	// Enroll student after payment completion
	
	err = h.EnrollStudentAfterPayment(userID, courseIDs)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not enroll student after payment: %v", err))
		return
	}

	log.Printf("User ID: %d, Course IDs: %v", userID, courseIDs)

	err = h.order.UpdateOrderStatus(order.ID, "completed")
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not update order status: %v", err))
        return
    }

	err = h.cart.DeleteCartItems(userID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not delete cart items: %v", err))
        return
    }

	// Respond with success message
	response := map[string]string{
		"message": "Payment succeeded and student enrolled successfully",
		"status":  "completed",
	}
	utils.WriteJSON(writer, http.StatusOK, response)
}




func (h *Handler) HandlePayPalCancel(writer http.ResponseWriter, request *http.Request) {
	response := map[string]string{
		"message": "Payment canceled by the user",
		"status": "canceled",
	}
	utils.WriteJSON(writer, http.StatusOK, response)
}