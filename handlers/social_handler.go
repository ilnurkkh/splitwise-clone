package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"splitwise-clone/db"
	"splitwise-clone/middleware"
)

// AddFriend handles POST /api/friends
func AddFriend(w http.ResponseWriter, r *http.Request) {

	userID1 := r.Context().Value(middleware.UserIDKey).(int)

	var input struct {
		UserID2 int `json:"user_id_2"` // The friend they are adding
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Insert the friendship
	sql := `INSERT INTO friends (user_id_1, user_id_2) VALUES ($1, $2)`
	_, err := db.Pool.Exec(context.Background(), sql, userID1, input.UserID2)

	if err != nil {
		http.Error(w, "Could not add friend (might already be friends or user doesn't exist)", http.StatusConflict)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{"message": "Friend added successfully"})
}

// CreateGroup handles POST /api/groups
func CreateGroup(w http.ResponseWriter, r *http.Request) {
	var input struct {
		Name    string `json:"name"`
		UserIDs []int  `json:"user_ids"` // List of user IDs to add to the group
	}

	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// This requires a transaction because we are inserting into 'groups' AND 'group_members'
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}
	// Defer rollback ensures that if anything fails before tx.Commit(), all changes are undone.
	defer tx.Rollback(context.Background())

	// 1. Create the group and get the new group ID
	var groupID int
	err = tx.QueryRow(context.Background(), `INSERT INTO groups (name) VALUES ($1) RETURNING id`, input.Name).Scan(&groupID)
	if err != nil {
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// 2. Add all requested members to the group
	memberSQL := `INSERT INTO group_members (group_id, user_id) VALUES ($1, $2)`
	for _, userID := range input.UserIDs {
		_, err = tx.Exec(context.Background(), memberSQL, groupID, userID)
		if err != nil {
			http.Error(w, "Failed to add user to group", http.StatusInternalServerError)
			return
		}
	}

	// 3. Save everything permanently
	tx.Commit(context.Background())
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":  "Group created successfully",
		"group_id": groupID,
	})
}
