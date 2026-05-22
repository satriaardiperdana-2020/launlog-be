package handlers

import (
	"context"
	"fmt"
	"github.com/jackc/pgx/v5/pgtype"
	openapi_types "github.com/oapi-codegen/runtime/types"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"

	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/helper"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
)

type TransactionHandler struct {
	Queries *postgresql.Queries
	//Service *ServiceHandler
}

func (h *TransactionHandler) CreateTransaction(ctx context.Context, req api.CreateTransactionRequestObject) (api.CreateTransactionResponseObject, error) {
	log.Println("🔵 CreateTransaction called")

	// Validation
	if req.Body.UserId <= 0 {
		msg := "User ID is required"
		return api.CreateTransaction400JSONResponse{Message: &msg}, nil
	}
	if req.Body.CustomerId == nil {
		msg := "Customer ID is required"
		return api.CreateTransaction400JSONResponse{Message: &msg}, nil
	}
	if req.Body.Items == nil || len(req.Body.Items) == 0 {
		msg := "At least one item is required"
		return api.CreateTransaction400JSONResponse{Message: &msg}, nil
	}

	// Calculate total
	var totalAmount float64
	for _, item := range req.Body.Items {
		if item.ServiceId == nil {
			msg := "Service ID is required for each item"
			return api.CreateTransaction400JSONResponse{Message: &msg}, nil
		}
		// Konversi *int ke int64
		serviceID := int64(*item.ServiceId)

		service, err := h.Queries.GetServiceById(ctx, serviceID)
		if err != nil {
			msg := fmt.Sprintf("Service ID %d not found", item.ServiceId)
			return api.CreateTransaction400JSONResponse{Message: &msg}, nil
		}
		price := helper.NumericToFloat64(service.Price)
		totalAmount += price * float64(*item.Qty)
		log.Println("Show total Amount :  ", totalAmount)
	}

	// Generate invoice
	invoiceNo := fmt.Sprintf("INV-%s-%d", time.Now().Format("200601"), time.Now().UnixNano()%1000)

	isDelivery := false
	if req.Body.IsDelivery != nil {
		isDelivery = *req.Body.IsDelivery
	}

	// ✅ Konversi CustomerId ke pgtype.Int8 (nullable)
	customerID := pgtype.Int8{}
	if req.Body.CustomerId != nil {
		customerID.Int64 = int64(*req.Body.CustomerId)
		customerID.Valid = true
	}

	// ✅ Konversi Notes ke pgtype.Text
	notes := pgtype.Text{}
	if req.Body.Notes != nil {
		notes.String = *req.Body.Notes
		notes.Valid = true
	}
	// ✅ Konversi TotalAmount ke pgtype.Numeric
	totalAmountNumeric := pgtype.Numeric{}
	if err := totalAmountNumeric.Scan(totalAmount); err != nil {
		msg := "Failed to convert total amount"
		return api.CreateTransaction400JSONResponse{Message: &msg}, nil
	}
	// Create transaction
	transaction, err := h.Queries.CreateTransaction(ctx, postgresql.CreateTransactionParams{
		InvoiceNo:   invoiceNo,
		UserID:      int64(req.Body.UserId),
		CustomerID:  customerID,
		IsDelivery:  isDelivery,
		TotalAmount: totalAmountNumeric,
		Notes:       notes,
	})
	if err != nil {
		msg := "Failed to create transaction: " + err.Error()
		return api.CreateTransaction400JSONResponse{Message: &msg}, nil
	}

	// ==================== CREATE ITEMS ====================
	for _, item := range req.Body.Items {
		if item.ServiceId == nil {
			continue
		}

		// ✅ ServiceId ke pgtype.Int8
		serviceID := pgtype.Int8{}
		if item.ServiceId != nil {
			serviceID.Int64 = int64(*item.ServiceId)
			serviceID.Valid = true
		}

		// ✅ Qty ke pgtype.Numeric (bukan float64)
		qtyNumeric := pgtype.Numeric{}
		if item.Qty != nil {
			qtyNumeric.Scan(float64(*item.Qty))
		}

		// Notes ke pgtype.Text
		itemNotes := pgtype.Text{}
		if item.Notes != nil {
			itemNotes.String = *item.Notes
			itemNotes.Valid = true
		}
		_, err := h.Queries.CreateTransactionItem(ctx, postgresql.CreateTransactionItemParams{
			TransactionID: transaction.ID,
			ServiceID:     serviceID,
			Qty:           qtyNumeric,
			Notes:         itemNotes,
		})
		if err != nil {
			log.Printf("Warning: Failed to create item: %v", err)
		}
	}

	// ==================== RESPONSE ====================
	respNotes := ""
	if transaction.Notes.Valid {
		respNotes = transaction.Notes.String
	}
	totalAmountFloat32 := helper.NumericToFloat32(transaction.TotalAmount)
	resp := api.Transaction{
		Id:              helper.Int64ToIntPtr(transaction.ID),
		InvoiceNo:       &transaction.InvoiceNo,
		UserId:          helper.Int64ToIntPtr(transaction.UserID),
		CustomerId:      helper.PgInt8ToIntPtr(transaction.CustomerID),
		PaymentStatus:   &transaction.PaymentStatus,
		TotalAmount:     &totalAmountFloat32,
		Notes:           &respNotes,
		TransactionDate: &transaction.TransactionDate,
	}
	return api.CreateTransaction201JSONResponse(resp), nil
}

func (h *TransactionHandler) ListTransactions(ctx context.Context, req api.ListTransactionsRequestObject) (api.ListTransactionsResponseObject, error) {
	log.Println("🔵 ListTransactions called")
	startDate := pgtype.Date{}
	endDate := pgtype.Date{}
	var transactionType *string

	// Jika ada parameter dari request, isi
	if req.Params.StartDate != nil {
		startDate.Scan(*req.Params.StartDate)
	}
	if req.Params.EndDate != nil {
		endDate.Scan(*req.Params.EndDate)
	}
	if req.Params.Type != nil {
		transactionType = (*string)(req.Params.Type)
	}

	transactions, err := h.Queries.ListTransactions(ctx, postgresql.ListTransactionsParams{
		StartDate:       startDate,
		EndDate:         endDate,
		TransactionType: *transactionType,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	resp := make([]api.Transaction, len(transactions))
	for i, t := range transactions {
		notes := ""
		if t.Notes.Valid {
			notes = t.Notes.String
		}

		idInt := int(t.ID)
		userIdInt := int(t.UserID)
		customerId := int(t.UserID)
		totalAmount := helper.NumericToFloat32(t.TotalAmount)
		paidAmount := helper.NumericToFloat32(t.PaidAmount)
		resp[i] = api.Transaction{
			Id:              &idInt,
			InvoiceNo:       &t.InvoiceNo,
			UserId:          &userIdInt,
			CustomerId:      &customerId,
			PaymentStatus:   &t.PaymentStatus,
			TotalAmount:     &totalAmount,
			PaidAmount:      &paidAmount,
			Notes:           &notes,
			TransactionDate: &t.TransactionDate,
		}
	}
	return api.ListTransactions200JSONResponse(resp), nil
}

// UpdateTransaction updates an existing income transaction
func (h *TransactionHandler) UpdateTransaction(ctx context.Context, req api.UpdateTransactionRequestObject) (api.UpdateTransactionResponseObject, error) {
	log.Printf("🔵 UpdateTransaction called for ID: %d", req.Id)

	// Check if transaction exists and is income
	existing, err := h.Queries.GetTransactionById(ctx, int64(req.Id))
	if err != nil {
		msg := "Transaction not found"
		return api.UpdateTransaction404JSONResponse{Message: &msg}, nil
	}
	if existing.TransactionType != "income" {
		msg := "Transaction is not an income"
		return api.UpdateTransaction400JSONResponse{Message: &msg}, nil
	}

	// Update transaction
	customerID := pgtype.Int8{}
	if req.Body.CustomerId != nil {
		customerID.Int64 = int64(*req.Body.CustomerId)
		customerID.Valid = true
	}
	paymentStatusStr := ""
	if req.Body.PaymentStatus != nil {
		paymentStatusStr = string(*req.Body.PaymentStatus)
	}

	// Kemudian konversi ke pgtype.Text
	paymentStatus := pgtype.Text{}
	if paymentStatusStr != "" {
		paymentStatus.String = paymentStatusStr
		paymentStatus.Valid = true
	}

	// Konversi IsDelivery (*bool -> pgtype.Bool)
	isDelivery := pgtype.Bool{}
	if req.Body.IsDelivery != nil {
		isDelivery.Bool = *req.Body.IsDelivery
		isDelivery.Valid = true
	}

	// Konversi TotalAmount (*float64 -> pgtype.Numeric)
	totalAmount := pgtype.Numeric{}
	if req.Body.TotalAmount != nil {
		totalAmount.Scan(*req.Body.TotalAmount)
	}

	// Konversi Notes (*string -> pgtype.Text)
	notes := pgtype.Text{}
	if req.Body.Notes != nil {
		notes.String = *req.Body.Notes
		notes.Valid = true
	}
	// ==================== UPDATE TRANSACTION ====================
	updated, err := h.Queries.UpdateIncome(ctx, postgresql.UpdateIncomeParams{
		ID:            int64(req.Id),
		CustomerID:    customerID,
		PaymentStatus: paymentStatus,
		IsDelivery:    isDelivery,
		TotalAmount:   totalAmount,
		Notes:         notes,
	})
	if err != nil {
		msg := "Failed to update transaction: " + err.Error()
		return api.UpdateTransaction400JSONResponse{Message: &msg}, nil
	}

	// ==================== UPDATE ITEMS ====================
	if req.Body.Items != nil {
		for _, item := range *req.Body.Items {
			if item.Id != nil {
				// Konversi Qty (*float32 -> pgtype.Numeric)
				qtyNumeric := pgtype.Numeric{}
				if item.Qty != nil {
					qtyNumeric.Scan(float64(*item.Qty))
				}

				// Konversi UnitPrice (*float32 -> pgtype.Numeric)
				unitPriceNumeric := pgtype.Numeric{}
				if item.UnitPrice != nil {
					unitPriceNumeric.Scan(float64(*item.UnitPrice))
				}

				// Konversi Notes (*string -> pgtype.Text)
				notes := pgtype.Text{}
				if item.Notes != nil {
					notes.String = *item.Notes
					notes.Valid = true
				}
				_, err := h.Queries.UpdateIncomeItem(ctx, postgresql.UpdateIncomeItemParams{
					ID:        int64(*item.Id),
					Qty:       qtyNumeric,
					UnitPrice: unitPriceNumeric,
					Notes:     notes,
				})
				if err != nil {
					log.Printf("Warning: Failed to update item %d: %v", *item.Id, err)
				}
			}
		}
	}

	// ==================== RESPONSE ====================
	respNotes := ""
	if updated.Notes.Valid {
		respNotes = updated.Notes.String
	}

	customerId := 0
	if updated.CustomerID.Valid {
		customerId = int(updated.CustomerID.Int64)
	}

	idInt := int(updated.ID)
	userIdInt := int(updated.UserID)
	totalAmountResp := helper.NumericToFloat32(updated.TotalAmount)
	paidAmountResp := helper.NumericToFloat32(updated.PaidAmount)
	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &updated.InvoiceNo,
		UserId:          &userIdInt,
		CustomerId:      &customerId,
		PaymentStatus:   &updated.PaymentStatus,
		TotalAmount:     &totalAmountResp,
		PaidAmount:      &paidAmountResp,
		Notes:           &respNotes,
		TransactionDate: &updated.TransactionDate,
	}
	return api.UpdateTransaction200JSONResponse(resp), nil
}

// SoftDeleteTransaction soft deletes a transaction -- GANTI SOFTDELETE TRANSACTION INCOME
/*func (h *TransactionHandler) SoftDeleteTransaction(ctx context.Context, req api.SoftDeleteTransactionRequestObject) (api.SoftDeleteTransactionResponseObject, error) {
	log.Printf("🔵 SoftDeleteTransaction called for ID: %d", req.Id)

	deleted, err := h.Queries.SoftDeleteTransaction(ctx, int64(req.Id))
	if err != nil {
		msg := "Transaction not found or already deleted"
		return api.SoftDeleteTransaction404JSONResponse{Message: &msg}, nil
	}

	notes := ""
	if deleted.Notes.Valid {
		notes = deleted.Notes.String
	}
	// Konversi tipe data
	idInt := int(deleted.ID)
	userIdInt := int(deleted.UserID)

	customerId := 0
	if deleted.CustomerID.Valid {
		customerId = int(deleted.CustomerID.Int64)
	}

	totalAmountResp := helper.NumericToFloat32(deleted.TotalAmount)
	paidAmountResp := helper.NumericToFloat32(deleted.PaidAmount)
	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &deleted.InvoiceNo,
		UserId:          &userIdInt,
		CustomerId:      &customerId,
		PaymentStatus:   &deleted.PaymentStatus,
		TotalAmount:     &totalAmountResp,
		PaidAmount:      &paidAmountResp,
		Notes:           &notes,
		TransactionDate: &deleted.TransactionDate,
	}
	return api.SoftDeleteTransaction200JSONResponse(resp), nil
}*/

// RestoreTransaction restores a soft deleted transaction
func (h *TransactionHandler) RestoreTransaction(ctx context.Context, req api.RestoreTransactionRequestObject) (api.RestoreTransactionResponseObject, error) {
	log.Printf("🔵 RestoreTransaction called for ID: %d", req.Id)

	restored, err := h.Queries.RestoreTransaction(ctx, int64(req.Id))
	if err != nil {
		msg := "Transaction not found or not deleted"
		return api.RestoreTransaction404JSONResponse{Message: &msg}, nil
	}

	notes := ""
	if restored.Notes.Valid {
		notes = restored.Notes.String
	}
	// Konversi tipe data
	idInt := int(restored.ID)
	userIdInt := int(restored.UserID)

	customerId := 0
	if restored.CustomerID.Valid {
		customerId = int(restored.CustomerID.Int64)
	}

	totalAmountResp := helper.NumericToFloat32(restored.TotalAmount)
	paidAmountResp := helper.NumericToFloat32(restored.PaidAmount)
	resp := api.Transaction{
		Id:              &idInt,
		InvoiceNo:       &restored.InvoiceNo,
		UserId:          &userIdInt,
		CustomerId:      &customerId,
		PaymentStatus:   &restored.PaymentStatus,
		TotalAmount:     &totalAmountResp,
		PaidAmount:      &paidAmountResp,
		Notes:           &notes,
		TransactionDate: &restored.TransactionDate,
	}
	return api.RestoreTransaction200JSONResponse(resp), nil
}

// GetTransactionReport returns transactions by date range
// GetTransactionReport returns transaction report by date range
func (h *TransactionHandler) GetTransactionReport(ctx context.Context, req api.GetTransactionReportRequestObject) (api.GetTransactionReportResponseObject, error) {
	log.Println("🔵 GetTransactionReport called")

	// ==================== VALIDATION ====================
	if req.Params.StartDate == (openapi_types.Date{}) {
		msg := "Start date is required"
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}
	if req.Params.EndDate == (openapi_types.Date{}) {
		msg := "End date is required"
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}

	// ==================== KONVERSI KE pgtype.Date ====================
	startDate := pgtype.Date{}
	if err := startDate.Scan(req.Params.StartDate.String()); err != nil {
		msg := "Invalid start date format"
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}

	endDate := pgtype.Date{}
	if err := endDate.Scan(req.Params.EndDate.String()); err != nil {
		msg := "Invalid end date format"
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}

	// Transaction type filter (optional)
	var typeStr string
	if req.Params.Type != nil {
		typeStr = string(*req.Params.Type)
	}

	// ==================== GET TRANSACTIONS ====================
	transactions, err := h.Queries.GetTransactionReport(ctx, postgresql.GetTransactionReportParams{
		StartDate: startDate,
		EndDate:   endDate,
		Type:      typeStr,
	})
	if err != nil {
		msg := "Failed to fetch transactions: " + err.Error()
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}

	// ==================== GET SUMMARY ====================
	summary, err := h.Queries.GetTransactionSummary(ctx, postgresql.GetTransactionSummaryParams{
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		msg := "Failed to fetch summary: " + err.Error()
		return api.GetTransactionReport400JSONResponse{Message: &msg}, nil
	}

	// ==================== BUILD RESPONSE ====================
	reportItems := make([]api.TransactionReportItem, len(transactions))
	for i, t := range transactions {
		// Konversi tipe data
		idInt := int(t.ID)
		userIdInt := int(t.UserID)
		customerIdInt := 0
		if t.CustomerID.Valid {
			customerIdInt = int(t.CustomerID.Int64)
		}

		totalAmount := helper.NumericToFloat32(t.TotalAmount)
		paidAmount := helper.NumericToFloat32(t.PaidAmount)

		notes := ""
		if t.Notes.Valid {
			notes = t.Notes.String
		}

		customerName := t.CustomerName
		supplier := t.Supplier
		expenseCategory := t.ExpenseCategory

		reportItems[i] = api.TransactionReportItem{
			Id:              &idInt,
			InvoiceNo:       &t.InvoiceNo,
			TransactionType: &t.TransactionType,
			UserId:          &userIdInt,
			CustomerId:      &customerIdInt,
			CustomerName:    &customerName,
			Supplier:        &supplier,
			ExpenseCategory: &expenseCategory,
			PaymentStatus:   &t.PaymentStatus,
			TotalAmount:     &totalAmount,
			PaidAmount:      &paidAmount,
			Notes:           &notes,
			TransactionDate: &t.TransactionDate,
		}
	}

	// Konversi summary
	// Konversi dari interface{} ke float32
	totalIncome := float32(summary.TotalIncome.(float64))
	totalExpense := float32(summary.TotalExpense.(float64))
	netProfit := float32(summary.NetProfit.(float64))

	// ✅ Gunakan tipe yang sudah didefinisikan
	resp := api.GetTransactionReport200JSONResponse{
		Transactions: &reportItems,
		Summary: &struct {
			NetProfit    *float32 `json:"netProfit,omitempty"`
			TotalExpense *float32 `json:"totalExpense,omitempty"`
			TotalIncome  *float32 `json:"totalIncome,omitempty"`
		}{
			NetProfit:    &netProfit,
			TotalExpense: &totalExpense,
			TotalIncome:  &totalIncome,
		},
	}
	return resp, nil
}
