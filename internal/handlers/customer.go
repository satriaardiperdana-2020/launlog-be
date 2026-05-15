package handlers

import (
	"github.com/labstack/echo/v4"
	"launlog-be/internal/api"
	"launlog-be/repository/sqlc"
	"net/http"
)

type CustomerHandler struct {
	queries *sqlc.Queries
}

func NewCustomerHandler(queries *sqlc.Queries) *CustomerHandler {
	return &CustomerHandler{queries: queries}
}

func (h *CustomerHandler) ListCustomers(ctx echo.Context, params api.ListCustomersParams) error {
	search := ""
	if params.Search != nil {
		search = *params.Search
	}
	customers, err := h.queries.ListCustomers(ctx.Request().Context(), search)
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusOK, customers)
}

func (h *CustomerHandler) CreateCustomer(ctx echo.Context) error {
	var input api.CustomerInput
	if err := ctx.Bind(&input); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request"})
	}
	cust, err := h.queries.CreateCustomer(ctx.Request().Context(), sqlc.CreateCustomerParams{
		Name:    input.Name,
		Phone:   input.Phone,
		Address: input.Address,
	})
	if err != nil {
		return ctx.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
	}
	return ctx.JSON(http.StatusCreated, cust)
}

func (h *CustomerHandler) GetCustomer(ctx echo.Context, id int64) error {
	cust, err := h.queries.GetCustomerById(ctx.Request().Context(), id)
	if err != nil {
		return ctx.JSON(http.StatusNotFound, map[string]string{"error": "Customer not found"})
	}
	return ctx.JSON(http.StatusOK, cust)
}
