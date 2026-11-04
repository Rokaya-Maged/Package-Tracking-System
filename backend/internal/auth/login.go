package auth

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/jackc/pgx/v5"
	"golang.org/x/crypto/bcrypt"
	"tracking-system/config"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func Login(w http.ResponseWriter, r *http.Request) {
	var request LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid login data", http.StatusBadRequest)
		return
	}

	request.Email = strings.ToLower(strings.TrimSpace(request.Email))

	if request.Email == "" || request.Password == "" {
		http.Error(w, "Email and password are required", http.StatusBadRequest)
		return
	}

	var userID int64
	var passwordHash string
	var role string

	err := config.DB.QueryRow(
		r.Context(),
		`SELECT id, password_hash, role
		 FROM users
		 WHERE LOWER(email) = $1`,
		request.Email,
	).Scan(&userID, &passwordHash, &role)

	if errors.Is(err, pgx.ErrNoRows) {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	if bcrypt.CompareHashAndPassword(
		[]byte(passwordHash),
		[]byte(request.Password),
	) != nil {
		http.Error(w, "Invalid email or password", http.StatusUnauthorized)
		return
	}

	token, err := CreateToken(userID, role)
	if err != nil {
		http.Error(w, "Login failed", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]interface{}{
		"message": "Login successful",
		"user_id": userID,
		"role":    role,
		"token":   token,
	})
}
