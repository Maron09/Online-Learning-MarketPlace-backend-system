package types




type StudentStore interface {
	GetEnrolledCourses(studentID int) ([]map[string]interface{}, error)
	GetEnrolledCoursesBySlug(studentID int, slug string) (map[string]interface{}, error)
	GetEnrolledCourseSectionsAndVideos(courseID int) ([]map[string]interface{}, error)
}