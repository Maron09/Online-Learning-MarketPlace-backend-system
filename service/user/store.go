package user

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/sikozonpc/ecom/types"
)


type Store struct {
    db *sql.DB
}



func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}





// func (s *Store) CreateUser(user *types.User) error {
// 	query := `
// 		INSERT INTO users (first_name, last_name, email, password, role, otp, otp_expires_at)
// 		VALUES ($1, $2, $3, $4, $5, $6, $7)
// 		RETURNING id`
	
// 	// Use QueryRow to get the ID of the newly created user
// 	err := s.db.QueryRow(query, user.FirstName, user.LastName, user.Email, user.Password, user.Role).Scan(&user.ID)
// 	if err != nil {
// 		return fmt.Errorf("failed to create user: %w", err)
// 	}
// 	return nil
// }



// func (s *Store) CreateUserProfile(userProfile types.UserProfile) error {
//     query := "INSERT INTO user_profiles (user_id, profile_picture, country, created_at, modified_at) VALUES ($1, $2, $3, NOW(), NOW()) RETURNING id"
//     err := s.db.QueryRow(query, userProfile.UserID, userProfile.ProfilePicture, userProfile.Country).Scan(&userProfile.ID)
//     if err != nil {
//         return err
//     }
//     return nil
// }



func (s *Store) GetUserByID(id int) (types.User, error) {
	var user types.User

	query := "SELECT id, first_name, last_name, email, role, created_at, last_login, is_active FROM users WHERE id=$1"

	row := s.db.QueryRow(query, id)
	err := row.Scan(
		&user.ID, 
		&user.FirstName , 
		&user.LastName , 
		&user.Email, 
		&user.Role, 
		&user.Created_at, 
		&user.Last_login, 
		&user.Is_active,
	)

	if err == sql.ErrNoRows {
		return types.User{}, fmt.Errorf("user with id %d not found", id)
	} else if err != nil {
		return types.User{}, err
	}
	return user, nil
}

func (s *Store) GetUserProfile(userID int) (types.UserProfile, error){
	var userProfile types.UserProfile

    query := "SELECT id, user_id, profile_picture, country, created_at FROM user_profiles WHERE user_id=$1"

    row := s.db.QueryRow(query, userID)
    err := row.Scan(
        &userProfile.ID, 
        &userProfile.UserID, 
        &userProfile.ProfilePicture, 
        &userProfile.Country, 
        &userProfile.CreatedAt,
    )

	if err == sql.ErrNoRows {
		return types.UserProfile{}, fmt.Errorf("user with userID %d not found", userID)
	} else if err!= nil {
        return types.UserProfile{}, err
    }
    return userProfile, nil
}


func (s *Store) GetUserByEmail(email string) (types.User, error) {
    var user types.User
    query := "SELECT id, first_name, last_name, email, role, created_at, last_login, is_active, otp, otp_expires_at FROM users WHERE LOWER(email) = LOWER($1)"

    row := s.db.QueryRow(query, email)
    err := row.Scan(&user.ID, &user.FirstName, &user.LastName, &user.Email, &user.Role, &user.Created_at, &user.Last_login, &user.Is_active, &user.OTP, &user.OTPExpiresAt)

    if err != nil {
        if err == sql.ErrNoRows {
            return types.User{}, fmt.Errorf("user with email %s not found", email)
        }
        log.Printf("Error scanning user: %v", err) // Log scanning error
        return types.User{}, err
    }

    return user, nil
}



func (s *Store) GetUserByEmailForLogin(email string) (*types.User, error) {
	var user types.User
	query := `SELECT id, email, password, last_login FROM users WHERE email = $1`
	err := s.db.QueryRow(query, email).Scan(&user.ID, &user.Email, &user.Password, &user.Last_login)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	return &user, nil
}







// CreateUserWithTransaction creates a user within a transaction
func (s *Store) CreateUserWithTransaction(tx *sql.Tx, user *types.User) error {
    query := `INSERT INTO users (first_name, last_name, email, password, role, otp, otp_expires_at) 
                VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
    err := tx.QueryRow(query, user.FirstName, user.LastName, user.Email, user.Password, user.Role, user.OTP, user.OTPExpiresAt).Scan(&user.ID)
    if err != nil {
        return fmt.Errorf("failed to create user: %v", err)
    }
    return nil
}

// CreateUserProfileWithTransaction creates a user profile within a transaction
func (s *Store) CreateUserProfileWithTransaction(tx *sql.Tx, profile types.UserProfile) error {
    query := `INSERT INTO user_profiles (user_id, profile_picture, country, created_at, modified_at) 
            VALUES ($1, $2, $3, $4, $5)`
    _, err := tx.Exec(query, profile.UserID, profile.ProfilePicture, profile.Country, profile.CreatedAt, profile.ModifiedAt)
    if err != nil {
        return fmt.Errorf("failed to create user profile: %v", err)
    }
    return nil
}


func (s *Store) UpdateUser(user types.User) error {
	query := `UPDATE users SET is_active = $1, otp = NULL, otp_expires_at = NULL WHERE id = $2`

	_, err := s.db.Exec(query, user.Is_active, user.ID)
	return err
}


func (s *Store) UpdateUserOTP(user *types.User) error {
	query := "UPDATE users SET otp = $1, otp_expires_at = $2 WHERE email = $3"

	_, err := s.db.Exec(query, user.OTP, user.OTPExpiresAt, user.Email)

	return err
}


func (s *Store) UpdateLastLogin(userID int64) error {
	query := `UPDATE users SET last_login = NOW() WHERE id = $1`
    _, err := s.db.Exec(query, userID)
    return err
}


// Password reset
func (s *Store) CreatePasswordResetToken(userID int64, token string, expiresAt time.Time) error {
	query := "INSERT INTO password_resets (user_id, token, expires_at) VALUES ($1, $2, $3)"
    _, err := s.db.Exec(query, userID, token, expiresAt)
    if err!= nil {
        log.Printf("Error creating password reset token: %v", err)
    }
	return nil
}

func (s *Store) GetPasswordResetToken(token string) (*types.PasswordResetToken, error) {
	var resetToken types.PasswordResetToken
    query := `SELECT token, user_id, expires_at FROM password_resets WHERE token = $1`
    err := s.db.QueryRow(query, token).Scan(&resetToken.Token, &resetToken.UserID, &resetToken.ExpiresAt)
    if err != nil {
        return nil, err
    }
    return &resetToken, nil
}


func (s *Store) UpdatePassword(userID int64, hashedPassword string) error{
	query := `UPDATE users SET password = $1 WHERE id = $2`
    _, err := s.db.Exec(query, hashedPassword, userID)
    return err
}


func (s *Store) DeletePasswordResetToken(token string) error{
	query := `DELETE FROM password_resets WHERE token = $1`
    _, err := s.db.Exec(query, token)
    return err
}