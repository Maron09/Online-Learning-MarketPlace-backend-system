package utils

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/smtp"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)


var Validate = validator.New()

func ParseJSON(r *http.Request, payload any) error {
	if r.Body == nil {
		return fmt.Errorf("missing request body")
	}

	return json.NewDecoder(r.Body).Decode(payload)	
}



func WriteJSON(w http.ResponseWriter, status int, v any) error {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	return json.NewEncoder(w).Encode(v)
}


func WriteError(w http.ResponseWriter, status int, err error){
	WriteJSON(w, status, map[string]string{"err": err.Error()})
}


func SendEmail(to string, subject string, body string) error {
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	from := os.Getenv("EMAIL_HOST_USER")
	password := os.Getenv("EMAIL_HOST_PASSWORD")
	emailHost := os.Getenv("EMAIL_HOST")
	port := os.Getenv("EMAIL_PORT")

	// Create the message
	msg := []byte(fmt.Sprintf("Subject: %s\n\n%s", subject, body))


	hostPort := fmt.Sprintf("%s:%s", emailHost, port)

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
		ServerName:       emailHost,
	}

	// Establish TLS connection
	conn, err := tls.Dial("tcp", hostPort, tlsConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to SMTP server: %v", err)
	}
	defer conn.Close()

	// Create a new SMTP client over the secure connection
	client, err := smtp.NewClient(conn, emailHost)
	if err != nil {
		return fmt.Errorf("failed to create SMTP client: %v", err)
	}

	// Authenticate with the server
	auth := smtp.PlainAuth("", from, password, emailHost)
	if err = client.Auth(auth); err != nil {
		return fmt.Errorf("failed to authenticate with SMTP server: %v", err)
	}

	// Set the sender and recipient
	if err = client.Mail(from); err != nil {
		return fmt.Errorf("failed to set sender: %v", err)
	}
	if err = client.Rcpt(to); err != nil {
		return fmt.Errorf("failed to set recipient: %v", err)
	}

	// Send the email body
	writer, err := client.Data()
	if err != nil {
		return fmt.Errorf("failed to send email data: %v", err)
	}
	_, err = writer.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write email body: %v", err)
	}
	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close email writer: %v", err)
	}

	// Close the SMTP client
	client.Quit()

	return nil
}


// generate token for password reset
func GenerateTOken() string {
	b := make([]byte, 32)
    _, err := rand.Read(b)
    if err!= nil {
        log.Fatal(err)
    }
    return base64.URLEncoding.EncodeToString(b)
}


func Slugify(name string, id int) string {
	slug := strings.ToLower(name)
	slug = strings.ReplaceAll(slug, " ", "-")
	reg := regexp.MustCompile("[^a-z0-9-]+")
	slug = reg.ReplaceAllString(slug, "")

	slug = fmt.Sprintf("%s-%d", slug, id)

	return slug
} 


func GenerateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().UnixNano())
}