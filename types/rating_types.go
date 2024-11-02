package types

import "time"

type RatingStore interface {
	CreateRating(rating *Rating) (int, error)
    UpdateRating(ratingID int, updatedRating *Rating) error
    GetRating(ratingID int) (*Rating, error)
    GetRatingsForCourse(slug string, page int, limit int) ([]Rating, int, error)
    GetRatingsByStudent(studentID int) ([]Rating, error)
	GetRatingByStudentID(ratingID int, studentID int) (*Rating, error)
    DeleteRating(ratingID int) error
    GetAverageRating(courseID int) (float32, error)
    CountRatingsAndReviews(courseID int) (int, int, error)
	IsStudentEnrolledInCourse(studentID int, courseID int) (bool, error)
}


type Rating struct {
	ID        int       `json:"id"`
	StudentID int       `json:"student_id"`
	StudentFirstName string `json:"student_first_name"`
	StudentLastName  string `json:"student_last_name"`
	CourseID  int       `json:"course_id"`
	CourseName string    `json:"course_name"`
	Rating    float32   `json:"rating"`
	Review    *string   `json:"review,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}


type RatingPayload struct {
	CourseID  int       `json:"course_id"`
	Rating    float32   `json:"rating"`
	Review    *string   `json:"review,omitempty"`
}

