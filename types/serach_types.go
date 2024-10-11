package types


type SearchResult interface {
	GetCategoryIDByName(name string) (int, error)
	SearchCourses(searchTerm string, categories []int, teacherID int, priceMin, priceMax float64, createdAfter, createdBefore string, page, limit int) ([]map[string]interface{}, int, error)
}