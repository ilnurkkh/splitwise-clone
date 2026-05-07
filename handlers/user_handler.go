package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"splitwise-clone/db"
	"splitwise-clone/middleware"
)

func GetUserProfile(w http.ResponseWriter, r *http.Request) {
	// Securely get the user ID from the JWT token context
	userID := r.Context().Value(middleware.UserIDKey).(int)

	var profile struct {
		FullName    string    `json:"full_name"`
		Email       string    `json:"email"`
		PhoneNumber *string   `json:"phone_number"` // Pointer because it can be NULL in the DB
		CreatedAt   time.Time `json:"created_at"`
		TotalGroups int       `json:"total_active_groups"`
	}

	// We use a subquery to count the total active groups for this specific user
	sql := `
		SELECT u.full_name, u.email, u.phone_number, u.created_at,
		       (SELECT COUNT(*) FROM group_members WHERE user_id = $1) as total_groups
		FROM users u
		WHERE u.id = $1
	`

	err := db.Pool.QueryRow(context.Background(), sql, userID).Scan(
		&profile.FullName, &profile.Email, &profile.PhoneNumber, &profile.CreatedAt, &profile.TotalGroups,
	)

	if err != nil {
		http.Error(w, "Could not fetch user profile", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(profile)
}
