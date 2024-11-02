package rating

import (
	"database/sql"
	"fmt"

	"github.com/sikozonpc/ecom/types"
)

type Store struct {
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}


func (s *Store) IsStudentEnrolledInCourse(studentID int, courseID int) (bool, error) {
    var count int
    query := `
        SELECT COUNT(*)
        FROM enrollments
        WHERE student_id = $1 AND course_id = $2
    `
    err := s.db.QueryRow(query, studentID, courseID).Scan(&count)
    if err != nil {
        return false, fmt.Errorf("failed to check enrollment: %v", err)
    }
    return count > 0, nil
}


func (s *Store) CreateRating(rating *types.Rating) (int, error) {
	var ratingID int
	query := `
		INSERT INTO ratings (student_id, course_id, rating, review, created_at)
		VALUES ($1, $2, $3, $4, NOW())
		RETURNING id`
	err := s.db.QueryRow(query, rating.StudentID, rating.CourseID, rating.Rating, rating.Review).Scan(&ratingID)
	if err != nil {
		return 0, fmt.Errorf("could not create rating: %v", err)
	}
	return ratingID, nil
}

// UpdateRating updates an existing rating and/or review by a student for a specific course.
func (s *Store) UpdateRating(ratingID int, updatedRating *types.Rating) error {
	query := `
		UPDATE ratings
		SET rating = $1, review = $2
		WHERE id = $3`
	_, err := s.db.Exec(query, updatedRating.Rating, updatedRating.Review, ratingID)
	if err != nil {
		return fmt.Errorf("could not update rating: %v", err)
	}
	return nil
}

// GetRating retrieves a specific rating for a course by rating ID.
func (s *Store) GetRating(ratingID int) (*types.Rating, error) {
	var rating types.Rating
	query := `
		SELECT id, student_id, course_id, rating, review, created_at
		FROM ratings
		WHERE id = $1`
	err := s.db.QueryRow(query, ratingID).Scan(
		&rating.ID, &rating.StudentID, &rating.CourseID,
		&rating.Rating, &rating.Review, &rating.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating not found")
		}
		return nil, fmt.Errorf("could not get rating: %v", err)
	}
	return &rating, nil
}
func (s *Store) GetRatingByStudentID(ratingID int, studentID int) (*types.Rating, error) {
	var rating types.Rating
	query := `
		SELECT r.id, r.student_id, r.course_id, c.name, r.rating, r.review, r.created_at
		FROM ratings AS r
		JOIN courses AS c ON r.course_id = c.id
		WHERE r.id = $1 AND r.student_id = $2`
	err := s.db.QueryRow(query, ratingID, studentID).Scan(
		&rating.ID, &rating.StudentID, &rating.CourseID, &rating.CourseName,
		&rating.Rating, &rating.Review, &rating.CreatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("rating not found")
		}
		return nil, fmt.Errorf("could not get rating: %v", err)
	}
	return &rating, nil
}

// GetRatingsForCourse retrieves all ratings and reviews for a given course.
func (s *Store) GetRatingsForCourse(slug string, page int, limit int) ([]types.Rating, int, error) {
	var ratings []types.Rating

	// Calculate offset for pagination
	offset := (page - 1) * limit

	query := `
		SELECT r.id, r.student_id, u.first_name, u.last_name, r.course_id, c.name, r.rating, r.review, r.created_at
		FROM ratings AS r
		JOIN courses AS c ON r.course_id = c.id
		JOIN users AS u ON r.student_id = u.id
		WHERE c.slug = $1
		LIMIT $2 OFFSET $3`
	rows, err := s.db.Query(query, slug, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("could not get ratings for course: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rating types.Rating
		if err := rows.Scan(
			&rating.ID, &rating.StudentID, &rating.StudentFirstName, &rating.StudentLastName, &rating.CourseID,
			&rating.CourseName, &rating.Rating, &rating.Review, &rating.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("could not scan rating: %v", err)
		}
		ratings = append(ratings, rating)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("row iteration error: %v", err)
	}

	// Get total count of ratings for the course
	var totalRatings int
	countQuery := `SELECT COUNT(*) FROM ratings AS r JOIN courses AS c ON r.course_id = c.id WHERE c.slug = $1`
	err = s.db.QueryRow(countQuery, slug).Scan(&totalRatings)
	if err != nil {
		return nil, 0, fmt.Errorf("could not count ratings: %v", err)
	}

	return ratings, totalRatings, nil
}


// GetRatingsByStudent retrieves all ratings submitted by a specific student.
func (s *Store) GetRatingsByStudent(studentID int) ([]types.Rating, error) {
	var ratings []types.Rating
	query := `
		SELECT r.id, r.student_id, u.first_name, u.last_name, r.course_id, c.name, r.rating, r.review, r.created_at
		FROM ratings AS r
		JOIN courses AS c ON r.course_id = c.id
		JOIN users AS u on r.student_id = u.id
		WHERE r.student_id = $1`
	rows, err := s.db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("could not get ratings by student: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var rating types.Rating
		if err := rows.Scan(
			&rating.ID, &rating.StudentID, &rating.StudentFirstName, &rating.StudentLastName, &rating.CourseID,
			&rating.CourseName, &rating.Rating, &rating.Review, &rating.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("could not scan rating: %v", err)
		}
		ratings = append(ratings, rating)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}
	return ratings, nil
}


// DeleteRating removes a rating by its ID.
func (s *Store) DeleteRating(ratingID int) error {
	query := `
		DELETE FROM ratings
		WHERE id = $1`
	_, err := s.db.Exec(query, ratingID)
	if err != nil {
		return fmt.Errorf("could not delete rating: %v", err)
	}
	return nil
}

// GetAverageRating calculates the average rating for a specific course.
func (s *Store) GetAverageRating(courseID int) (float32, error) {
	var avgRating float32
	query := `
		SELECT COALESCE(AVG(rating), 0)
		FROM ratings
		WHERE course_id = $1`
	err := s.db.QueryRow(query, courseID).Scan(&avgRating)
	if err != nil {
		return 0, fmt.Errorf("could not calculate average rating: %v", err)
	}
	return avgRating, nil
}

// CountRatingsAndReviews returns the total count of ratings and reviews for a course.
func (s *Store) CountRatingsAndReviews(courseID int) (int, int, error) {
	var ratingCount, reviewCount int
	query := `
		SELECT COUNT(rating) AS rating_count,
		       COUNT(CASE WHEN review IS NOT NULL AND review <> '' THEN 1 END) AS review_count
		FROM ratings
		WHERE course_id = $1`
	err := s.db.QueryRow(query, courseID).Scan(&ratingCount, &reviewCount)
	if err != nil {
		return 0, 0, fmt.Errorf("could not count ratings and reviews: %v", err)
	}
	return ratingCount, reviewCount, nil
}