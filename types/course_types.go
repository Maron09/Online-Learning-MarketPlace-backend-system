package types

import (
	"time"

)

type TeacherStore interface {

	GetTeacherByUserID(userID int) (*Teacher, error)
	// category CRUD operations
	CreateCategory(category *Category) error
	GetCategory() ([]Category, error)
	GetCategoriesByTeacher(teacherID int) ([]Category, error)
	GetCategoryByID(id int) (*Category, error)
	UpdateCategory(category *Category) error
	DeleteCategory(id int) error

	// course CRUD operations
	CreateCourse(course *Course) error
	GetCourses(limit, offset int) ([]Course, error)
	CountCourses() (int, error)
	GetCourseByID(id int) (*Course, error)
	UpdateCourse(course *Course) error
	DeleteCourse(id int) error
	CheckDuplicateCourse(teacherID int, courseName string) (bool, error)

	// section CRUD operations
	CreateSection(section *Section) error
	GetSectionByID(id int) (*Section, error)
	UpdateSection(section *Section) error
	DeleteSection(id int) error

	// Video CRUD operations
	// CreateVideo(video *Video) error
	// GetVideoByID(id int) (*Video, error)
	// UpdateVideo(video *Video) error
	// DeleteVideo(id int) error
}

type Category struct {
	ID          int    `json:"id"`
	TeacherID   int    `json:"teacher_id"`
	Name        string `json:"name"`
	Slug        string `json:"slug"`
	Description string `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	ModifiedAt 	time.Time `json:"modified_at"`
}


type Course struct {
	ID                int    `json:"id"`
	TeacherID         int    `json:"teacher_id"`
	CategoryID        int    `json:"category_id"`
	Name 			  string `json:"name"`
	Slug              string `json:"slug"`
	Description       string `json:"description"`
	ForWho      	  string    `json:"for_who,omitempty"` 
	Reason      	  string    `json:"reason,omitempty"`  
	IntroVideo  	  string    `json:"intro_video,omitempty"`
	Image             string `json:"image"`
	Price             float64 `json:"price"`
	CreatedAt 		  time.Time `json:"created_at"`
	ModifiedAt 		  time.Time `json:"modified_at"`
}


type Section struct {
	ID                int    `json:"id"`
	CourseID          int    `json:"course_id"`
	Title             string  `json:"title"`
	Order             int     `json:"order"`
	CreatedAt 		  time.Time `json:"created_at"`
	ModifiedAt 		  time.Time `json:"modified_at"`
}



type Video struct {
	ID                int    `json:"id"`
	SectionID         int    `json:"section_id"`
	Title             string  `json:"title"`
	VideoFile 		  string `json:"video_file"`
	Order 			  int 	 `json:"order"`
	CreatedAt 		  time.Time `json:"created_at"`
	ModifiedAt 		  time.Time `json:"modified_at"`
}


// Payloads
type CreateCategoryPayload struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}


type UpdateCategoryPayload struct {
	Name        string `json:"name" validate:"required"`
	Description string `json:"description"`
}

type CreateCoursePayload struct {
	CategoryID  int     `json:"category_id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ForWho      string  `json:"for_who"`
	Reason      string  `json:"reason"`
	IntroVideo  string  `json:"intro_video"`
	Image       string  `json:"image"`
	Price       float64 `json:"price" validate:"required,min=0"`
}


type UpdateCoursePayload struct {
	CategoryID  int     `json:"category_id" validate:"required"`
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description"`
	ForWho      string  `json:"for_who"`
	Reason      string  `json:"reason"`
	IntroVideo  string  `json:"intro_video"`
	Image       string  `json:"image"`
	Price       float64 `json:"price" validate:"min=0"`
}

type CreateSectionPayload struct {
	CourseID int    `json:"course_id" validate:"required"`
	Title    string `json:"title" validate:"required"`
	Order    int    `json:"order"`
}

type UpdateSectionPayload struct {
    Title    string `json:"title" validate:"required"`
    Order    int    `json:"order"`
}

type CreateVideoPayload struct {
	SectionID int    `json:"section_id" validate:"required"`
	Title     string `json:"title" validate:"required"`
	VideoFile string `json:"video_file" validate:"required,url"`
	Order     int    `json:"order"`
}

type UpdateVideoPayload struct {
    Title     string `json:"title" validate:"required"`
    VideoFile string `json:"video_file" validate:"omitempty,url"`
    Order     int    `json:"order"`
}
