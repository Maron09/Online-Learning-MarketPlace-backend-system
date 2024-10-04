package user

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/auth"
	"github.com/sikozonpc/ecom/types"
	"github.com/sikozonpc/ecom/utils"
)

type Handler struct {
	store types.UserStore
	db *sql.DB
}


func NewHandler(store types.UserStore, db *sql.DB) *Handler {
	return &Handler{
		store: store,
		db: db,
	}
}

func (h *Handler) AuthRoutes(router *mux.Router){
	router.HandleFunc("/register", h.registerHandler).Methods(http.MethodPost)
	router.HandleFunc("/verify_otp", h.verifyHandler).Methods(http.MethodPost)
	router.HandleFunc("/regenerate_otp", h.regenerateOTPHandler).Methods(http.MethodPost)
	router.HandleFunc("/login", h.loginHandler).Methods(http.MethodPost)
	router.HandleFunc("/forgot_password", h.forgotPasswordHandler).Methods(http.MethodPost)
	router.HandleFunc("/reset_password", h.resetPasswordHandler).Methods(http.MethodPost)
}


func (h *Handler) registerHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.RegisterUserPayload

	// Parse the JSON payload
	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}

	

	// Check if the user already exists
	_, err := h.store.GetUserByEmail(payload.Email)
	if err == nil {
		utils.WriteError(writer, http.StatusConflict, fmt.Errorf("user with email %s already registered", payload.Email))
		return
	}

	// Start a database transaction
    tx, err := h.db.Begin()
    if err != nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to start transaction: %v", err))
        return
    }

	// Hash the password
	hashedPassword, err := auth.HashedPassword(payload.Password)
	if err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
		return
	}

	//Generate otp
	otp := auth.GenerateOTP()
	otpExpiration := time.Now().Add(30 * time.Minute)

	// Create a new user
	user := &types.User{
		FirstName: payload.FirstName,
		LastName:  payload.LastName,
		Email:     payload.Email,
		Password:  hashedPassword,
		Role:      2, 
		OTP:          otp,
		OTPExpiresAt: otpExpiration,
	}

	// Attempt to create the user in the database
	err = h.store.CreateUserWithTransaction(tx, user) // Pass the pointer here
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create user: %v", err))
		_ = tx.Rollback()
		return
	}
	

	// Ensure the user ID is populated
	if user.ID == 0 {
		_ = tx.Rollback() // Rollback if user creation fails
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create user: ID not set"))
		return
	}


	// Create a user profile for the new user
	userProfile := types.UserProfile{
		UserID: user.ID, // Use the populated user ID
	}

	err = h.store.CreateUserProfileWithTransaction(tx, userProfile)
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create user profile: %v", err))
		_ = tx.Rollback()
		return
	}

	// Commit the transaction if everything went well
	err = tx.Commit()
	if err != nil {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to commit transaction: %v", err))
		return
	}


	//send the OTP
	err = utils.SendEmail(payload.Email, "Your Otp Code", fmt.Sprintf("Your OTP is: %s", otp))
	if err != nil {
		log.Printf("Error sending OTP: %v", err)
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to send OTP: %v", err))
		return
	}

	// Respond with success
	response := map[string]string{"message": "User registered successfully. Please check your email for the OTP."}
	utils.WriteJSON(writer, http.StatusCreated, response)
}



func (h *Handler) verifyHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.VerifyOTPPayload


	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
	}

	// Validate the payload
	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
		return
	}


	// Trim whitespace and convert to lowercase
	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	// Check if the user exists
	user, err := h.store.GetUserByEmail(payload.Email)
	if err != nil {
		utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("user not found with email %s", payload.Email))
		return
	}

	if user.OTP != payload.OTP{
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid OTP"))
        return
	}

	if time.Now().After(user.OTPExpiresAt){
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("OTP has expired"))
        return
	}

	user.Is_active = true
	err = h.store.UpdateUser(user)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to update user status: %v", err))
        return
    }

	response := map[string]string{"message": "User Activated successfully"}
	utils.WriteJSON(writer, http.StatusOK, response)
}



func (h *Handler) regenerateOTPHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.RegenerateOTPPayload


	if err := utils.ParseJSON(request, &payload); err != nil {
		utils.WriteError(writer, http.StatusBadRequest, err)
        return
	}

	if err := utils.Validate.Struct(payload); err != nil {
		errors := err.(validator.ValidationErrors)
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
        return
	}

	user, err := h.store.GetUserByEmail(payload.Email)

	if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("user not found with email %s", payload.Email))
        return
    }

	if time.Now().After(user.OTPExpiresAt) {
		otp := auth.GenerateOTP()
		otpExpiration := time.Now().Add(30 * time.Minute)


		user.OTP = otp
        user.OTPExpiresAt = otpExpiration

		if err := h.store.UpdateUserOTP(&user); err != nil {
			utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to update user OTP: %v", err))
            return
		}

		err = utils.SendEmail(payload.Email, "Your New OTP code", fmt.Sprintf("Your New OTP is: %s", otp))
		if err!= nil {
            log.Printf("Error sending OTP: %v", err)
            utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to send OTP: %v", err))
            return
        }


		response := map[string]string{"message": "OTP regenerated successfully. Please check your email for the new OTP."}
        utils.WriteJSON(writer, http.StatusOK, response)
	} else {
		utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("OTP has not expired yet"))
        return
	}
}


func (h *Handler) loginHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.LoginUserPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	if err := utils.Validate.Struct(payload); err!= nil {
        errors := err.(validator.ValidationErrors)
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
        return
    }

	payload.Email = strings.ToLower(strings.TrimSpace(payload.Email))

	user, err := h.store.GetUserByEmailForLogin(payload.Email)
	if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("user not found with email %s", payload.Email))
        return
    }
	if !auth.CheckPassword(user.Password, []byte(payload.Password)) {
		log.Printf("Failed login attempt for email: %s", payload.Email)
		utils.WriteError(writer, http.StatusUnauthorized, fmt.Errorf("invalid email or password"))
        return
	}


	err = h.store.UpdateLastLogin(int64(user.ID))
	if err != nil {
    log.Printf("Failed to update last login for user %d: %v", user.ID, err)
    utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to update last login"))
    return
}

	// Handling token for authentication
	secret := os.Getenv("JWTSecret")
	if secret == "" {
		utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("JWT secret is not set"))
		return
	}
	
	token, err := auth.CreateJWT([]byte(secret), user.ID)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create JWT: %v", err))
        return
    }
	response := map[string]string{"message": "Login successful", "token": token}

	utils.WriteJSON(writer, http.StatusOK, response)
}


func (h *Handler) forgotPasswordHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.ForgotPasswordPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }

	if err := utils.Validate.Struct(payload); err!= nil {
        errors := err.(validator.ValidationErrors)
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
        return
    }

	user, err := h.store.GetUserByEmailForLogin(strings.ToLower(strings.TrimSpace(payload.Email)))
	if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("user not found with email %s", payload.Email))
        return
    }

	resetToken := utils.GenerateTOken()
	expiresAt := time.Now().Add(30 * time.Minute)


	err = h.store.CreatePasswordResetToken(int64(user.ID), resetToken, expiresAt)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to create password reset token: %v", err))
        return
    }


	emailBody := fmt.Sprintf("Here is Your reset token link: http://localhost:8000/reset_password?token=%s", resetToken)
	err = utils.SendEmail(payload.Email, "Reset Password", emailBody)
	if err!= nil {
        log.Printf("Error sending reset password email: %v", err)
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to send reset password email: %v", err))
        return
    }

	response := map[string]string{"message": "Reset password email sent successfully. Please check your email for the link."}

	utils.WriteJSON(writer, http.StatusOK, response)
}


func (h *Handler) resetPasswordHandler(writer http.ResponseWriter, request *http.Request) {
	var payload types.ResetPasswordPayload

	if err := utils.ParseJSON(request, &payload); err!= nil {
        utils.WriteError(writer, http.StatusBadRequest, err)
        return
    }
	if err := utils.Validate.Struct(payload); err!= nil {
        errors := err.(validator.ValidationErrors)
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("invalid payload: %v", errors))
        return
    }

	resetToken, err := h.store.GetPasswordResetToken(payload.Token)
	if err!= nil {
        utils.WriteError(writer, http.StatusNotFound, fmt.Errorf("invalid or expired reset token"))
        return
    }

	if time.Now().After(resetToken.ExpiresAt) {
        utils.WriteError(writer, http.StatusBadRequest, fmt.Errorf("reset token has expired"))
        return
    }

	hashedPassword, err := auth.HashedPassword(payload.NewPassword)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to hash password: %v", err))
        return
    }

	err = h.store.UpdatePassword(int64(resetToken.UserID), hashedPassword)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to update password: %v", err))
        return
    }


	err = h.store.DeletePasswordResetToken(payload.Token)
	if err!= nil {
        utils.WriteError(writer, http.StatusInternalServerError, fmt.Errorf("failed to delete password reset token: %v", err))
        return
    }

	response := map[string]string{"message": "Password reset successful"}

	utils.WriteJSON(writer, http.StatusOK, response)
}