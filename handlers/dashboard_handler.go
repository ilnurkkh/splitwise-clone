package handlers

import (
	"context"
	"encoding/json"
	"net/http"

	"splitwise-clone/db"
	"splitwise-clone/middleware"
)

type Balance struct {
	OtherUserID int     `json:"user_id"`
	Amount      float64 `json:"amount"`
}

func GetDashboardBalances(w http.ResponseWriter, r *http.Request) {
	userID := r.Context().Value(middleware.UserIDKey).(int)

	// 1. Calculate WHO I OWE (Others paid, I was in the split)
	iOweSQL := `
		SELECT e.payer_id, SUM(es.amount_owed) 
		FROM expense_splits es
		JOIN expenses e ON es.expense_id = e.id
		WHERE es.user_id = $1 AND e.payer_id != $1
		GROUP BY e.payer_id
	`
	rowsIOwe, _ := db.Pool.Query(context.Background(), iOweSQL, userID)
	defer rowsIOwe.Close()

	var whoIOwe []Balance
	for rowsIOwe.Next() {
		var b Balance
		rowsIOwe.Scan(&b.OtherUserID, &b.Amount)
		whoIOwe = append(whoIOwe, b)
	}

	// 2. Calculate WHO OWES ME (I paid, others are in the split)
	owesMeSQL := `
		SELECT es.user_id, SUM(es.amount_owed) 
		FROM expense_splits es
		JOIN expenses e ON es.expense_id = e.id
		WHERE e.payer_id = $1 AND es.user_id != $1
		GROUP BY es.user_id
	`
	rowsOwesMe, _ := db.Pool.Query(context.Background(), owesMeSQL, userID)
	defer rowsOwesMe.Close()

	var whoOwesMe []Balance
	for rowsOwesMe.Next() {
		var b Balance
		rowsOwesMe.Scan(&b.OtherUserID, &b.Amount)
		whoOwesMe = append(whoOwesMe, b)
	}

	// 3. Send it all back as clean JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"user_id":     userID,
		"who_i_owe":   whoIOwe,
		"who_owes_me": whoOwesMe,
	})
}
