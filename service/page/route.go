package page

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	page types.PageStore
	store  types.UserStore
	teacher types.TeacherStore
}


func NewHandler(page types.PageStore, store types.UserStore, teacher types.TeacherStore) *Handler {
    return &Handler{
		page: page,
		store: store,
		teacher: teacher,
	}
}



func (h *Handler) PageRoutes(router *mux.Router) {

	usersOnly := []types.UserRole{
		types.ADMIN,
		types.STUDENT,
	}

	router.HandleFunc("/lecture/courses", h.getCoursesHandle).Methods(http.MethodGet)
	router.HandleFunc("/course/{slug}", h.courseDetailsHandler).Methods(http.MethodGet)
	router.HandleFunc("/cart", auth.WithJWTAuth(h.cartHandler, h.store, usersOnly)).Methods(http.MethodGet)
}

func (h *Handler) getCoursesHandle(writer http.ResponseWriter, request *http.Request) {
	// Default values for pagination
	limit := 10
	offset := 0

	// Get query parameters for limit and offset
	queryParams := request.URL.Query()

	if l, ok := queryParams["limit"]; ok {
		parsedLimit, err := strconv.Atoi(l[0])
		if err == nil && parsedLimit > 0 {
			limit = parsedLimit
		}
	}

	if o, ok := queryParams["offset"]; ok {
		parsedOffset, err := strconv.Atoi(o[0])
		if err == nil && parsedOffset >= 0 {
			offset = parsedOffset
		}
	}

	// Fetch total number of courses
	totalCourses, err := h.teacher.CountCourses()
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to count courses: %v", err))
		return
	}

	// Fetch paginated courses
	courses, err := h.teacher.GetCourses(limit, offset)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch courses: %v", err))
		return
	}

	// Generate next and previous page URLs
	nextPageOffset := offset + limit
	prevPageOffset := offset - limit
	if prevPageOffset < 0 {
		prevPageOffset = 0
	}

	baseURL := request.URL.Path
	nextPageURL := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, nextPageOffset)
	prevPageURL := fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, prevPageOffset)

	// Response with metadata
	response := map[string]interface{}{
		"data":           courses,
		"total":          totalCourses,
		"limit":          limit,
		"offset":         offset,
		"next_page_url":  nextPageURL,
		"prev_page_url":  prevPageURL,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}


func (h *Handler) courseDetailsHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]

	courseDetail, err := h.page.GetCourseDetailBySlug(slug)
	if err != nil {
        if err == sql.ErrNoRows {  
            utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("course not found"))
            return
        }
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error getting course details: %v", err))
        return
    }

	utils.WriteJSON(writer, http.StatusOK, courseDetail)
}



func (h *Handler) cartHandler(writer http.ResponseWriter, request *http.Request) {
	
}