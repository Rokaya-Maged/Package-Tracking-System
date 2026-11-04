package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/mail"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"golang.org/x/crypto/bcrypt"
	"tracking-system/config"
)

type RegisterRequest struct {
	FullName     string `json:"full_name"`
	BusinessName string `json:"business_name"`
	Email        string `json:"email"`
	Phone        string `json:"phone"`
	Password     string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {

	var user RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, "Invalid registration data", http.StatusBadRequest)
		return
	}

	user.FullName = strings.TrimSpace(user.FullName)
	user.BusinessName = strings.TrimSpace(user.BusinessName)
	user.Email = strings.ToLower(strings.TrimSpace(user.Email))
	user.Phone = strings.TrimSpace(user.Phone)

	if user.FullName == "" || user.Email == "" ||
		user.Phone == "" || user.Password == "" {
		http.Error(w, "All required fields must be filled", http.StatusBadRequest)
		return
	}

	email, err := mail.ParseAddress(user.Email)
	if err != nil || email.Address != user.Email {
		http.Error(w, "Invalid email address", http.StatusBadRequest)
		return
	}

	if len(user.Password) < 8 || len(user.Password) > 72 {
		http.Error(w, "Password must be between 8 and 72 bytes", http.StatusBadRequest)
		return
	}

	if len(user.FullName) > 100 || len(user.BusinessName) > 150 ||
		len(user.Email) > 255 || len(user.Phone) > 20 {
		http.Error(w, "Registration field is too long", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(
		[]byte(user.Password),
		bcrypt.DefaultCost,
	)

	if err != nil {
		http.Error(w, "Password processing failed", http.StatusInternalServerError)
		return
	}

	query := `
		INSERT INTO users
		(full_name, business_name, email, phone, password_hash, role)
		VALUES ($1, $2, $3, $4, $5, 'business')
		RETURNING id
	`

	var userID int64

	err = config.DB.QueryRow(
		r.Context(),
		query,
		user.FullName,
		user.BusinessName,
		user.Email,
		user.Phone,
		string(hashedPassword),
	).Scan(&userID)

	if err != nil {
		var dbError *pgconn.PgError

		if errors.As(err, &dbError) && dbError.Code == "23505" {
			http.Error(w, "Email already registered", http.StatusConflict)
			return
		}

		http.Error(w, "Registration failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Registration successful",
		"user_id": userID,
	})
}
