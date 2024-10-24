package page

import (
	"database/sql"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	page types.PageStore
}


func NewHandler(page types.PageStore) *Handler {
    return &Handler{page: page}
}



func (h *Handler) PageRoutes(router *mux.Router) {
	router.HandleFunc("/course/{slug}/", h.courseDetailsHandler).Methods(http.MethodGet)
}




// courseDetailsHandler handles fetching course details
// @Summary Get course details by slug
// @Description Retrieve course details including sections and videos
// @Tags courses
// @Param slug path string true "Course Slug"
// @Success 200 {object} map[string]interface{} "Success"
// @Failure 500 {object} utils.ErrorResponse "Internal Server Error"
// @Router /course/{slug}/ [get]
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