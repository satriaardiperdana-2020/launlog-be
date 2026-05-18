package handlers

import (
	"context"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
	"log"
	"net/http"
)

type CustomerHandler struct {
	Queries *postgresql.Queries
}

// CreateCustomer implements strict server interface
func (h *CustomerHandler) CreateCustomer(ctx context.Context, req api.CreateCustomerRequestObject) (api.CreateCustomerResponseObject, error) {
	// Konversi string ke pgtype.Text untuk field nullable
	phone := pgtype.Text{String: *req.Body.Phone, Valid: *req.Body.Phone != ""}
	address := pgtype.Text{String: *req.Body.Address, Valid: *req.Body.Address != ""}

	cust, err := h.Queries.CreateCustomer(ctx, postgresql.CreateCustomerParams{
		Name:    req.Body.Name,
		Phone:   phone,
		Address: address,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to create customer")
	}

	// Konversi balik ke string untuk response
	phoneStr := ""
	if cust.Phone.Valid {
		phoneStr = cust.Phone.String
	}
	addrStr := ""
	if cust.Address.Valid {
		addrStr = cust.Address.String
	}

	resp := api.Customer{
		Id:        &cust.ID,
		Name:      &cust.Name,
		Phone:     &phoneStr,
		Address:   &addrStr,
		IsActive:  &cust.IsActive,
		CreatedAt: &cust.CreatedAt,
	}
	return api.CreateCustomer201JSONResponse(resp), nil
}

// GetCustomerById implements GET /customers/{id}
func (h *CustomerHandler) GetCustomerById(ctx context.Context, req api.GetCustomerByIdRequestObject) (api.GetCustomerByIdResponseObject, error) {
	// Call database query
	customer, err := h.Queries.GetCustomerById(ctx, req.Id)
	if err != nil {
		// Check if error is "no rows" (customer not found)
		return nil, echo.NewHTTPError(http.StatusNotFound, "Customer not found")
	}

	// Handle nullable fields
	phone := ""
	if customer.Phone.Valid {
		phone = customer.Phone.String
	}
	address := ""
	if customer.Address.Valid {
		address = customer.Address.String
	}

	// Convert to API response
	resp := api.CustomerDetail{
		Id:        &customer.ID,
		Name:      &customer.Name,
		Phone:     &phone,
		Address:   &address,
		IsActive:  &customer.IsActive,
		CreatedAt: &customer.CreatedAt,
	}

	return api.GetCustomerById200JSONResponse(resp), nil
}

func (h *CustomerHandler) UpdateCustomer(ctx context.Context, req api.UpdateCustomerRequestObject) (api.UpdateCustomerResponseObject, error) {
	// Konversi pointer *string ke pgtype.Text
	phone := pgtype.Text{}
	if req.Body.Phone != nil {
		phone.String = *req.Body.Phone
		phone.Valid = true
	}
	address := pgtype.Text{}
	if req.Body.Address != nil {
		address.String = *req.Body.Address
		address.Valid = true
	}

	isActive := false
	if req.Body.IsActive != nil {
		isActive = *req.Body.IsActive
	}

	log.Printf("UpdateCustomer: ID=%d, Name=%s, Phone={%s valid=%v}, Address={%s valid=%v}, IsActive=%v",
		req.Id, req.Body.Name, phone.String, phone.Valid, address.String, address.Valid, isActive)

	cust, err := h.Queries.UpdateCustomer(ctx, postgresql.UpdateCustomerParams{
		ID:       req.Id,
		Name:     req.Body.Name,
		Phone:    phone,
		Address:  address,
		IsActive: isActive,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to update customer: "+err.Error())
	}

	// Konversi balik ke string untuk response
	phoneStr := ""
	if cust.Phone.Valid {
		phoneStr = cust.Phone.String
	}
	addrStr := ""
	if cust.Address.Valid {
		addrStr = cust.Address.String
	}

	resp := api.Customer{
		Id:        &cust.ID,
		Name:      &cust.Name,
		Phone:     &phoneStr,
		Address:   &addrStr,
		IsActive:  &cust.IsActive,
		CreatedAt: &cust.CreatedAt,
	}
	return api.UpdateCustomer200JSONResponse(resp), nil
}

// SoftDeleteCustomer implements strict server interface
func (h *CustomerHandler) SoftDeleteCustomer(ctx context.Context, req api.SoftDeleteCustomerRequestObject) (api.SoftDeleteCustomerResponseObject, error) {
	cust, err := h.Queries.SoftDeleteCustomer(ctx, req.Id)
	if err != nil {
		// Jika tidak ada baris yang diupdate, error "no rows"
		return nil, echo.NewHTTPError(http.StatusNotFound, "Customer not found")
	}

	// Konversi untuk response
	phoneStr := ""
	if cust.Phone.Valid {
		phoneStr = cust.Phone.String
	}
	addrStr := ""
	if cust.Address.Valid {
		addrStr = cust.Address.String
	}

	resp := api.Customer{
		Id:        &cust.ID,
		Name:      &cust.Name,
		Phone:     &phoneStr,
		Address:   &addrStr,
		IsActive:  &cust.IsActive,
		CreatedAt: &cust.CreatedAt,
	}
	return api.SoftDeleteCustomer200JSONResponse(resp), nil
}

// ListCustomers implements GET /customers
func (h *CustomerHandler) ListCustomers(ctx context.Context, req api.ListCustomersRequestObject) (api.ListCustomersResponseObject, error) {
	// Get search parameter from query (optional)
	search := ""
	if req.Params.Search != nil {
		search = *req.Params.Search
	}

	// Call database query
	customers, err := h.Queries.ListCustomers(ctx, search)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, "Failed to fetch customers: "+err.Error())
	}

	// Convert database response to API response
	resp := make([]api.Customer, len(customers))
	for i, c := range customers {
		// Handle nullable phone
		phone := ""
		if c.Phone.Valid {
			phone = c.Phone.String
		}

		// Handle nullable address
		address := ""
		if c.Address.Valid {
			address = c.Address.String
		}

		resp[i] = api.Customer{
			Id:        &c.ID,
			Name:      &c.Name,
			Phone:     &phone,
			Address:   &address,
			IsActive:  &c.IsActive,
			CreatedAt: &c.CreatedAt,
		}
	}

	return api.ListCustomers200JSONResponse(resp), nil
}
