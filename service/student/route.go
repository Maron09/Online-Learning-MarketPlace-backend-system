package student

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	student types.StudentStore
	store types.UserStore
}

func NewHandler(student types.StudentStore, store types.UserStore) *Handler {
    return &Handler{
		student: student,
		store: store,
	}
}


func (h *Handler) StudentRoutes(router *mux.Router) {
	usersOnly := []types.UserRole{
		types.ADMIN,
		types.STUDENT,
	}
	
	router.HandleFunc("/student/learning", auth.WithJWTAuth(h.studentLearning, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/student/learning/{slug}", auth.WithJWTAuth(h.studentLearningDetail, h.store, usersOnly)).Methods(http.MethodGet)
}


func (h *Handler) studentLearning(writer http.ResponseWriter, request *http.Request) {
	studentID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	courses, err := h.student.GetEnrolledCourses(studentID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get enrolled courses: %v", err))
        return
    }
	utils.WriteJSON(writer, http.StatusOK, courses)
}



func (h *Handler) studentLearningDetail(writer http.ResponseWriter, request *http.Request) {
	studentID, err := auth.GetStudentIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized access"))
        return
    }

	vars := mux.Vars(request)
	slug := vars["slug"]
	if slug == "" {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("missing course slug"))
		return
	}

	courseData, err := h.student.GetEnrolledCoursesBySlug(studentID, slug)
	if err != nil {
		if err.Error() == "no course found for student" {
			utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("course not found for the given slug"))
		} else {
			utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("could not retrieve course details: %v", err))
		}
		return
	}

	utils.WriteJSON(writer, http.StatusOK, courseData)
}