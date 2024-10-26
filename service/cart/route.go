package cart

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	cart types.CartStore
	store types.UserStore
}


func NewHandler(cart types.CartStore, store types.UserStore) *Handler {
    return &Handler{
		cart: cart,
        store: store,
	}
}


func (h *Handler) CartRoutes(router *mux.Router) {
	usersOnly := []types.UserRole{
		types.ADMIN,
		types.STUDENT,
	}


	router.HandleFunc("/cart", auth.WithJWTAuth(h.cartHandler, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/cart/add", auth.WithJWTAuth(h.addToCartHandler, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/cart/{id}", auth.WithJWTAuth(h.deleteFromCartHandler, h.store, usersOnly)).Methods(http.MethodDelete)
}


// Get the User cart items
func (h *Handler) cartHandler(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	cartItems, err := h.cart.GetCart(userID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get cart: %v", err))
        return
    }
	utils.WriteJSON(writer, http.StatusOK, cartItems)
}


// Add to cart
func (h *Handler) addToCartHandler(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	var payload types.AddToCartPayload
	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	if err := utils.Validate.Struct(payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	exists, err := h.cart.CheckIfCourseInCart(userID, payload.CourseID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to check if course in cart: %v", err))
        return
    }
	if exists {
        utils.WriteError(writer, http.StatusConflict, fmt.Errorf("course already in cart"))
        return
    }

	cartItem := types.Cart{
		UserID: userID,
        CourseID: payload.CourseID,
		CreatedAt: time.Now(),
		ModifiedAt: time.Now(),
	}

	err = h.cart.AddToCart(&cartItem)
	if err!= nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to add to cart: %v", err))
        return
	}

	response := map[string]string{
		"message": "course added to cart successfully",
	}
	utils.WriteJSON(writer, http.StatusOK, response)
}

// Delete From Cart
func (h *Handler) deleteFromCartHandler(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }
	vars := mux.Vars(request)
	cartID, err := strconv.Atoi(vars["id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid cart ID: %s", vars["id"]))
        return
    }

	err = h.cart.DeleteFromCart(cartID, userID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete from cart: %v", err))
        return
    }
	response := map[string]string{
        "message": "course deleted from cart successfully",
    }

	utils.WriteJSON(writer, http.StatusOK, response)
}