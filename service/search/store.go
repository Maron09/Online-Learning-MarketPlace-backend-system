package search

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/lib/pq"
)

type Store struct {
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}


func (s *Store) GetCategoryIDByName(name string) (int, error) {
	var categoryID int

	query := `SELECT id FROM categories WHERE TRIM(name) ILIKE $1`

	err := s.db.QueryRow(query, name).Scan(&categoryID)
	if err != nil {
        if err == sql.ErrNoRows {
            fmt.Println("No category found for:", name)  // Log category search failure
            return 0, fmt.Errorf("category not found")
        }
        fmt.Println("Error querying category:", err)  // Log any query errors
        return 0, err
    }


	return categoryID, nil
}


func (s *Store) SearchCourses(searchTerm string, categories []int, teacherID int, priceMin, priceMax float64, createdAfter, createdBefore string, page, limit int) ([]map[string]interface{}, int, error) {
	var courses []map[string]interface{}
	offset := (page - 1) * limit

	query := `
		SELECT c.id, c.teacher_id, c.category_id, c.name, c.slug, c.description, c.intro_video, c.image, c.price, c.created_at, c.modified_at,
		       	 t.id, u.first_name, u.last_name
		FROM courses c
		JOIN teachers t ON c.teacher_id = t.id
		JOIN users u ON t.user_id = u.id
		WHERE 1=1`

	var args []interface{}
	argIndex := 1

	// Search by name or description
	if searchTerm != "" {
		query += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.description ILIKE $%d)`, argIndex, argIndex)
		args = append(args, "%"+searchTerm+"%")
		argIndex++
	}

	// Filter by teacher ID if provided
	if teacherID != 0 {
		query += fmt.Sprintf(` AND c.teacher_id = $%d`, argIndex)
		args = append(args, teacherID)
		argIndex++
	}

	// Filter by categories if provided
	if len(categories) > 0 {
		query += fmt.Sprintf(` AND c.category_id = ANY($%d)`, argIndex)
		args = append(args, pq.Array(categories))
		argIndex++
	}

	// Filter by price range if provided
	if priceMin > 0 && priceMax > 0 {
		query += fmt.Sprintf(` AND c.price BETWEEN $%d AND $%d`, argIndex, argIndex+1)
		args = append(args, priceMin, priceMax)
		argIndex += 2
	}

	// Filter by creation date range if provided
	if createdAfter != "" && createdBefore != "" {
		query += fmt.Sprintf(` AND c.created_at BETWEEN $%d AND $%d`, argIndex, argIndex+1)
		args = append(args, createdAfter, createdBefore)
		argIndex += 2
	}

	// Add pagination
	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Get total number of courses (for pagination metadata)
	totalQuery := `SELECT COUNT(*) FROM courses c WHERE 1=1`
	var totalArgs []interface{}
	totalArgIndex := 1

	// Search by name or description for total count
	if searchTerm != "" {
		totalQuery += fmt.Sprintf(` AND (c.name ILIKE $%d OR c.description ILIKE $%d)`, totalArgIndex, totalArgIndex)
		totalArgs = append(totalArgs, "%"+searchTerm+"%")
		totalArgIndex++
	}

	// Filter by teacher ID for total count
	if teacherID != 0 {
		totalQuery += fmt.Sprintf(` AND c.teacher_id = $%d`, totalArgIndex)
		totalArgs = append(totalArgs, teacherID)
		totalArgIndex++
	}

	// Filter by categories for total count
	if len(categories) > 0 {
		totalQuery += fmt.Sprintf(` AND c.category_id = ANY($%d)`, totalArgIndex)
		totalArgs = append(totalArgs, pq.Array(categories))
		totalArgIndex++
	}

	// Filter by price range for total count
	if priceMin > 0 && priceMax > 0 {
		totalQuery += fmt.Sprintf(` AND c.price BETWEEN $%d AND $%d`, totalArgIndex, totalArgIndex+1)
		totalArgs = append(totalArgs, priceMin, priceMax)
		totalArgIndex += 2
	}

	// Filter by creation date range for total count
	if createdAfter != "" && createdBefore != "" {
		totalQuery += fmt.Sprintf(` AND c.created_at BETWEEN $%d AND $%d`, totalArgIndex, totalArgIndex+1)
		totalArgs = append(totalArgs, createdAfter, createdBefore)
		totalArgIndex += 2
	}

	var totalCourses int
	err := s.db.QueryRow(totalQuery, totalArgs...).Scan(&totalCourses)
	if err != nil {
		return nil, 0, err
	}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var courseID, teacherID, categoryID int
		var name, slug, description, firstName, lastName string
		var price float64
		var createdAt, modifiedAt time.Time
		var introVideo, image sql.NullString
	
		
		err := rows.Scan(
			&courseID,
			&teacherID,
			&categoryID,
			&name,
			&slug,
			&description,
			&introVideo,
			&image,
			&price,
			&createdAt,
			&modifiedAt,
			&teacherID,
			&firstName,
			&lastName,
		)
		if err != nil {
			return nil, 0, err
		}
	
		// Add the course with the instructor details to the response
		course := map[string]interface{}{
			"id":          courseID,
			"teacher_id":  teacherID,
			"category_id": categoryID,
			"name":        name,
			"slug":        slug,
			"description": description,
			"intro_video": introVideo,
			"image":       image,
			"price":       price,
			"created_at":  createdAt,
			"modified_at": modifiedAt,
			"instructor":  fmt.Sprintf("%s %s", firstName, lastName), // Combine first name and last name
		}
	
		courses = append(courses, course)
	}	

	return courses, totalCourses, nil
}
