package teacher

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/sikozonpc/ecom/types"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}


var ErrDuplicateCategory = errors.New("category already exists")
var ErrDuplicateCourse = errors.New("course with the same name already exists")

// GetTeacherByUserID retrieves the teacher associated with the given user ID
func (s *Store) GetTeacherByUserID(userID int) (*types.Teacher, error) {
	var teacher types.Teacher
	query := `SELECT id, user_id, user_profile_id, bio, profession, certificate, is_approved, created_at, modified_at 
			FROM teachers 
			WHERE user_id = $1`

	err := s.db.QueryRow(query, userID).Scan(
		&teacher.ID,
		&teacher.UserID,
		&teacher.UserProfileID,
		&teacher.Bio,
		&teacher.Profession,
		&teacher.Certificate,
		&teacher.IsApproved,
		&teacher.CreatedAt,
		&teacher.ModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("teacher not found for user ID: %d", userID)
		}
		return nil, err
	}

	return &teacher, nil
}


func (s *Store) GetCoursesByTeacherID(teacherID int) ([]types.Course, error) {
	var courses []types.Course

	query := `SELECT id, teacher_id, category_id, name, slug, description, price 
			FROM courses 
			WHERE teacher_id = $1`
	
	rows, err := s.db.Query(query, teacherID)
	if err!= nil {
        return nil, err
    }
	defer rows.Close()

	for rows.Next() {
		var course types.Course
		err := rows.Scan(
			&course.ID,
			&course.TeacherID,
			&course.CategoryID,
			&course.Name,
			&course.Slug,
			&course.Description,
			&course.Price,
		)
		if err != nil {
			return nil, err
		}
		courses = append(courses, course)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return courses, nil
}


func (s *Store) CreateCategory(category *types.Category) error {
	var existingCategoryID int
	err := s.db.QueryRow("SELECT id FROM categories WHERE slug = $1", category.Slug).Scan(&existingCategoryID)
	
	if err != nil && err != sql.ErrNoRows { 
        // If there's an error other than no rows, return it
        return err
    }
	
    if err == nil { 
        // If no error, that means a category with the same slug exists (duplicate)
        return ErrDuplicateCategory
    }

	query := `INSERT INTO categories (teacher_id, name, slug, description, created_at, modified_at)
	        VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id;`
	
	// Insert the new category and return the new ID
	err = s.db.QueryRow(query, category.TeacherID, category.Name, category.Slug, category.Description).Scan(&category.ID)
	if err != nil {
		return err
	}

	return nil
}


func (s *Store) GetCategory() ([]types.Category, error) {
	query := `SELECT id, teacher_id, name, slug, description, created_at, modified_at FROM categories`

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []types.Category
	for rows.Next() {
		var category types.Category
		err := rows.Scan(
			&category.ID,
			&category.TeacherID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.CreatedAt,
			&category.ModifiedAt,
		)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}


func (s *Store) GetCategoriesByTeacher(teacherID int) ([]types.Category, error) {
	query := `SELECT id, teacher_id, name, slug, description, created_at, modified_at 
	        FROM categories WHERE teacher_id = $1`

	rows, err := s.db.Query(query, teacherID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []types.Category
	for rows.Next() {
		var category types.Category
		err := rows.Scan(
			&category.ID,
			&category.TeacherID,
			&category.Name,
			&category.Slug,
			&category.Description,
			&category.CreatedAt,
			&category.ModifiedAt,
		)
		if err != nil {
			return nil, err
		}

		categories = append(categories, category)
	}

	// Check for any error encountered during iteration
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}



func (s *Store) GetCategoryByID(id int) (*types.Category, error) {
	query := `SELECT id, teacher_id, name, slug, description, created_at, modified_at
	    	FROM categories WHERE id = $1;`

	var category types.Category
	err := s.db.QueryRow(query, id).Scan(
		&category.ID,
		&category.TeacherID,
        &category.Name,
        &category.Slug,
        &category.Description,
        &category.CreatedAt,
		&category.ModifiedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
            return nil, nil
        }
        return nil, err
	}
	return &category, nil
}


func (s *Store) UpdateCategory(category *types.Category) error {
	query := `UPDATE categories SET name = $1, slug = $2, description = $3, modified_at = NOW()
	    	WHERE id = $4;`
	
	_, err := s.db.Exec(query, category.Name, category.Slug, category.Description, category.ID)
	if err != nil {
		return err
	}
	return nil
}


func (s *Store) DeleteCategory(id int) error {
	query := `DELETE FROM categories WHERE id = $1;`

	_, err := s.db.Exec(query, id)
	if err!= nil {
        return err
    }
	return nil
}

// COURSE MANAGEMENT

func (s *Store) CheckDuplicateCourse(teacherID int, courseName string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM courses WHERE teacher_id = $1 AND name = $2)`
	err := s.db.QueryRow(query, teacherID, courseName).Scan(&exists)
	if err != nil {
		return false, err
	}
	return exists, nil
}



func (s *Store) CreateCourse(course *types.Course) error {
	query := `INSERT INTO courses (teacher_id, category_id, name, slug, description, price, created_at, modified_at)
	    	VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW()) RETURNING id`
	err := s.db.QueryRow(query, course.TeacherID, course.CategoryID, course.Name, course.Slug, course.Description, course.Price).Scan(&course.ID)
	if err!= nil {
        return err
    }
	return nil
}


func (s *Store) UpdateCourse(course *types.Course) error {
	query := `UPDATE courses SET category_id = $1, name = $2, slug = $3, description = $4, price = $5, modified_at = NOW()
	        WHERE id = $6`
	_, err := s.db.Exec(query, course.CategoryID, course.Name, course.Slug, course.Description, course.Price, course.ID)
	if err != nil {

		return err
	}
	return nil
}

func (s *Store) GetCourseByID(courseID int) (*types.Course, error) {
    var course types.Course
    query := `SELECT id, teacher_id, category_id, name, slug, description, price, created_at  FROM courses WHERE id = $1`
    err := s.db.QueryRow(query, courseID).Scan(
        &course.ID,
        &course.TeacherID,
        &course.CategoryID,
        &course.Name,
        &course.Slug,
        &course.Description,
        &course.Price,
		&course.CreatedAt,
    )
    if err != nil {
        return nil, err
    }
    return &course, nil
}



func (s *Store) DeleteCourse(id int) error {
	query := `DELETE FROM courses WHERE id = $1;`
    _, err := s.db.Exec(query, id)
    if err!= nil {
        return err
    }
    return nil
}


func (s *Store) GetCourses(limit, offset int) ([]types.Course, error) {
	query := `SELECT c.id, c.teacher_id, u.first_name, u.last_name, c.category_id, c.name, c.slug, c.description, c.image, c.price, c.created_at  
	FROM courses AS c
	JOIN teachers AS t ON c.teacher_id = t.id
	JOIN users AS u ON t.user_id = u.id
	ORDER BY created_at 
	DESC LIMIT $1 OFFSET $2
	`

	rows, err := s.db.Query(query, limit, offset)
	if err!= nil {
        return nil, err
    }
	defer rows.Close()
	var courses []types.Course

	for rows.Next() {
		var course types.Course
		var image sql.NullString
        err := rows.Scan(
            &course.ID,
            &course.TeacherID,
			&course.FirstName,
            &course.LastName,
            &course.CategoryID,
            &course.Name,
            &course.Slug,
            &course.Description,
			&image,
            &course.Price,
			&course.CreatedAt,
        )
        if err!= nil {
            return nil, err
        }
        courses = append(courses, course)
	}
	if err := rows.Err(); err!= nil {
        return nil, err
    }
	return courses, nil
}


func (s *Store) CountCourses() (int, error) {
	var count int

	query := `SELECT COUNT(*) FROM courses`
	err := s.db.QueryRow(query).Scan(&count)
	if err!= nil {
        return 0, err
    }
	return count, nil
}


// SECTION MANAGEMENT

func (s *Store) CreateSection(section *types.Section) error {
	query := `INSERT INTO sections (course_id, title, "order", created_at, modified_at)
			VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id`
	err := s.db.QueryRow(
		query,
		section.CourseID,
		section.Title,
		section.Order, 
	).Scan(&section.ID)
	if err != nil {
		return err
	}
	return nil
}

func (s *Store) GetSectionByID(id int) (*types.Section, error) {
	var section types.Section
	query := `SELECT id, course_id, title, "order", created_at FROM sections WHERE id = $1`

	err := s.db.QueryRow(query, id).Scan(
		&section.ID,
        &section.CourseID,
        &section.Title,
        &section.Order,
        &section.CreatedAt,
	)
	if err!= nil {
		if err == sql.ErrNoRows {
            return nil, fmt.Errorf("section not found for ID: %d", id)
        }
        return nil, err
	}
	return &section, nil
}


func (s *Store) UpdateSection(section *types.Section) error {
	query := `UPDATE sections SET course_id = $1, title = $2, "order" = $3, modified_at = NOW() WHERE id = $4`
    _, err := s.db.Exec(query, section.CourseID, section.Title, section.Order, section.ID)
    if err!= nil {
        return err
    }
    return nil
}


func (s *Store) DeleteSection(id int) error {
	query := `DELETE FROM sections WHERE id = $1;`
    _, err := s.db.Exec(query, id)
    if err!= nil {
        return err
    }
    return nil
}



// VIDEO MANAGEMENT

func (s *Store) CreateVideo(video *types.Video) error {
    query := `INSERT INTO videos (section_id, title, video_file, "order", created_at, modified_at)
            VALUES ($1, $2, $3, $4, NOW(), NOW()) RETURNING id`
    err := s.db.QueryRow(
        query,
        video.SectionID,
        video.Title,
        video.VideoFile,
        video.Order,
    ).Scan(&video.ID)
    if err != nil {
        return err
    }
    return nil
}


func (s *Store) GetVideoByID(id int) (*types.Video, error) {
	var video types.Video
    query := `SELECT id, section_id, title, video_file, "order", created_at FROM videos WHERE id = $1`

    err := s.db.QueryRow(query, id).Scan(
        &video.ID,
        &video.SectionID,
        &video.Title,
        &video.VideoFile,
        &video.Order,
        &video.CreatedAt,
    )
    if err!= nil {
        if err == sql.ErrNoRows {
            return nil, fmt.Errorf("video not found for ID: %d", id)
        }
        return nil, err
    }
    return &video, nil
}


func (s *Store) UpdateVideo(video *types.Video) error {
	query := `UPDATE videos SET section_id = $1, title = $2, video_file = $3, "order" = $4, modified_at = NOW() WHERE id = $5`
    _, err := s.db.Exec(query, video.SectionID, video.Title, video.VideoFile, video.Order, video.ID)
    if err!= nil {
        return err
    }
    return nil
}

func (s *Store) DeleteVideo(id int) error {
	query := `DELETE FROM videos WHERE id = $1;`
    _, err := s.db.Exec(query, id)
    if err!= nil {
        return err
    }
    return nil
}



// Course Builder Management

func (s *Store) GetCoursesByCategory(categoryID int, teacherID int) ([]types.Course, error) {
	query := `
	SELECT id, teacher_id, category_id, name, slug, description, image, price, created_at
	FROM courses WHERE category_id = $1 AND teacher_id = $2
	ORDER BY created_at DESC
	`

	rows, err := s.db.Query(query, categoryID, teacherID)
	if err!= nil {
        return nil, err
    }
	defer rows.Close()

	var courses []types.Course
	for rows.Next() {
		var course types.Course
        var image sql.NullString
        err := rows.Scan(
            &course.ID,
            &course.TeacherID,
			&course.CategoryID,
            &course.Name,
            &course.Slug,
            &course.Description,
            &image,
            &course.Price,
            &course.CreatedAt,
        )
        if err!= nil {
            return nil, err
        }
        courses = append(courses, course)
	}
	return courses, nil
}


func (s *Store) GetSectionsByCourse(courseID int, teacherID int) ([]types.Section, error) {
	query := `
	SELECT s.id, s.course_id, s.title, s.order 
	FROM sections s
	JOIN courses c ON s.course_id = c.id
	WHERE s.course_id = $1 AND c.teacher_id = $2
	`
	rows, err := s.db.Query(query, courseID, teacherID)
	if err!= nil {
        return nil, err
    }
	defer rows.Close()

	var sections []types.Section
	for rows.Next() {
		var section types.Section
        err := rows.Scan(
            &section.ID,
            &section.CourseID,
            &section.Title,
			&section.Order,
        )
        if err!= nil {
            return nil, err
        }
        sections = append(sections, section)
	}
	return sections, nil
}


func (s *Store) GetVideosBySection(sectionID int, teacherID int) ([]types.Video, error) {
	query := `
	SELECT v.id, v.section_id, v.title, v.video_file, v.order
		FROM videos v
		JOIN sections s ON v.section_id = s.id
		JOIN courses c ON s.course_id = c.id
		WHERE v.section_id = $1 AND c.teacher_id = $2
	`

	rows, err := s.db.Query(query, sectionID, teacherID)
	if err!= nil {
        return nil, err
    }
	defer rows.Close()

	var videos []types.Video
	for rows.Next() {
		var video types.Video
        err := rows.Scan(
            &video.ID,
            &video.SectionID,
            &video.Title,
            &video.VideoFile,
			&video.Order,
        )
        if err!= nil {
            return nil, err
        }
        videos = append(videos, video)
	}
	return videos, nil
}



func (s *Store) CountEnrolledStudents(teacherID int) (int, error) {
	query := `
		SELECT COUNT(DISTINCT student_id)
		FROM enrollments AS e
		JOIN courses AS c ON e.course_id = c.id
		WHERE c.teacher_id = $1`
	
	var totalStudents int
	err := s.db.QueryRow(query, teacherID).Scan(&totalStudents)
	if err != nil {
		return 0, fmt.Errorf("could not count enrolled students for teacher: %v", err)
	}

	return totalStudents, nil
}


func (s *Store) CountCoursesByTeacher(teacherID int) (int, error) {
	query := `
		SELECT COUNT(*)
		FROM courses
		WHERE teacher_id = $1`

	var totalCourses int
	err := s.db.QueryRow(query, teacherID).Scan(&totalCourses)
	if err != nil {
		return 0, fmt.Errorf("could not count courses for teacher: %v", err)
	}

	return totalCourses, nil
}



func (s *Store) GetTeacherProfileByID(teacherID int) (map[string]interface{}, error) {
	query := `
		SELECT t.id, u.first_name, u.last_name, t.bio, t.profession, up.profile_picture,
		up.country
		FROM teachers AS t
		JOIN users AS u ON t.user_id = u.id
		JOIN user_profiles AS up ON t.user_id = up.user_id
		WHERE t.id = $1`
	
	profile := make(map[string]interface{})

	row := s.db.QueryRow(query, teacherID)

	var id int
	var firstName, lastName, bio, profession, profilePicture, country string
	err := row.Scan(
		&id, &firstName, &lastName, &bio, &profession, &profilePicture, &country,
	)
	if err!= nil {
        return nil, fmt.Errorf("could not get teacher profile by ID: %v", err)
    }
	profile["id"] = id
	profile["first_name"] = firstName
	profile["last_name"] = lastName
	profile["bio"] = bio
	profile["profession"] = profession
	profile["profile_picture"] = profilePicture
	profile["country"] = country
	return profile, nil
}



func (s *Store) CountRatingsAndReviewsForTeacher(teacherID int) (int, int, error) {
	query := `
		SELECT 
			COUNT(r.rating) AS total_ratings,
			COUNT(CASE WHEN r.review IS NOT NULL AND r.review <> '' THEN 1 END) AS total_reviews
		FROM ratings AS r
		JOIN courses AS c ON r.course_id = c.id
		WHERE c.teacher_id = $1`

	var totalRatings, totalReviews int

	err := s.db.QueryRow(query, teacherID).Scan(&totalRatings, &totalReviews)
	if err!= nil {
        return 0, 0, fmt.Errorf("could not count ratings and reviews for teacher: %v", err)
    }

	return totalRatings, totalReviews, nil
}