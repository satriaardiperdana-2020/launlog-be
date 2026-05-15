package handlers

import (
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"launlog-be/internal/api"
	"launlog-be/repository/sqlc"
	"net/http"
	"time"
)

type TransactionHandler struct {
	queries *sqlc.Queries
	db      *pgxpool.Pool
}

func NewTransactionHandler(queries *sqlc.Queries, db *pgxpool.Pool) *TransactionHandler {
	return &TransactionHandler{queries: queries, db: db}
}

func (h *TransactionHandler) CreateTransaction(ctx echo.Context) error {
	var input api.TransactionInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid input"})
	}
	tx, err := h.db.Begin(ctx.Request().Context())
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	defer tx.Rollback(ctx.Request().Context())
	qtx := h.queries.WithTx(tx)

	invoiceNo := fmt.Sprintf("INV-%s-%d", time.Now().Format("200601"), time.Now().UnixNano()%1000)
	var total float64
	for _, itm := range input.Items {
		service, err := qtx.GetServiceById(ctx.Request().Context(), itm.ServiceId)
		if err != nil {
			return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Service not found"})
		}
		total += itm.Qty * service.Price
	}
	trans, err := qtx.CreateTransaction(ctx.Request().Context(), sqlc.CreateTransactionParams{
		InvoiceNo:     invoiceNo,
		UserID:        input.UserId,
		CustomerID:    *input.CustomerId,
		PaymentStatus: "unpaid",
		IsDelivery:    input.IsDelivery,
		TotalAmount:   total,
		Notes:         input.Notes,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	for _, itm := range input.Items {
		_, err = qtx.CreateTransactionItem(ctx.Request().Context(), sqlc.CreateTransactionItemParams{
			TransactionID: trans.ID,
			ServiceID:     itm.ServiceId,
			Qty:           itm.Qty,
			Notes:         itm.Notes,
		})
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
	}
	if err = tx.Commit(ctx.Request().Context()); err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, trans)
}

// Additional methods (listTransactions, addPayment, etc.) can be added similarly
