package teacher

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



func (s *Store) CreateTeacher(teacher *types.Teacher) error {
	query := `INSERT INTO teachers (user_id, user_profile_id, bio, profession, certificate, is_approved) 
			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	var id int
	err  := s.db.QueryRow(query, teacher.UserID, teacher.UserProfileID, teacher.Bio, teacher.Profession, teacher.Certificate, teacher.IsApproved).Scan(&id)
	if err != nil {
		return err
	}
	teacher.ID = id

	return nil
}


func (s *Store) GetTeacherByID(id int) (*types.Teacher, error) {
	var teacher types.Teacher

    query := `SELECT id, user_id, user_profile_id, bio, profession, certificate, is_approved FROM teachers WHERE id=$1`

    row := s.db.QueryRow(query, id)
    err := row.Scan(&teacher.ID, 
		&teacher.UserID, 
		&teacher.UserProfileID, 
		&teacher.Bio, 
		&teacher.Profession, 
		&teacher.Certificate, 
		&teacher.IsApproved,
	)

    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("teacher with id %d not found", id)
    } else if err!= nil {
        return nil, err
    }

    return &teacher, nil
}


func (s *Store) ApproveTeacher(teacherID int) error {
    query := "UPDATE teachers SET is_approved=true WHERE id=$1"
    result, err := s.db.Exec(query, teacherID)
    
    if err != nil {
        return fmt.Errorf("error approving teacher with ID %d: %v", teacherID, err)
    }

	rowsAffected, err := result.RowsAffected()
    if err != nil {
        return fmt.Errorf("error getting affected rows for teacher with ID %d: %v", teacherID, err)
    }

	if rowsAffected == 0 {
        return fmt.Errorf("no teacher found with ID %d", teacherID)
    }

    return nil

}
