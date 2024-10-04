package teacher

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/types"
)

type Handler struct {
	teacher types.TeacherStore
}



func NewHandler(teacher types.TeacherStore) *Handler {
	return &Handler{
		teacher: teacher,
	}
}



func (h *Handler) TeachRoutes(router *mux.Router){
	router.HandleFunc("/register_teacher", h.registerTeacherHandler).Methods(http.MethodPost)
}



func (h *Handler) registerTeacherHandler(writer http.ResponseWriter, request *http.Request){

}