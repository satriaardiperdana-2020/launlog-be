package handlers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/helper"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
	"log"
	"net/http"
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
	startDate := pgtype.Date{}
	if req.Params.StartDate != nil {
		startDate.Scan(*req.Params.StartDate)
	}

	endDate := pgtype.Date{}
	if req.Params.EndDate != nil {
		endDate.Scan(*req.Params.EndDate)
	}
	expenses, err := h.Queries.ListExpenses(ctx, postgresql.ListExpensesParams{
		StartDate: startDate,
		EndDate:   endDate,
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
		// Konversi tipe data
		idInt := int(e.ID)
		userIdInt := int(e.UserID)
		totalAmount := helper.NumericToFloat32(e.TotalAmount)
		paidAmount := helper.NumericToFloat32(e.PaidAmount)
		resp[i] = api.Transaction{
			Id:              &idInt,
			InvoiceNo:       &e.InvoiceNo,
			UserId:          &userIdInt,
			PaymentStatus:   &e.PaymentStatus,
			TotalAmount:     &totalAmount,
			PaidAmount:      &paidAmount,
			Notes:           &notes,
			TransactionDate: &e.TransactionDate,
		}
	}
	return api.ListExpenses200JSONResponse(resp), nil
}

// UpdateExpense updates an existing expense
func (h *ExpenseHandler) UpdateExpense(ctx context.Context, req api.UpdateExpenseRequestObject) (api.UpdateExpenseResponseObject, error) {
	log.Printf("🔵 UpdateExpense called for ID: %d", req.Id)

	// ==================== CHECK IF EXPENSE EXISTS ====================
	_, err := h.Queries.GetExpenseById(ctx, int64(req.Id))
	if err != nil {
		msg := "Expense not found"
		return api.UpdateExpense404JSONResponse{Message: &msg}, nil
	}

	// ==================== KONVERSI KE pgtype ====================
	// Supplier (*string -> pgtype.Text)
	supplier := pgtype.Text{}
	if req.Body.Supplier != nil {
		supplier.String = *req.Body.Supplier
		supplier.Valid = true
	}

	// ExpenseCategory (*string -> pgtype.Text)
	expenseCategory := pgtype.Text{}
	if req.Body.ExpenseCategory != nil {
		expenseCategory.String = *req.Body.ExpenseCategory
		expenseCategory.Valid = true
	}

	// TotalAmount (*float64 -> pgtype.Numeric)
	totalAmount := pgtype.Numeric{}
	if req.Body.TotalAmount != nil {
		if err := totalAmount.Scan(*req.Body.TotalAmount); err != nil {
			msg := "Invalid total amount format"
			return api.UpdateExpense400JSONResponse{Message: &msg}, nil
		}
	}

	// PaidAmount (*float64 -> pgtype.Numeric)
	paidAmount := pgtype.Numeric{}
	if req.Body.PaidAmount != nil {
		if err := paidAmount.Scan(*req.Body.PaidAmount); err != nil {
			msg := "Invalid paid amount format"
			return api.UpdateExpense400JSONResponse{Message: &msg}, nil
		}
	} else if req.Body.TotalAmount != nil {
		// Jika paidAmount tidak dikirim, default ke totalAmount
		paidAmount.Scan(*req.Body.TotalAmount)
	}

	// Notes (*string -> pgtype.Text)
	notes := pgtype.Text{}
	if req.Body.Notes != nil {
		notes.String = *req.Body.Notes
		notes.Valid = true
	}

	// ==================== UPDATE EXPENSE ====================
	updated, err := h.Queries.UpdateExpense(ctx, postgresql.UpdateExpenseParams{
		ID:              int64(req.Id),
		Supplier:        supplier,
		ExpenseCategory: expenseCategory,
		TotalAmount:     totalAmount,
		PaidAmount:      paidAmount,
		Notes:           notes,
	})
	if err != nil {
		msg := "Failed to update expense: " + err.Error()
		return api.UpdateExpense400JSONResponse{Message: &msg}, nil
	}

	// ==================== UPDATE ITEMS ====================
	if req.Body.Items != nil {
		for _, item := range *req.Body.Items {
			if item.Id == nil {
				continue
			}

			// Konversi ItemName (*string -> pgtype.Text)
			itemName := pgtype.Text{}
			if item.ItemName != nil {
				itemName.String = *item.ItemName
				itemName.Valid = true
			}

			// Konversi Qty (float64 -> pgtype.Numeric)
			qty := pgtype.Numeric{}
			if err := qty.Scan(item.Qty); err != nil {
				log.Printf("Warning: Failed to convert qty: %v", err)
			}

			// Konversi UnitPrice (float64 -> pgtype.Numeric)
			unitPrice := pgtype.Numeric{}
			if err := unitPrice.Scan(item.UnitPrice); err != nil {
				log.Printf("Warning: Failed to convert unit price: %v", err)
			}

			// Konversi Notes (*string -> pgtype.Text)
			itemNotes := pgtype.Text{}
			if item.Notes != nil {
				itemNotes.String = *item.Notes
				itemNotes.Valid = true
			}

			_, err := h.Queries.UpdateExpenseItem(ctx, postgresql.UpdateExpenseItemParams{
				ID:        int64(*item.Id),
				ItemName:  itemName,
				Qty:       qty,
				UnitPrice: unitPrice,
				Notes:     itemNotes,
			})
			if err != nil {
				log.Printf("Warning: Failed to update expense item %d: %v", *item.Id, err)
			}
		}
	}

	// ==================== RESPONSE ====================
	respNotes := ""
	if updated.Notes.Valid {
		respNotes = updated.Notes.String
	}

	idInt := int(updated.ID)
	userIdInt := int(updated.UserID)
	totalAmountResp := helper.NumericToFloat32(updated.TotalAmount)
	paidAmountResp := helper.NumericToFloat32(updated.PaidAmount)

	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &updated.InvoiceNo,
		UserId:          &userIdInt,
		PaymentStatus:   &updated.PaymentStatus,
		TotalAmount:     &totalAmountResp,
		PaidAmount:      &paidAmountResp,
		Notes:           &respNotes,
		TransactionDate: &updated.TransactionDate,
	}
	return api.UpdateExpense200JSONResponse(resp), nil
}

// SoftDeleteExpense soft deletes an expense
func (h *ExpenseHandler) SoftDeleteExpense(ctx context.Context, req api.SoftDeleteExpenseRequestObject) (api.SoftDeleteExpenseResponseObject, error) {
	log.Printf("🔵 SoftDeleteExpense called for ID: %d", req.Id)

	// Call database query
	deleted, err := h.Queries.SoftDeleteExpense(ctx, int64(req.Id))
	if err != nil {
		msg := "Expense not found or already deleted"
		return api.SoftDeleteExpense404JSONResponse{Message: &msg}, nil // ✅ sekarang tersedia
	}

	log.Printf("✅ Expense %d soft deleted successfully", req.Id)

	// Convert response
	notes := ""
	if deleted.Notes.Valid {
		notes = deleted.Notes.String
	}

	idInt := int(deleted.ID)
	userIdInt := int(deleted.UserID)
	totalAmount := helper.NumericToFloat32(deleted.TotalAmount)
	paidAmount := helper.NumericToFloat32(deleted.PaidAmount)

	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &deleted.InvoiceNo,
		UserId:          &userIdInt,
		PaymentStatus:   &deleted.PaymentStatus,
		TotalAmount:     &totalAmount,
		PaidAmount:      &paidAmount,
		Notes:           &notes,
		TransactionDate: &deleted.TransactionDate,
	}
	return api.SoftDeleteExpense200JSONResponse(resp), nil
}

// GetExpense returns expense by ID
func (h *ExpenseHandler) GetExpense(ctx context.Context, req api.GetExpenseRequestObject) (api.GetExpenseResponseObject, error) {
	log.Printf("🔵 GetExpense called for ID: %d", req.Id)

	// ==================== GET EXPENSE FROM DATABASE ====================
	expense, err := h.Queries.GetExpenseById(ctx, int64(req.Id))
	if err != nil {
		msg := "Expense not found"
		return api.GetExpense404JSONResponse{Message: &msg}, nil
	}

	// ==================== RESPONSE ====================
	notes := ""
	if expense.Notes.Valid {
		notes = expense.Notes.String
	}

	// Konversi tipe data untuk response
	idInt := int(expense.ID)
	userIdInt := int(expense.UserID)
	totalAmount := helper.NumericToFloat32(expense.TotalAmount)
	paidAmount := helper.NumericToFloat32(expense.PaidAmount)

	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &expense.InvoiceNo,
		UserId:          &userIdInt,
		PaymentStatus:   &expense.PaymentStatus,
		TotalAmount:     &totalAmount,
		PaidAmount:      &paidAmount,
		Notes:           &notes,
		TransactionDate: &expense.TransactionDate,
	}
	return api.GetExpense200JSONResponse(resp), nil
}
