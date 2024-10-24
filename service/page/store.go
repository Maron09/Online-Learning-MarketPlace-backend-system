package page

import (
	"database/sql"
	"fmt"
	"time"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
    return &Store{db: db}
}


func (s *Store) GetCourseSectionsAndVideos(courseID int) ([]map[string]interface{}, error) {
	var sections []map[string]interface{}

	query := `
	SELECT s.id, s.title, COALESCE(v.video_file, '') as video_url
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
        var videoURL sql.NullString

		err := rows.Scan(&sectionID, &title, &videoURL)
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
                "url": videoURL.String,
            })
        }
	}

	for _, section := range sectionMap {
		sections = append(sections, section)
	}

	return sections, nil
}



func (s *Store) GetCourseDetailBySlug(slug string) (map[string]interface{}, error) {
	courseDetail := make(map[string]interface{})

	query := `
	SELECT c.id, c.name, c.price, c.modified_at, u.first_name, u.last_name
	FROM courses c
	JOIN teachers t ON c.teacher_id = t.id
	JOIN users u ON t.user_id = u.id
	WHERE c.slug = $1
	`

	var courseID int
	var name string
	var price float64
	var modifiedAt time.Time
	var firstName string
	var lastName string

	err := s.db.QueryRow(query, slug).Scan(
		&courseID,
        &name,
        &price,
        &modifiedAt,
        &firstName,
        &lastName,
	)
	if err!= nil {
		if err == sql.ErrNoRows {
            return nil, fmt.Errorf("course not found: %v", err)
        } else {
            return nil, err
        }
	}

	courseDetail["name"] = name
	courseDetail["price"] = price
	courseDetail["modified_at"] = modifiedAt.Format("2006-01-02")
	courseDetail["teacher"] = map[string]string{
		"first_name": firstName,
        "last_name":  lastName,
	}

	sections, err := s.GetCourseSectionsAndVideos(courseID)
	if err != nil {
		return nil, err
	}

	courseDetail["sections"] = sections

	return courseDetail, nil

}
