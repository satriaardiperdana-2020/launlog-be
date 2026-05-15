package api

import (
	"github.com/labstack/echo/v4"
	"launlog-be/handlers"
	"launlog-be/repository"
)

type LaunlogServer struct {
	AuthHandler        *handlers.AuthHandler
	CustomerHandler    *handlers.CustomerHandler
	TransactionHandler *handlers.TransactionHandler
	// Add other handlers (Service, Payment, Report) as needed
}

func NewLaunlogServer(repo *repository.Repository, jwtSecret string) *LaunlogServer {
	return &LaunlogServer{
		AuthHandler:        handlers.NewAuthHandler(repo.Queries, jwtSecret),
		CustomerHandler:    handlers.NewCustomerHandler(repo.Queries),
		TransactionHandler: handlers.NewTransactionHandler(repo.Queries, repo.DB),
	}
}

// Implement all methods from ServerInterface (generated in internal/openapi.openapi.go)
func (s *LaunlogServer) Register(ctx echo.Context) error {
	return s.AuthHandler.Register(ctx)
}
func (s *LaunlogServer) Login(ctx echo.Context) error {
	return s.AuthHandler.Login(ctx)
}
func (s *LaunlogServer) ListCustomers(ctx echo.Context, params ListCustomersParams) error {
	return s.CustomerHandler.ListCustomers(ctx, params)
}
func (s *LaunlogServer) CreateCustomer(ctx echo.Context) error {
	return s.CustomerHandler.CreateCustomer(ctx)
}
func (s *LaunlogServer) GetCustomer(ctx echo.Context, id int64) error {
	return s.CustomerHandler.GetCustomer(ctx, id)
}
func (s *LaunlogServer) CreateTransaction(ctx echo.Context) error {
	return s.TransactionHandler.CreateTransaction(ctx)
}

// Stub implementations for other required methods (remove "not implemented" when ready)
func (s *LaunlogServer) ListServices(ctx echo.Context, params ListServicesParams) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) ListServiceCategories(ctx echo.Context) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) ListPaymentMethods(ctx echo.Context) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) ListTransactions(ctx echo.Context, params ListTransactionsParams) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) AddPayment(ctx echo.Context, id int64) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) CreateExpenditure(ctx echo.Context) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) GetDashboard(ctx echo.Context) error {
	return echo.ErrNotImplemented
}
func (s *LaunlogServer) ProfitLossReport(ctx echo.Context, params ProfitLossReportParams) error {
	return echo.ErrNotImplemented
}
