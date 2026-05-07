package handlers

import (
	"context"
	"encoding/json"
	"math"
	"net/http"

	"splitwise-clone/db"
	"splitwise-clone/middleware"
	"splitwise-clone/models"
)

// CreateExpense handles POST /api/expenses
func CreateExpense(w http.ResponseWriter, r *http.Request) {
	var req models.ExpenseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}

	// Safety check to prevent dividing by zero or logging negative debts
	if req.TotalAmount <= 0 {
		http.Error(w, "Amount must be greater than zero", http.StatusBadRequest)
		return
	}

	req.PayerID = r.Context().Value(middleware.UserIDKey).(int)

	// Run the algorithmic splitting logic
	calculatedSplits := calculateSplits(req)

	// Start Database Transaction
	tx, err := db.Pool.Begin(context.Background())
	if err != nil {
		http.Error(w, "Transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback(context.Background())

	// 1. Insert the main expense record
	var expenseID int
	expSQL := `INSERT INTO expenses (group_id, payer_id, description, total_amount) 
	           VALUES ($1, $2, $3, $4) RETURNING id`
	err = tx.QueryRow(context.Background(), expSQL, req.GroupID, req.PayerID, req.Description, req.TotalAmount).Scan(&expenseID)

	if err != nil {
		http.Error(w, "Failed to insert expense", http.StatusInternalServerError)
		return
	}

	// 2. Insert the individual splits
	splitSQL := `INSERT INTO expense_splits (expense_id, user_id, amount_owed) VALUES ($1, $2, $3)`
	for userID, amount := range calculatedSplits {
		// Only log a debt if they actually owe money (amount > 0)
		if amount > 0 {
			_, err = tx.Exec(context.Background(), splitSQL, expenseID, userID, amount)
			if err != nil {
				http.Error(w, "Failed to insert splits", http.StatusInternalServerError)
				return
			}
		}
	}

	// 3. Save everything permanently
	tx.Commit(context.Background())
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"message":    "Expense logged successfully",
		"expense_id": expenseID,
	})
}

// calculateSplits handles Equal, Percentage, Ratio, and Exact amount math.
func calculateSplits(req models.ExpenseRequest) map[int]float64 {
	finalAmounts := make(map[int]float64)
	totalUsers := len(req.Splits)

	if totalUsers == 0 {
		return finalAmounts
	}

	switch req.SplitType {
	case "EQUAL":
		// Dividing the total evenly among all selected participants.
		splitAmount := math.Round((req.TotalAmount/float64(totalUsers))*100) / 100
		for _, s := range req.Splits {
			finalAmounts[s.UserID] = splitAmount
		}

	case "PERCENTAGE":
		// Dividing the bill by percentages (e.g., 60/40)
		for _, s := range req.Splits {
			amount := (req.TotalAmount * s.Value) / 100.0
			finalAmounts[s.UserID] = math.Round(amount*100) / 100
		}

	case "RATIO":
		// Dividing by shares/ratios
		totalShares := 0.0
		for _, s := range req.Splits {
			totalShares += s.Value
		}
		for _, s := range req.Splits {
			amount := (req.TotalAmount * s.Value) / totalShares
			finalAmounts[s.UserID] = math.Round(amount*100) / 100
		}

	case "EXACT":
		// Specifying exactly how much each person owes.
		// Remaining amount must be split evenly among users whose debt isn't specified (value = 0).
		specifiedTotal := 0.0
		unspecifiedCount := 0
		var unspecifiedUsers []int

		for _, s := range req.Splits {
			if s.Value > 0 {
				finalAmounts[s.UserID] = s.Value
				specifiedTotal += s.Value
			} else {
				unspecifiedCount++
				unspecifiedUsers = append(unspecifiedUsers, s.UserID)
			}
		}

		remainingAmount := req.TotalAmount - specifiedTotal
		if unspecifiedCount > 0 && remainingAmount > 0 {
			evenSplit := math.Round((remainingAmount/float64(unspecifiedCount))*100) / 100
			for _, uid := range unspecifiedUsers {
				finalAmounts[uid] = evenSplit
			}
		}
	}

	return finalAmounts
}
