package auth

import "golang.org/x/crypto/bcrypt"

func HashedPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err!= nil {
        return "", err
    }

	return string(hash), nil
}


func CheckPassword(hashed string, plain []byte) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashed), plain)
    return err == nil
    // If err == bcrypt.ErrMismatchedHashAndPassword, then the password does not match
    // If err == nil, then the password matches
}