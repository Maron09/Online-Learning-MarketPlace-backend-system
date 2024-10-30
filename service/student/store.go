package student

import (
	"database/sql"
	"fmt"


)

type Store struct {
	db *sql.DB
}


func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}


func (s *Store) GetEnrolledCourses(studentID int) ([]map[string]interface{}, error) {
	var courses []map[string]interface{}

	query := `
	SELECT c.id, c.teacher_id, c.name, c.slug, 
	       c.description, c.image, 
	       u.first_name, u.last_name
	FROM enrollments e
	JOIN courses c ON e.course_id = c.id
	JOIN teachers t ON c.teacher_id = t.id
	JOIN users u ON t.user_id = u.id
	WHERE e.student_id = $1`

	rows, err := s.db.Query(query, studentID)
	if err != nil {
		return nil, fmt.Errorf("could not fetch enrolled courses: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		var (
			courseID    int
			teacherID   int
			name        string
			slug        string
			description string
			image       sql.NullString
			firstName   string
			lastName    string
		)

		if err := rows.Scan(
			&courseID,
			&teacherID,
			&name,
			&slug,
			&description,
			&image,
			&firstName,
			&lastName,
		); err != nil {
			return nil, fmt.Errorf("could not scan course data: %v", err)
		}

		courseData := map[string]interface{}{
			"id":          courseID,
			"teacher_id":  teacherID,
			"name":        name,
			"slug":        slug,
			"description": description,
			"image":       image.String,
			"instructor":  fmt.Sprintf("%s %s", firstName, lastName),
		}

		courses = append(courses, courseData)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	return courses, nil
}



func (s *Store) GetEnrolledCoursesBySlug(studentID int, slug string) (map[string]interface{}, error) {
	courseData := make(map[string]interface{})

	query := `
	SELECT c.id, c.teacher_id, c.name, c.slug, 
	       c.description, c.image, c.for_who, c.reason,
	       u.first_name, u.last_name
	FROM enrollments e
	JOIN courses c ON e.course_id = c.id
	JOIN teachers t ON c.teacher_id = t.id
	JOIN users u ON t.user_id = u.id
	WHERE e.student_id = $1 AND c.slug = $2`

	rows, err := s.db.Query(query, studentID, slug)
	if err != nil {
		return nil, fmt.Errorf("could not fetch enrolled course by slug: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		var (
			courseID    int
			teacherID   int
			name        string
			courseSlug  string
			description string
			forWho      sql.NullString
			reason      sql.NullString
			image       sql.NullString
			firstName   string
			lastName    string
		)

		if err := rows.Scan(
			&courseID,
			&teacherID,
			&name,
			&courseSlug,
			&description,
			&image,
			&forWho,
			&reason,
			&firstName,
			&lastName,
		); err != nil {
			return nil, fmt.Errorf("could not scan course data: %v", err)
		}

		courseData = map[string]interface{}{
			"id":         courseID,
			"teacher_id": teacherID,
			"name":       name,
			"slug":       courseSlug,
			"About": map[string]string{
				"description": description,
				"for_who":    forWho.String,
				"reason":     reason.String,
			},
			"image":      image.String,
			"instructor": fmt.Sprintf("%s %s", firstName, lastName),
		}
	} else {
		return nil, fmt.Errorf("no course found for student %d with slug %s", studentID, slug)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %v", err)
	}

	// Ensure courseID is properly retrieved from courseData
	courseID := courseData["id"].(int) 
	sections, err := s.GetEnrolledCourseSectionsAndVideos(courseID) 
	if err != nil {
		return nil, err
	}

	courseData["sections"] = sections

	return courseData, nil
}



func (s *Store) GetEnrolledCourseSectionsAndVideos(courseID int) ([]map[string]interface{}, error) {
	var sections []map[string]interface{}

	query := `
	SELECT s.id, s.title, v.id, v.title, COALESCE(v.video_file, '') as video_url
	FROM sections s
	LEFT JOIN videos v ON s.id = v.section_id
	WHERE s.course_id = $1
	`

	rows, err := s.db.Query(query, courseID)
	if err!= nil {
        return nil, fmt.Errorf("failed to query sections and videos: %v", err)
    }
	defer rows.Close()

	sectionMap := make(map[int]map[string]interface{})

	for rows.Next() {
		var sectionID int
        var title string
        var videoID int
        var videoTitle string
        var videoURL sql.NullString

		err := rows.Scan(&sectionID, &title, &videoID, &videoTitle, &videoURL)
        if err!= nil {
            return nil, fmt.Errorf("failed to scan row: %v", err)
        }

		if _, exists := sectionMap[sectionID]; !exists {
			sectionMap[sectionID] = map[string]interface{}{
				"id":   sectionID,
                "title": title,
                "videos": []map[string]interface{}{},
			}
		}

		if videoURL.Valid {
            sectionMap[sectionID]["videos"] = append(sectionMap[sectionID]["videos"].([]map[string]interface{}), map[string]interface{}{
                "id": videoID,
                "title": videoTitle,
                "file": videoURL.String,
            })
        }
	}

	for _, section := range sectionMap {
		sections = append(sections, section)
	}

	return sections, nil
}