package handlers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/satriaardiperdana-2020/launlog-be/internal/helper"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
)

type ExpenseHandler struct {
	Queries *postgresql.Queries
}

func (h *ExpenseHandler) CreateExpense(ctx context.Context, req api.CreateExpenseRequestObject) (api.CreateExpenseResponseObject, error) {
	log.Println("🔵 CreateExpense called")

	// Konversi Supplier dari string ke pgtype.Text
	supplier := pgtype.Text{}
	if req.Body.Supplier != "" {
		supplier.String = req.Body.Supplier
		supplier.Valid = true
	}

	// Konversi ExpenseCategory dari string ke pgtype.Text
	expenseCategory := pgtype.Text{}
	if req.Body.ExpenseCategory != "" {
		expenseCategory.String = req.Body.ExpenseCategory
		expenseCategory.Valid = true
	}

	// ✅ Validasi Total Amount (tidak boleh 0 atau negatif)
	if req.Body.TotalAmount <= 0 {
		msg := "Total amount must be greater than 0"
		return api.CreateExpense400JSONResponse{Message: &msg}, nil
	}

	// ✅ BENAR - dereference pointer dulu
	if req.Body.PaidAmount != nil && *req.Body.PaidAmount < 0 {
		msg := "Paid amount cannot be negative"
		return api.CreateExpense400JSONResponse{Message: &msg}, nil
	}

	if *req.Body.PaidAmount > req.Body.TotalAmount {
		msg := "Paid amount cannot exceed total amount"
		return api.CreateExpense400JSONResponse{Message: &msg}, nil
	}

	// ✅ Tentukan payment_status berdasarkan paid_amount
	var paymentStatus string
	if *req.Body.PaidAmount == 0 {
		paymentStatus = "unpaid"
	} else if *req.Body.PaidAmount < req.Body.TotalAmount {
		paymentStatus = "dp"
	} else {
		paymentStatus = "paid"
	}
	// Konversi PaidAmount dari *float32 ke pgtype.Numeric
	var paidAmountNumeric pgtype.Numeric
	if req.Body.PaidAmount != nil {
		paidAmountNumeric.Scan(float64(*req.Body.PaidAmount))
	} else {
		// Default ke totalAmount
		paidAmountNumeric.Scan(req.Body.TotalAmount)
	}
	// ✅ TotalAmount (float64 -> pgtype.Numeric)
	totalAmountNumeric := pgtype.Numeric{}
	if err := totalAmountNumeric.Scan(req.Body.TotalAmount); err != nil {
		msg := "Invalid total amount format"
		return api.CreateExpense400JSONResponse{Message: &msg}, nil
	}
	// Konversi Notes (sudah ada)
	notes := pgtype.Text{}
	if req.Body.Notes != nil {
		notes.String = *req.Body.Notes
		notes.Valid = true
	}
	//invoiceNo := fmt.Sprintf("EXP-%s-%d", time.Now().Format("200601"), time.Now().UnixNano()%1000)
	invoiceNo := helper.GenerateInvoiceNo("EXP")
	// Create expense
	expense, err := h.Queries.CreateExpense(ctx, postgresql.CreateExpenseParams{
		InvoiceNo:       invoiceNo,
		UserID:          int64(req.Body.UserId),
		Supplier:        supplier,        // ← pgtype.Text
		ExpenseCategory: expenseCategory, // ← pgtype.Text
		PaidAmount:      paidAmountNumeric,
		TotalAmount:     totalAmountNumeric,
		PaymentStatus:   paymentStatus,
		Notes:           notes,
	})
	if err != nil {
		msg := "Failed to create expense: " + err.Error()
		return api.CreateExpense400JSONResponse{Message: &msg}, nil
	}

	// Create items if any
	if req.Body.Items != nil {
		for _, item := range *req.Body.Items {
			itemName := pgtype.Text{}
			if item.ItemName != nil {
				itemName.String = *item.ItemName
				itemName.Valid = true
			}

			itemNotes2 := pgtype.Text{}
			if item.Notes != nil {
				itemNotes2.String = *item.Notes
				itemNotes2.Valid = true
			}

			qtyNumeric := pgtype.Numeric{}
			qtyNumeric.Scan(item.Qty)

			unitPriceNumeric := pgtype.Numeric{}
			unitPriceNumeric.Scan(item.UnitPrice)
			_, err := h.Queries.CreateExpenseItem(ctx, postgresql.CreateExpenseItemParams{
				TransactionID: expense.ID,
				ItemName:      itemName,
				Qty:           qtyNumeric,
				UnitPrice:     unitPriceNumeric,
				Notes:         itemNotes2,
			})
			if err != nil {
				log.Printf("Warning: Failed to create expense item: %v", err)
			}
		}
	}

	// ✅ Konversi int64 ke int
	idInt := int(expense.ID)
	userIdInt := int(expense.UserID)
	respNotes := ""
	totalAmount := helper.NumericToFloat32(expense.TotalAmount)
	paidAmount := helper.NumericToFloat32(expense.PaidAmount)
	if expense.Notes.Valid {
		respNotes = expense.Notes.String
	}
	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &expense.InvoiceNo,
		UserId:          &userIdInt,
		PaymentStatus:   &expense.PaymentStatus,
		TotalAmount:     &totalAmount,
		PaidAmount:      &paidAmount,
		Notes:           &respNotes,
		TransactionDate: &expense.TransactionDate,
	}
	return api.CreateExpense201JSONResponse(resp), nil
}

func (h *ExpenseHandler) ListExpenses(ctx context.Context, req api.ListExpensesRequestObject) (api.ListExpensesResponseObject, error) {
	log.Println("🔵 ListExpenses called")

	expenses, err := h.Queries.ListExpenses(ctx, postgresql.ListExpensesParams{
		StartDate: nil,
		EndDate:   nil,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.Transaction, len(expenses))
	for i, e := range expenses {
		notes := ""
		if e.Notes.Valid {
			notes = e.Notes.String
		}
		resp[i] = api.Transaction{
			Id:              &e.ID,
			InvoiceNo:       &e.InvoiceNo,
			UserId:          &e.UserID,
			PaymentStatus:   &e.PaymentStatus,
			TotalAmount:     &e.TotalAmount,
			PaidAmount:      &e.PaidAmount,
			Notes:           &notes,
			TransactionDate: &e.TransactionDate,
		}
	}
	return api.ListExpenses200JSONResponse(resp), nil
}

// UpdateExpense updates an existing expense
func (h *ExpenseHandler) UpdateExpense(ctx context.Context, req api.UpdateExpenseRequestObject) (api.UpdateExpenseResponseObject, error) {
	log.Printf("🔵 UpdateExpense called for ID: %d", req.Id)

	existing, err := h.Queries.GetExpenseById(ctx, req.Id)
	if err != nil {
		msg := "Expense not found"
		return api.UpdateExpense404JSONResponse{Message: &msg}, nil
	}

	updated, err := h.Queries.UpdateExpense(ctx, db.UpdateExpenseParams{
		ID:              req.Id,
		Supplier:        req.Body.Supplier,
		ExpenseCategory: req.Body.ExpenseCategory,
		TotalAmount:     req.Body.TotalAmount,
		PaidAmount:      req.Body.TotalAmount, // expense always paid
		Notes:           req.Body.Notes,
	})
	if err != nil {
		msg := "Failed to update expense: " + err.Error()
		return api.UpdateExpense400JSONResponse{Message: &msg}, nil
	}

	if req.Body.Items != nil {
		for _, item := range *req.Body.Items {
			if item.Id != nil {
				h.Queries.UpdateExpenseItem(ctx, db.UpdateExpenseItemParams{
					ID:        *item.Id,
					ItemName:  item.ItemName,
					Qty:       item.Qty,
					UnitPrice: item.UnitPrice,
					Notes:     item.Notes,
				})
			}
		}
	}

	notes := ""
	if updated.Notes.Valid {
		notes = updated.Notes.String
	}
	resp := api.Transaction{
		Id:              &updated.ID,
		InvoiceNo:       &updated.InvoiceNo,
		UserId:          &updated.UserID,
		PaymentStatus:   &updated.PaymentStatus,
		TotalAmount:     &updated.TotalAmount,
		PaidAmount:      &updated.PaidAmount,
		Notes:           &notes,
		TransactionDate: &updated.TransactionDate,
	}
	return api.UpdateExpense200JSONResponse(resp), nil
}

// SoftDeleteExpense soft deletes an expense
func (h *ExpenseHandler) SoftDeleteExpense(ctx context.Context, req api.SoftDeleteExpenseRequestObject) (api.SoftDeleteExpenseResponseObject, error) {
	log.Printf("🔵 SoftDeleteExpense called for ID: %d", req.Id)

	deleted, err := h.Queries.SoftDeleteTransaction(ctx, req.Id)
	if err != nil {
		msg := "Expense not found or already deleted"
		return api.SoftDeleteExpense404JSONResponse{Message: &msg}, nil
	}

	notes := ""
	if deleted.Notes.Valid {
		notes = deleted.Notes.String
	}
	resp := api.Transaction{
		Id:              &deleted.ID,
		InvoiceNo:       &deleted.InvoiceNo,
		UserId:          &deleted.UserID,
		TotalAmount:     &deleted.TotalAmount,
		Notes:           &notes,
		TransactionDate: &deleted.TransactionDate,
	}
	return api.SoftDeleteExpense200JSONResponse(resp), nil
}
