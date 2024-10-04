package types

import (
	"database/sql"
	"time"
)


type UserStore interface {
	// CreateUser(*User) error
	// CreateUserProfile(UserProfile) error
	GetUserByID(int) (User, error)
	UpdateUser(user User) error
	GetUserProfile(UserID int) (UserProfile, error)
	GetUserByEmail(email string) (User, error)
	UpdateUserOTP(user *User) error
	GetUserByEmailForLogin(email string) (*User, error)
	UpdateLastLogin(userID int64) error

	CreatePasswordResetToken(userID int64, token string, expiresAt time.Time) error
	UpdatePassword(userID int64, hashedPassword string) error
	DeletePasswordResetToken(token string) error
	GetPasswordResetToken(token string) (*PasswordResetToken, error)


	// New transaction-based methods
    CreateUserWithTransaction(tx *sql.Tx, user *User) error
    CreateUserProfileWithTransaction(tx *sql.Tx, profile *UserProfile) error
    CreateTeacherUserProfileWithTransaction(tx *sql.Tx, profile *UserProfile) error

	CreateTeacherWithTransaction(tx *sql.Tx, user *Teacher) error
	GetTeacherByID(id int) (*Teacher, error)
	ApproveTeacher(teacherID int) error
}



type TeacherStore interface {
	// CreateTeacher(teacher *Teacher) error
	
}



type UserRole int

const (
	TEACHER UserRole = 1
	STUDENT UserRole = 2
	ADMIN   UserRole = 3
)

type User struct {
	ID         int       `json:"id"`
	FirstName  string    `json:"first_name"`
	LastName   string    `json:"last_name"`
	Email      string    `json:"email"`
	Role       UserRole  `json:"role"`
	Password   string    `json:"password"`
	Created_at time.Time `json:"created_at"`
	Last_login *time.Time `json:"last_login"`
	Is_active  bool      `json:"is_active"`
	OTP        string    `json:"otp"`
	OTPExpiresAt time.Time `json:"otp_expires_at"`
}


type UserProfile struct {
	ID     int `json:"id"`
	UserID int `json:"user_id"`
	ProfilePicture string `json:"profile_picture"`
	Country string `json:"country"`
	CreatedAt time.Time `json:"created_at"`
	ModifiedAt time.Time `json:"modified_at"`
}

type Teacher struct {
	ID 		int `json:"id"`
	UserID  int `json:"user_id"`
	UserProfileID  int `json:"user_profile_id"`
	Bio            string    `json:"bio,omitempty"`
	Profession     string    `json:"Profession,omitempty"`
	Certificate    string    `json:"Certificate,omitempty"`
	IsApproved    bool      `json:"is_approved"`
	CreatedAt    time.Time `json:"created_at"`
	ModifiedAt   time.Time `json:"modified_at"`
}


type RegisterUserPayload struct {
	FirstName  string `json:"first_name" validate:"required"`
	LastName   string `json:"last_name" validate:"required"`
	Email      string `json:"email" validate:"required,email"`
	Password   string `json:"password" validate:"required,min=3,max=130"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}



type RegisterTeacherPayload struct {
	RegisterUserPayload
	Profession          string `json:"profession,omitempty"`
	Certificate         string `json:"certificate,omitempty"`
}



type VerifyOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
	OTP   string `json:"otp" validate:"required,len=6"`
}


type RegenerateOTPPayload struct {
	Email string `json:"email" validate:"required,email"`
}

type LoginUserPayload struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=3,max=130"`
}

type PasswordResetToken struct {
    Token     string    `json:"token"`
    UserID    int       `json:"user_id"`
    ExpiresAt time.Time `json:"expires_at"`
}

type ForgotPasswordPayload struct {
    Email string `json:"email" validate:"required,email"`
}


type ResetPasswordPayload struct {
    Token          string `json:"token" validate:"required"`
    NewPassword    string `json:"new_password" validate:"required,min=3,max=130"`
    ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=NewPassword"`
}


type RegisterAdminPayload struct {
	FirstName string `json:"first_name" validate:"required"`
    LastName  string `json:"last_name" validate:"required"`
    Email     string `json:"email" validate:"required,email"`
    Password  string `json:"password" validate:"required,min=3,max=130"`
	ConfirmPassword string `json:"confirm_password" validate:"required,eqfield=Password"`
}

type ApproveTeacherPayload struct {
	TeacherID int `json:"teacher_id" validate:"required"`
}