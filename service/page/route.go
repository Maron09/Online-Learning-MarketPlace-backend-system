package page

import (
	"database/sql"
	"fmt"

	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	page types.PageStore
	store  types.UserStore
	teacher types.TeacherStore
	rating  types.RatingStore
}


func NewHandler(page types.PageStore, store types.UserStore, teacher types.TeacherStore, rating  types.RatingStore) *Handler {
    return &Handler{
		page: page,
		store: store,
		teacher: teacher,
		rating:  rating,
	}
}



func (h *Handler) PageRoutes(router *mux.Router) {


	router.HandleFunc("/lecture/courses", h.getCoursesHandle).Methods(http.MethodGet)
	router.HandleFunc("/course/{slug}", h.courseDetailsHandler).Methods(http.MethodGet)
	router.HandleFunc("/course/{slug}/ratings", h.courseRatingsHandler).Methods(http.MethodGet)
	router.HandleFunc("/teacher/profile/{id}", h.getTeacherProfile).Methods(http.MethodGet)
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

	// Create a slice to hold the courses with average ratings
	coursesWithRatings := make([]map[string]interface{}, len(courses))

	// Iterate through courses to get average ratings
	for i, course := range courses {
		avgRating, err := h.rating.GetAverageRating(course.ID) 
		
		if err != nil {
			utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get average rating for course ID %d: %v", course.ID, err))
			return
		}
		_, reviewCount, err := h.rating.CountRatingsAndReviews(course.ID)
		if err!= nil {
            utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get review count and ratings for course ID %d: %v", course.ID, err))
            return
        }

		// Create a map for the course with the average rating
		courseData := map[string]interface{}{
			"id" : course.ID,
			"name" : course.Name,
			"description": course.Description,
            "price":       course.Price,
            "image_url":  course.Image,
			"teacher" : map[string]string{
                "full_name" : fmt.Sprintf("%s %s", course.FirstName, course.LastName),
			},
			"rating":  avgRating,
			"review_count" : reviewCount,
		}
		coursesWithRatings[i] = courseData
	}

	// Generate next and previous page URLs
	nextPageOffset := offset + limit
	prevPageOffset := offset - limit
	if prevPageOffset < 0 {
		prevPageOffset = 0
	}

	baseURL := request.URL.Path
	nextPageURL := ""
	prevPageURL := ""

	// Assign URLs only if they are valid
	if nextPageOffset < totalCourses {
		nextPageURL = fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, nextPageOffset)
	}
	if prevPageOffset >= 0 {
		prevPageURL = fmt.Sprintf("%s?limit=%d&offset=%d", baseURL, limit, prevPageOffset)
	}

	// Response with metadata
	response := map[string]interface{}{
		"data":           coursesWithRatings,
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
	if slug == "" {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("missing course slug"))
		return
	}

	// Fetch course details by slug
	courseDetail, err := h.page.GetCourseDetailBySlug(slug)
	if err != nil {
		if err == sql.ErrNoRows {
			utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("course not found"))
			return
		}
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error getting course details: %v", err))
		return
	}

	// Get course ID
	courseID, ok := courseDetail["id"].(int)
	if !ok {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("invalid course ID"))
		return
	}

	// Fetch average rating for the course
	averageRating, err := h.rating.GetAverageRating(courseID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error getting average rating: %v", err))
		return
	}


	previewLimit := 5
	previewRatings, _, err := h.rating.GetRatingsForCourse(slug, 1, previewLimit)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error getting preview ratings: %v", err))
		return
	}

	// Construct URL to view all ratings
	viewAllRatingsURL := fmt.Sprintf("/api/v1/course/%s/ratings?limit=10&page=1", slug)

	// Create a response structure to include course details, average rating, preview ratings, and the view-all link
	response := map[string]interface{}{
		"course":               courseDetail,
		"rating":      			averageRating,
		"preview_ratings":      previewRatings,
		"view_all_ratings_and_reviews_url": viewAllRatingsURL,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}




func (h *Handler) courseRatingsHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	slug := vars["slug"]
	if slug == "" {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("missing course slug"))
		return
	}

	// Parse pagination parameters
	pageStr := request.URL.Query().Get("page")
	limitStr := request.URL.Query().Get("limit")

	// Set default values for pagination
	page := 1  // default to first page
	limit := 10 // default to 10 items per page

	// Parse page number
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid page number"))
			return
		}
	}

	// Parse limit number
	if limitStr != "" {
		var err error
		limit, err = strconv.Atoi(limitStr)
		if err != nil || limit < 1 {
			utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid limit number"))
			return
		}
	}

	// Calculate offsets
	offset := (page - 1) * limit

	// Call the function to get ratings with pagination
	ratings, totalRatings, err := h.rating.GetRatingsForCourse(slug, page, limit)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error getting ratings for course: %v", err))
		return
	}

	// Generate next and previous page URLs
	nextPageOffset := offset + limit
	prevPageOffset := offset - limit
	if prevPageOffset < 0 {
		prevPageOffset = 0
	}

	baseURL := request.URL.Path
	nextPageURL := ""
	prevPageURL := ""

	// Assign URLs only if they are valid
	if nextPageOffset < totalRatings {
		nextPageURL = fmt.Sprintf("%s?limit=%d&page=%d", baseURL, limit, (nextPageOffset/limit)+1)
	}
	if prevPageOffset >= 0 {
		prevPageURL = fmt.Sprintf("%s?limit=%d&page=%d", baseURL, limit, (prevPageOffset/limit)+1)
	}

	// Prepare response map
	response := map[string]interface{}{
		"data":          ratings,
		"total":         totalRatings,
		"limit":         limit,
		"offset":        offset,
		"next_page_url": nextPageURL,
		"prev_page_url": prevPageURL,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}



func (h *Handler) getTeacherProfile(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	teacherID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid teacher ID"))
		return
	}


	teacherProfile, err := h.teacher.GetTeacherProfileByID(teacherID)
	if err!= nil {
		if err == sql.ErrNoRows {
			utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("teacher not found"))
			return
		}
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error fetching teacher profile: %v", err))
		return
	}

	totalCourses, err := h.teacher.CountCoursesByTeacher(teacherID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error counting courses: %v", err))
        return
    }

	totalEnrolledStudents, err := h.teacher.CountEnrolledStudents(teacherID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error counting enrolled students: %v", err))
        return
    }

	totalRatings, totalReviews, err := h.teacher.CountRatingsAndReviewsForTeacher(teacherID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("error counting ratings and reviews: %v", err))
        return
    }
	response := map[string]interface{}{
		"profile":           teacherProfile,
        "total_courses":      totalCourses,
        "total_students": totalEnrolledStudents,
        "total_ratings":      totalRatings,
        "total_reviews":     totalReviews,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}