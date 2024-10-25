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
	router.HandleFunc("/course_builder/categories", auth.WithJWTAuth(h.getCategoriesByTeacherHandle, h.store, usersOnly)).Methods(http.MethodGet)
	router.HandleFunc("/course_builder/category/create", auth.WithJWTAuth(h.createCategoryHandle, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/course_builder/category/edit/{id}", auth.WithJWTAuth(h.editCategoryHandle, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/course_builder/category/delete/{id}", auth.WithJWTAuth(h.deleteCategoryHandle, h.store, usersOnly)).Methods(http.MethodDelete)

	// courses
	router.HandleFunc("/course_builder/courses/create", auth.WithJWTAuth(h.createCourseHandle, h.store, usersOnly)).Methods((http.MethodPost))
	router.HandleFunc("/course_builder/course/edit/{id}", auth.WithJWTAuth(h.editCourseHandle, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/course_builder/course/delete/{id}", auth.WithJWTAuth(h.deleteCourseHandle, h.store, usersOnly)).Methods(http.MethodDelete)

	// sections
	router.HandleFunc("/course_builder/sections/create", auth.WithJWTAuth(h.createSectionHandle, h.store, usersOnly)).Methods(http.MethodPost)
	router.HandleFunc("/course_builder/section/edit/{id}", auth.WithJWTAuth(h.editSectionHandle, h.store, usersOnly)).Methods(http.MethodPatch)
	router.HandleFunc("/course_builder/section/delete/{id}", auth.WithJWTAuth(h.deleteSectionHandle, h.store, usersOnly)).Methods(http.MethodDelete)

	// videos
	router.HandleFunc("/course_builder/videos/create", auth.WithJWTAuth(h.createVideoHandle, h.store, usersOnly)).Methods(http.MethodPost)
    router.HandleFunc("/course_builder/video/edit/{id}", auth.WithJWTAuth(h.editVideoHandle, h.store, usersOnly)).Methods(http.MethodPatch)
    router.HandleFunc("/course_builder/video/delete/{id}", auth.WithJWTAuth(h.deleteVideoHandle, h.store, usersOnly)).Methods(http.MethodDelete)

	// course by category
	router.HandleFunc("/course_builder/category/{id}", auth.WithJWTAuth(h.courseByCategoryHandle, h.store, usersOnly)).Methods(http.MethodGet)

	// section by course
	router.HandleFunc("/course_builder/category/{category_id}/course/{course_id}/section/{id}", auth.WithJWTAuth(h.sectionByCourseHandle, h.store, usersOnly)).Methods(http.MethodGet)

    // video by section
    router.HandleFunc("/course_builder/category/{category_id}/course/{course_id}/section/{section_id}/video/{id}", auth.WithJWTAuth(h.videoBySectionHandle, h.store, usersOnly)).Methods(http.MethodGet)
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

	response := map[string]interface{}{
		"message": "Category Created successfully",
		"video":   category,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
}



func (h *Handler) getCategoriesByTeacherHandle(writer http.ResponseWriter, request *http.Request) {
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

	teacherID := teacher.ID
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

	response := map[string]interface{}{
		"message": "Category updated successfully",
		"video":   category,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
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
	response := map[string]interface{}{
		"message": "Course Created successfully",
		"video":   course,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
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

	response := map[string]interface{}{
		"message": "Course updated successfully",
		"video":   course,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
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

	response := map[string]interface{}{
		"message": "Section Created successfully",
		"video":   section,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
}




func (h *Handler) editSectionHandle(writer http.ResponseWriter, request *http.Request) {
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
	sectionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid section ID: %s", vars["id"]))
		return
	}

	section, err := h.teacher.GetSectionByID(sectionID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch section: %v", err))
		return
	}

	// Handle case where section is not found
	if section == nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("section not found for ID: %d", sectionID))
		return
	}

	courses, err := h.teacher.GetCoursesByTeacherID(teacher.ID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch courses"))
        return
    }

	var courseBelongsToTeacher bool

	for _, course := range courses {
		if course.ID == section.CourseID {
			courseBelongsToTeacher = true
			break
		}
	}

	if !courseBelongsToTeacher {
		auth.PermissionDenied(writer, "you do not have permission to edit this section")
        return
	}

	var payload types.UpdateSectionPayload
	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	// Validate payload
	if err := utils.Validate.Struct(payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

    section.Title = payload.Title
    section.Order = payload.Order

    err = h.teacher.UpdateSection(section)
    if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }

    response := map[string]interface{}{
		"message": "Section updated successfully",
		"video":   section,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
}



func (h *Handler) deleteSectionHandle(writer http.ResponseWriter, request *http.Request) {
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
	sectionID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid section ID: %s", vars["id"]))
		return
	}

	section, err := h.teacher.GetSectionByID(sectionID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch section: %v", err))
		return
	}

	// Handle case where section is not found
	if section == nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("section not found for ID: %d", sectionID))
		return
	}

	courses, err := h.teacher.GetCoursesByTeacherID(teacher.ID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch courses"))
        return
    }
	var courseBelongsToTeacher bool

	for _, course := range courses {
		if course.ID == section.CourseID {
			courseBelongsToTeacher = true
			break
		}
	}

	if !courseBelongsToTeacher {
		auth.PermissionDenied(writer, "you do not have permission to delete this section")
        return
	}

	err = h.teacher.DeleteSection(sectionID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete section: %v", err))
        return
    }

	response := map[string]string{"message": "section deleted successfully"}

	utils.WriteJSON(writer, http.StatusOK, response)
}


// VIDEO MANAGEMENT

func (h *Handler) createVideoHandle(writer http.ResponseWriter, request *http.Request) {
	var payload types.CreateVideoPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	
	video := &types.Video{
		SectionID: payload.SectionID,
		Title: payload.Title,
		VideoFile: payload.VideoFile,
		Order: payload.Order,
	}

	err := h.teacher.CreateVideo(video)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }
	response := map[string]interface{}{
		"message": "Video Created successfully",
		"video":   video,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
}




func (h *Handler) editVideoHandle(writer http.ResponseWriter, request *http.Request) {
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

	// Get video ID from the URL params
	vars := mux.Vars(request)
	videoID, err := strconv.Atoi(vars["id"])
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid video ID: %s", vars["id"]))
		return
	}

	// Fetch the video by videoID
	video, err := h.teacher.GetVideoByID(videoID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch video: %v", err))
		return
	}

	// Fetch the section associated with the video
	section, err := h.teacher.GetSectionByID(video.SectionID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch section: %v", err))
		return
	}

	// Fetch the courses associated with the teacher
	courses, err := h.teacher.GetCoursesByTeacherID(teacher.ID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch courses for the teacher"))
		return
	}

	var courseBelongsToTeacher bool
	for _, course := range courses {
		if course.ID == section.CourseID {
			courseBelongsToTeacher = true
			break
		}
	}

	if !courseBelongsToTeacher {
		auth.PermissionDenied(writer, "you do not have permission to edit this video")
		return
	}

	var payload types.UpdateVideoPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	// Validate the payload
	if err := utils.Validate.Struct(payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	video.Title = payload.Title
	video.VideoFile = payload.VideoFile
	video.Order = payload.Order

	err = h.teacher.UpdateVideo(video)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, err)
        return
    }
	
	response := map[string]interface{}{
		"message": "Video updated successfully",
		"video":   video,
	}
	
	utils.WriteJSON(writer, http.StatusOK, response)
}


func (h *Handler) deleteVideoHandle(writer http.ResponseWriter, request *http.Request) {
	// Get user ID from the token
    userID, err := auth.GetTeacherIDFromToken(request)
    if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("unauthorized"))
        return
    }

    // Fetch the teacher associated with the userID
    teacher, err := h.teacher.GetTeacherByUserID(userID)
    if err!= nil {
        utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("teacher not found for this user"))
        return
    }

    // Get video ID from the URL params
    vars := mux.Vars(request)
    videoID, err := strconv.Atoi(vars["id"])
    if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid video ID: %s", vars["id"]))
        return
    }

	// Fetch the video by videoID
	video, err := h.teacher.GetVideoByID(videoID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch video: %v", err))
		return
	}

	// Fetch the section associated with the video
	section, err := h.teacher.GetSectionByID(video.SectionID)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("failed to fetch section: %v", err))
		return
	}

	// Fetch the courses associated with the teacher
	courses, err := h.teacher.GetCoursesByTeacherID(teacher.ID)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to fetch courses for the teacher"))
		return
	}

	var courseBelongsToTeacher bool
	for _, course := range courses {
		if course.ID == section.CourseID {
			courseBelongsToTeacher = true
			break
		}
	}

	if !courseBelongsToTeacher {
		auth.PermissionDenied(writer, "you do not have permission to delete this video")
		return
	}

	err = h.teacher.DeleteVideo(videoID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete video: %v", err))
        return
    }

	response := map[string]string{"message": "Video deleted successfully"}
	
    utils.WriteJSON(writer, http.StatusOK, response)
}


// COURSE BY CATEGORY
func (h *Handler) courseByCategoryHandle(writer http.ResponseWriter, request *http.Request) {
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

	courses, err := h.teacher.GetCoursesByCategory(categoryID, teacher.ID)
	if err != nil  {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("could not fetch courses: %v", err))
		return
	}
	
    utils.WriteJSON(writer, http.StatusOK, courses)
}


// SECTION BY COURSE
func (h *Handler) sectionByCourseHandle(writer http.ResponseWriter, request *http.Request) {
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
	courseID, err := strconv.Atoi(vars["course_id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid course ID: %s", vars["id"]))
        return
    }

	sections, err := h.teacher.GetSectionsByCourse(courseID, teacher.ID)
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("could not fetch sections: %v", err))
        return
    }

	utils.WriteJSON(writer, http.StatusOK, sections)
}


// VIDEO BY SECTION
func (h *Handler) videoBySectionHandle(writer http.ResponseWriter, request *http.Request) {
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
	sectionID, err := strconv.Atoi(vars["section_id"])
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid section ID: %s", vars["id"]))
        return
    }

	videos, err := h.teacher.GetVideosBySection(sectionID, teacher.ID)
	if err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("could not fetch videos: %v", err))
        return
    }

	utils.WriteJSON(writer, http.StatusOK, videos)
}