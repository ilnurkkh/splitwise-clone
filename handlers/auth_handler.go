package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"time"

	"splitwise-clone/db"
	"splitwise-clone/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func RegisterUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		FullName string `json:"full_name"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Bad request", http.StatusBadRequest)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	sql := `INSERT INTO users (full_name, email, password_hash) VALUES ($1, $2, $3) RETURNING id`
	var newID int
	err = db.Pool.QueryRow(context.Background(), sql, input.FullName, input.Email, string(hashedPassword)).Scan(&newID)

	if err != nil {
		http.Error(w, "Could not create user (email might exist)", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]int{"user_id": newID})
}

func LoginUser(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&input)

	var user models.User
	sql := `SELECT id, password_hash FROM users WHERE email = $1`
	err := db.Pool.QueryRow(context.Background(), sql, input.Email).Scan(&user.ID, &user.PasswordHash)

	if err != nil || bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(input.Password)) != nil {
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}

	expirationTime := time.Now().Add(24 * time.Hour)
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     expirationTime.Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// FIX: Read the secret directly here
	secret := os.Getenv("JWT_SECRET")
	tokenString, _ := token.SignedString([]byte(secret))

	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}
