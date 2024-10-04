package teacher

import (
	"database/sql"
	
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}



// func (s *Store) CreateTeacher(teacher *types.Teacher) error {
// 	query := `INSERT INTO teachers (user_id, user_profile_id, bio, profession, certificate, is_approved) 
// 			VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

// 	var id int
// 	err  := s.db.QueryRow(query, teacher.UserID, teacher.UserProfileID, teacher.Bio, teacher.Profession, teacher.Certificate, teacher.IsApproved).Scan(&id)
// 	if err != nil {
// 		return err
// 	}
// 	teacher.ID = id

// 	return nil
// }

