package handlers

import (
	"context"
	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
)

type LaunlogServer struct {
	Queries            *postgresql.Queries
	JWTSecret          []byte
	AuthHandler        *AuthHandler
	CustomerHandler    *CustomerHandler
	ServiceHandler     *ServiceHandler
	TransactionHandler *TransactionHandler
	ExpenseHandler     *ExpenseHandler
	// Add other handlers (Service, Payment, Report) as needed
}

/*func NewLaunlogServer(repo *repository.Repository, jwtSecret string) *LaunlogServer {
	return &LaunlogServer{
		AuthHandler:        handlers.NewAuthHandler(repo.Queries, jwtSecret),
		CustomerHandler:    handlers.NewCustomerHandler(repo.Queries),
		TransactionHandler: handlers.NewTransactionHandler(repo.Queries, repo.DB),
	}
}*/

// Implement all methods from ServerInterface (generated in internal/openapi.openapi.go)
func (s *LaunlogServer) Register(ctx context.Context, req api.RegisterRequestObject) (api.RegisterResponseObject, error) {
	return s.AuthHandler.Register(ctx, req)
}

func (s *LaunlogServer) Login(ctx context.Context, req api.LoginRequestObject) (api.LoginResponseObject, error) {
	return s.AuthHandler.Login(ctx, req)
}

// ---------- Customers ----------
func (s *LaunlogServer) CreateCustomer(ctx context.Context, req api.CreateCustomerRequestObject) (api.CreateCustomerResponseObject, error) {
	return s.CustomerHandler.CreateCustomer(ctx, req)
}

func (s *LaunlogServer) ListCustomers(ctx context.Context, req api.ListCustomersRequestObject) (api.ListCustomersResponseObject, error) {
	return s.CustomerHandler.ListCustomers(ctx, req)
}

func (s *LaunlogServer) GetCustomerById(ctx context.Context, req api.GetCustomerByIdRequestObject) (api.GetCustomerByIdResponseObject, error) {
	return s.CustomerHandler.GetCustomerById(ctx, req)
}

func (s *LaunlogServer) UpdateCustomer(ctx context.Context, req api.UpdateCustomerRequestObject) (api.UpdateCustomerResponseObject, error) {
	return s.CustomerHandler.UpdateCustomer(ctx, req)
}

func (s *LaunlogServer) SoftDeleteCustomer(ctx context.Context, req api.SoftDeleteCustomerRequestObject) (api.SoftDeleteCustomerResponseObject, error) {
	return s.CustomerHandler.SoftDeleteCustomer(ctx, req)
}

// ---------- Services ----------
func (s *LaunlogServer) ListServices(ctx context.Context, req api.ListServicesRequestObject) (api.ListServicesResponseObject, error) {
	// Stub: return empty list (200)
	return s.ServiceHandler.ListServices(ctx, req)
}
func (s *LaunlogServer) CreateService(ctx context.Context, req api.CreateServiceRequestObject) (api.CreateServiceResponseObject, error) {
	return s.ServiceHandler.CreateService(ctx, req)
}

func (s *LaunlogServer) UpdateService(ctx context.Context, req api.UpdateServiceRequestObject) (api.UpdateServiceResponseObject, error) {
	return s.ServiceHandler.UpdateService(ctx, req)
}

func (s *LaunlogServer) GetServiceById(ctx context.Context, req api.GetServiceByIdRequestObject) (api.GetServiceByIdResponseObject, error) {
	return s.ServiceHandler.GetServiceById(ctx, req)
}

func (s *LaunlogServer) GetServiceDetail(ctx context.Context, req api.GetServiceDetailRequestObject) (api.GetServiceDetailResponseObject, error) {
	return s.ServiceHandler.GetServiceDetail(ctx, req)
}
func (s *LaunlogServer) SoftDeleteService(ctx context.Context, req api.SoftDeleteServiceRequestObject) (api.SoftDeleteServiceResponseObject, error) {
	return s.ServiceHandler.SoftDeleteService(ctx, req)
}

// ==================== SERVICE CATEGORIES ====================
func (s *LaunlogServer) CreateServiceCategory(ctx context.Context, req api.CreateServiceCategoryRequestObject) (api.CreateServiceCategoryResponseObject, error) {
	return s.ServiceHandler.CreateServiceCategory(ctx, req)
}

func (s *LaunlogServer) UpdateServiceCategory(ctx context.Context, req api.UpdateServiceCategoryRequestObject) (api.UpdateServiceCategoryResponseObject, error) {
	return s.ServiceHandler.UpdateServiceCategory(ctx, req)
}

func (s *LaunlogServer) ListServiceCategories(ctx context.Context, req api.ListServiceCategoriesRequestObject) (api.ListServiceCategoriesResponseObject, error) {
	return s.ServiceHandler.ListServiceCategories(ctx, req)
}
func (s *LaunlogServer) GetServiceCategoryById(ctx context.Context, req api.GetServiceCategoryByIdRequestObject) (api.GetServiceCategoryByIdResponseObject, error) {
	return s.ServiceHandler.GetServiceCategoryById(ctx, req)
}

func (s *LaunlogServer) SoftDeleteServiceCategory(ctx context.Context, req api.SoftDeleteServiceCategoryRequestObject) (api.SoftDeleteServiceCategoryResponseObject, error) {
	return s.ServiceHandler.SoftDeleteServiceCategory(ctx, req)
}

func (s *LaunlogServer) ListPaymentMethods(ctx context.Context, req api.ListPaymentMethodsRequestObject) (api.ListPaymentMethodsResponseObject, error) {
	// Stub: return empty list (200)
	return api.ListPaymentMethods200JSONResponse([]api.PaymentMethod{}), nil
}

// ---------- Transactions ----------
func (s *LaunlogServer) CreateTransaction(ctx context.Context, req api.CreateTransactionRequestObject) (api.CreateTransactionResponseObject, error) {
	return s.TransactionHandler.CreateTransaction(ctx, req)
}

func (s *LaunlogServer) ListTransactions(ctx context.Context, req api.ListTransactionsRequestObject) (api.ListTransactionsResponseObject, error) {
	return s.TransactionHandler.ListTransactions(ctx, req)
}

func (s *LaunlogServer) CreateExpense(ctx context.Context, req api.CreateExpenseRequestObject) (api.CreateExpenseResponseObject, error) {
	return s.ExpenseHandler.CreateExpense(ctx, req)
}

func (s *LaunlogServer) ListExpenses(ctx context.Context, req api.ListExpensesRequestObject) (api.ListExpensesResponseObject, error) {
	return s.ExpenseHandler.ListExpenses(ctx, req)
}

func (s *LaunlogServer) UpdateExpense(ctx context.Context, req api.UpdateExpenseRequestObject) (api.UpdateExpenseResponseObject, error) {
	return s.ExpenseHandler.UpdateExpense(ctx, req)
}

func (s *LaunlogServer) SoftDeleteExpense(ctx context.Context, req api.SoftDeleteExpenseRequestObject) (api.SoftDeleteExpenseResponseObject, error) {
	return s.ExpenseHandler.SoftDeleteExpense(ctx, req)
}

// GetExpense returns expense by ID
func (s *LaunlogServer) GetExpense(ctx context.Context, req api.GetExpenseRequestObject) (api.GetExpenseResponseObject, error) {
	return s.ExpenseHandler.GetExpense(ctx, req)
}

// ---------- Dashboard & Reports ----------
func (s *LaunlogServer) GetDashboard(ctx context.Context, req api.GetDashboardRequestObject) (api.GetDashboardResponseObject, error) {
	// Stub: return zero dashboard (200)
	var zero float32 = 0.0
	return api.GetDashboard200JSONResponse(api.DashboardResponse{
		TodayIncome:  &zero,
		TodayExpense: &zero,
	}), nil
}

func (s *LaunlogServer) ProfitLossReport(ctx context.Context, req api.ProfitLossReportRequestObject) (api.ProfitLossReportResponseObject, error) {
	// Stub: return empty report (200)
	var zero float32 = 0.0
	return api.ProfitLossReport200JSONResponse(api.ProfitLossReport{
		TotalIncome:  &zero,
		TotalExpense: &zero,
		Profit:       &zero,
		//Details:      []api.ProfitLossReportDetailsItem{},
	}), nil
}

// ==================== HEALTH CHECK ====================
func (s *LaunlogServer) Health(ctx context.Context, req api.HealthRequestObject) (api.HealthResponseObject, error) {
	status := "ok"
	return api.Health200JSONResponse(struct {
		Status *string `json:"status,omitempty"`
	}{Status: &status}), nil
}
func (s *LaunlogServer) AddPayment(ctx context.Context, req api.AddPaymentRequestObject) (api.AddPaymentResponseObject, error) {
	// Stub: return empty Transaction (200)
	return api.AddPayment200JSONResponse(api.Transaction{}), nil
}
func (s *LaunlogServer) UpdateTransaction(ctx context.Context, req api.UpdateTransactionRequestObject) (api.UpdateTransactionResponseObject, error) {
	return s.TransactionHandler.UpdateTransaction(ctx, req)
}

func (s *LaunlogServer) SoftDeleteTransaction(ctx context.Context, req api.SoftDeleteTransactionRequestObject) (api.SoftDeleteTransactionResponseObject, error) {
	return s.TransactionHandler.SoftDeleteTransaction(ctx, req)
}

func (s *LaunlogServer) RestoreTransaction(ctx context.Context, req api.RestoreTransactionRequestObject) (api.RestoreTransactionResponseObject, error) {
	return s.TransactionHandler.RestoreTransaction(ctx, req)
}

func (s *LaunlogServer) GetTransactionReport(ctx context.Context, req api.GetTransactionReportRequestObject) (api.GetTransactionReportResponseObject, error) {
	return s.TransactionHandler.GetTransactionReport(ctx, req)
}
