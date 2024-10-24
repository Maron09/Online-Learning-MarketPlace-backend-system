package types




type PageStore interface {
	GetCourseDetailBySlug(slug string) (map[string]interface{}, error)
	GetCourseSectionsAndVideos(courseID int) ([]map[string]interface{}, error)
}