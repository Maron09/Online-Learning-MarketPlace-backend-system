package rating

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	rating types.RatingStore
	store  types.UserStore
}


func NewHandler(rating types.RatingStore, store types.UserStore) *Handler {
	return &Handler{
		rating: rating,
        store:  store,
    }
}


func (h *Handler) RatingRoutes(router *mux.Router) {

	usersOnly := []types.UserRole{
		types.ADMIN,
		types.STUDENT,
	}

	router.HandleFunc("/student/rating/post", auth.WithJWTAuth(h.createRatingHandler, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/student/rating/edit/{id}", auth.WithJWTAuth(h.updateRatingHandler, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/student/rating/{id}", auth.WithJWTAuth(h.GetRatingHandler, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/student/ratings", auth.WithJWTAuth(h.GetStudentRatingHandler, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/student/ratings", auth.WithJWTAuth(h.GetStudentRatingHandler, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/student/rating/delete/{id}", auth.WithJWTAuth(h.deleteRatingHandler, h.store, usersOnly)).Methods(http.MethodDelete)
}





func (h *Handler) createRatingHandler(writer http.ResponseWriter, request *http.Request) {
	studentID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	var payload types.RatingPayload
	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	// Validate the payload
	if err := utils.Validate.Struct(payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	enrolled, err := h.rating.IsStudentEnrolledInCourse(studentID, payload.CourseID)
    if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error checking enrollment: %v", err))
        return
    }
    if !enrolled {
        utils.WriteError(writer, http.StatusForbidden, fmt.Errorf("you must be enrolled in the course to rate it"))
        return
    }
	if err := utils.ValidateRating(payload.Rating); err != nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	
	rating := &types.Rating{
        StudentID: studentID,
        CourseID: payload.CourseID,
        Rating: payload.Rating,
		Review: payload.Review,
    }

	_, err = h.rating.CreateRating(rating) 
    if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create rating: %v", err))
        return
    }


	response := map[string]string{"message": "rating created successfully"}
	utils.WriteJSON(writer, http.StatusOK, response)

}


func (h *Handler) updateRatingHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	ratingID, err := strconv.Atoi(vars["id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid rating ID: %s", vars["id"]))
        return
    }
	studentID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	var payload types.RatingPayload
	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
	}
	if err := utils.ValidateRating(payload.Rating); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	rating, err := h.rating.GetRating(ratingID)
	if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("rating not found: %v", err))
        return
    }
	if rating.StudentID!= studentID {
        utils.WriteError(writer, http.StatusForbidden, fmt.Errorf("you cannot update this rating"))
        return
    }

	updateRating := &types.Rating{
		Rating: payload.Rating,
		Review: payload.Review,
	}

	if err := h.rating.UpdateRating(ratingID, updateRating); err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to update rating: %v", err))
		return
	}
	response := map[string]string{"message": "rating updated successfully"}
	utils.WriteJSON(writer, http.StatusOK, response)
}


func (h *Handler) GetRatingHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
    ratingID, err := strconv.Atoi(vars["id"])
    if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid rating ID: %s", vars["id"]))
        return
    }

	studentID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

    rating, err := h.rating.GetRatingByStudentID(ratingID, studentID)
    if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("rating not found: %v", err))
        return
    }

    utils.WriteJSON(writer, http.StatusOK, rating)
}


func (h *Handler) GetStudentRatingHandler(writer http.ResponseWriter, request *http.Request) {
	studentID, err := auth.GetStudentIDFromToken(request)
    if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

    ratings, err := h.rating.GetRatingsByStudent(studentID)
    if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get student ratings: %v", err))
        return
    }

    utils.WriteJSON(writer, http.StatusOK, ratings)
}


func (h *Handler) deleteRatingHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
    ratingID, err := strconv.Atoi(vars["id"])
    if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid rating ID: %s", vars["id"]))
        return
    }

    studentID, err := auth.GetStudentIDFromToken(request)
    if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

    rating, err := h.rating.GetRating(ratingID)
    if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("rating not found: %v", err))
        return
    }
    if rating.StudentID!= studentID {
        utils.WriteError(writer, http.StatusForbidden, fmt.Errorf("you cannot delete this rating"))
        return
    }

    err = h.rating.DeleteRating(ratingID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete rating: %v", err))
        return
    }
	response := map[string]string{"message": "rating deleted successfully"}
	utils.WriteJSON(writer, http.StatusNoContent, response)
}