package teacher

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
	teacher types.TeacherStore
	store  types.UserStore
}



func NewHandler(teacher types.TeacherStore, store types.UserStore) *Handler {
	return &Handler{
		teacher: teacher,
		store: store,
	}
}



func (h *Handler) TeachRoutes(router *mux.Router){

	usersOnly := []types.UserRole{
		types.ADMIN,
		types.TEACHER,
	}

	// Categories
	router.HandleFunc("/course_builder/category", auth.WithJWTAuth(h.getCategoryHandle, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/course_builder/categories/{teacher_id}", auth.WithJWTAuth(h.getCategoriesByTeacherHandle, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/course_builder/category/create", auth.WithJWTAuth(h.createCategoryHandle, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/course_builder/category/edit/{id}", auth.WithJWTAuth(h.editCategoryHandle, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/course_builder/category/delete/{id}", auth.WithJWTAuth(h.deleteCategoryHandle, h.store, usersOnly)).Methods(http.MethodDelete)

	// courses
	router.HandleFunc("/course_builder/courses/create", auth.WithJWTAuth(h.createCourseHandle, h.store, usersOnly)).Methods((http.MethodPost))
	router.HandleFunc("/course_builder/course/edit/{id}", auth.WithJWTAuth(h.editCourseHandle, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/course_builder/course/delete/{id}", auth.WithJWTAuth(h.deleteCourseHandle, h.store, usersOnly)).Methods(http.MethodDelete)
	router.HandleFunc("/course_builder/courses", auth.WithJWTAuth(h.getCoursesHandle, h.store, usersOnly)).Methods(http.MethodGet)

	// sections
	router.HandleFunc("/course_builder/sections/create", auth.WithJWTAuth(h.createSectionHandle, h.store, usersOnly)).Methods(http.MethodPost)
}


// Category

func (h *Handler) createCategoryHandle(writer http.ResponseWriter, request *http.Request) {
	var payload types.CreateCategoryPayload

	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
	}

	userID, err := auth.GetTeacherIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized: %v", err))
        return
    }

	// Fetch the teacher associated with the user ID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
        return
    }


	category := &types.Category{
		TeacherID: teacher.ID,
		Name: payload.Name,
		Description: payload.Description,
	}

	err = h.teacher.CreateCategory(category)
	if err != nil {
		if err == ErrDuplicateCategory {
			utils.WriteError(writer, http.StatusConflict, fmt.Errorf("category with the same name already exists"))
			return
		}
		utils.WriteError(writer, http.StatusInternalServerError, err)
		return
	}

	category.Slug = utils.Slugify(payload.Name, category.ID)

	err = h.teacher.UpdateCategory(category)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }

	utils.WriteJSON(writer, http.StatusCreated, category)
}



func (h *Handler) getCategoryHandle(writer http.ResponseWriter, request *http.Request) {
	cs, err := h.teacher.GetCategory()
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }
	utils.WriteJSON(writer, http.StatusOK, cs)
}


func (h *Handler) getCategoriesByTeacherHandle(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	teacherID, err := strconv.Atoi(vars["teacher_id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid teacher ID: %s", vars["teacher_id"]))
        return
    }
	ts, err := h.teacher.GetCategoriesByTeacher(teacherID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }
	utils.WriteJSON(writer, http.StatusOK, ts)
}


func (h *Handler) editCategoryHandle(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetTeacherIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Fetch the teacher associated with the userID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
		return
	}

	vars := mux.Vars(request)
	categoryID, err := strconv.Atoi(vars["id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid category ID: %s", vars["id"]))
        return
    }

	category, err := h.teacher.GetCategoryByID(categoryID)
	if err != nil || category.TeacherID != teacher.ID {
		auth.PermissionDenied(writer, "you are not allowed to edit this category")
		return
	}

	var payload types.UpdateCategoryPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	if err := utils.Validate.Struct(payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	category.Name = payload.Name
	category.Description = payload.Description
	category.Slug = utils.Slugify(payload.Name, category.ID)


	err = h.teacher.UpdateCategory(category)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }

	utils.WriteJSON(writer, http.StatusOK, category)
}


func (h *Handler) deleteCategoryHandle(writer http.ResponseWriter, request *http.Request) {
    // Extract the teacher ID from the JWT token
    userID, err := auth.GetTeacherIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Fetch the teacher associated with the userID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
		return
	}

    // Extract category ID from the request parameters
    vars := mux.Vars(request)
    categoryID, err := strconv.Atoi(vars["id"])
    if err != nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid category ID: %s", vars["id"]))
        return
    }

    // Fetch the category to verify ownership
    category, err := h.teacher.GetCategoryByID(categoryID)
    if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch category"))
        return
    }

    // Check if the category belongs to the logged-in teacher
    if category.TeacherID != teacher.ID {
        auth.PermissionDenied(writer, "you do not have permission to delete this category")
        return
    }

    // Proceed with category deletion
    err = h.teacher.DeleteCategory(categoryID)
    if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete category"))
        return
    }

    // Send a success response
    response := map[string]string{"message": "Category deleted successfully"}
    utils.WriteJSON(writer, http.StatusOK, response)
}


// COURSE MANAGEMENT

func (h *Handler) createCourseHandle(writer http.ResponseWriter, request *http.Request) {
	var payload types.CreateCoursePayload

	// Parse the incoming JSON payload
	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	userID, err := auth.GetTeacherIDFromToken(request)
	if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized: %v", err))
        return
    }

	// Fetch the teacher associated with the user ID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
        return
    }

	// Create the course
	course := &types.Course{
		TeacherID:   teacher.ID,
		CategoryID:  payload.CategoryID,
		Name:        payload.Name,
		Description: payload.Description,
		Price:       payload.Price,
	}

	// Insert course into DB
	err = h.teacher.CreateCourse(course)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, err)
		return
	}

	// Generate slug and update the course
	course.Slug = utils.Slugify(payload.Name, course.ID)
	err = h.teacher.UpdateCourse(course)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, err)
		return
	}

	// Return success response
	utils.WriteJSON(writer, http.StatusCreated, course)
}




func (h *Handler) editCourseHandle(writer http.ResponseWriter, request *http.Request) {
	// Get user ID from the token
	userID, err := auth.GetTeacherIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Fetch the teacher associated with the userID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
		return
	}
	
	vars := mux.Vars(request)
	courseID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid course ID: %s", vars["id"]))
		return
	}

	// Get course by ID
	course, err := h.teacher.GetCourseByID(courseID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch course"))
		return
	}

	// Check if the teacher owns this course
	if course.TeacherID != teacher.ID {
		auth.PermissionDenied(writer, "you are not allowed to edit this course")
		return
	}

	// Parse the incoming update payload
	var payload types.UpdateCoursePayload
	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Validate payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Update course fields
	course.CategoryID = payload.CategoryID
	course.Name = payload.Name
	course.Description = payload.Description
	course.Slug = utils.Slugify(payload.Name, course.ID)
	course.Price = payload.Price

	// Perform course update
	err = h.teacher.UpdateCourse(course)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, err)
		return
	}

	utils.WriteJSON(writer, http.StatusOK, course)
}



func (h *Handler) deleteCourseHandle(writer http.ResponseWriter, request *http.Request) {
	userID, err := auth.GetTeacherIDFromToken(request)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
		return
	}

	// Fetch the teacher associated with the userID
	teacher, err := h.teacher.GetTeacherByUserID(userID)
	if err != nil {
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
		return
	}

	vars := mux.Vars(request)
    courseID, err := strconv.Atoi(vars["id"])
    if err != nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid course ID: %s", vars["id"]))
        return
    }
	course, err := h.teacher.GetCourseByID(courseID)
	if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch course"))
        return
    }


	if course.TeacherID != teacher.ID {
        auth.PermissionDenied(writer, "you do not have permission to delete this course")
        return
    }
	err = h.teacher.DeleteCourse(courseID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete course"))
        return
    }

	response := map[string]string{"message": "Course deleted successfully"}
	utils.WriteJSON(writer, http.StatusOK, response)
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



// func (h *Handler) getCourseHandle(writer http.ResponseWriter, request *http.Request) {
// 	vars := mux.Vars(request)
// 	courseID, err := strconv.Atoi(vars["id"])
// 	if err != nil {
// 		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid course ID"))
// 		return
// 	}

// 	course, err := h.teacher.GetCourseByID(courseID)
// 	if err != nil {
// 		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch course"))
// 		return
// 	}

// 	utils.WriteJSON(writer, http.StatusOK, course)
// }


// SECTION MANAGEMENT

func (h *Handler) createSectionHandle(writer http.ResponseWriter, request *http.Request) {
	var payload types.CreateSectionPayload

	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	section := &types.Section{
		CourseID: payload.CourseID,
		Title: payload.Title,
		Order: payload.Order,
	}

	err := h.teacher.CreateSection(section)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }

	utils.WriteJSON(writer, http.StatusCreated, section)
}