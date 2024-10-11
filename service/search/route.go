package search

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	search types.SearchResult
}

func NewHandler(search types.SearchResult) *Handler {
    return &Handler{search: search}
}


func (h *Handler) SearchRoutes(router *mux.Router) {
	router.HandleFunc("/search", h.searchCoursesHandler).Methods(http.MethodGet)
}

func (h *Handler) searchCoursesHandler(writer http.ResponseWriter, request *http.Request) {
	queryParams := request.URL.Query()

	searchTerm := queryParams.Get("q")
	categoryIDs := queryParams.Get("category_id")
	categoryNames := queryParams.Get("category_name")
	teacherID := queryParams.Get("teacher_id")

	priceMin := queryParams.Get("price_min")
	priceMax := queryParams.Get("price_max")
	createdAfter := queryParams.Get("created_after")
	createdBefore := queryParams.Get("created_before")

	pageStr := queryParams.Get("page")
	limitStr := queryParams.Get("limit")

	page := 1
	limit := 10

	var err error
	if pageStr!= "" {
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid page parameter"))
			return
		}
	}

	if limitStr!= "" {
		limit, err = strconv.Atoi(limitStr)
        if err!= nil || limit < 1 || limit > 100 {
            utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid limit parameter"))
            return
        }
	}

	var categories []int

	if categoryNames != "" {
		names := strings.Split(categoryNames, ",")
		for _, name := range names {
			categoryID, err := h.search.GetCategoryIDByName(name)
			if err!= nil {
                utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to get category ID by name"))
                return
            }
			categories = append(categories, categoryID)
		}
	}

	if categoryIDs != "" {
		ids := strings.Split(categoryIDs, ",")
		for _, id := range ids {
			categoryID, err := strconv.Atoi(id)
			if err != nil {
				utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid category ID"))
				return
			}
			categories = append(categories, categoryID)
		}
	}

	var teacher int

	if teacherID!= "" {
		teacher, err = strconv.Atoi(teacherID)
        if err!= nil {
            utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid teacher ID"))
            return
        }
	}

	var priceMinFloat, priceMaxFloat float64

	if priceMin != "" {
		priceMinFloat, err = strconv.ParseFloat(priceMin, 64)
        if err!= nil {
            utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid price_min parameter"))
            return
        }
	}
	if priceMax!= "" {
		priceMaxFloat, err = strconv.ParseFloat(priceMax, 64)
		if err!= nil {
            utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid price_max parameter"))
            return
        }
	}

	offset := (page - 1) * limit

	results, totalCourses, err := h.search.SearchCourses(searchTerm, categories, teacher, priceMinFloat, priceMaxFloat, createdAfter, createdBefore, page, limit)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to search courses: %v", err))
        return
    }

	nextPageOffset := offset + limit
	prevPageOffset := offset - limit
	if prevPageOffset < 0 {
        prevPageOffset = 0
    }

	baseURL := request.URL.Path
	nextPageURL := fmt.Sprintf("%s?limit=%d&page=%d", baseURL, limit, nextPageOffset)
	prevPageURL := ""
	if page > 1 {
		prevPageURL = fmt.Sprintf("%s?limit=%d&page=%d", baseURL, limit, prevPageOffset)
	}

	response := map[string]interface{}{
		"data":           results,
		"total":          totalCourses,
		"limit":          limit,
		"offset":         offset,
		"current_page":   page,
		"next_page_url":  nextPageURL,
		"prev_page_url":  prevPageURL,
	}

	utils.WriteJSON(writer, http.StatusOK, response)
}