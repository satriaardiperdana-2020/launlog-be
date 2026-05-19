package handlers

import (
	"context"
	"database/sql"
	"github.com/labstack/gommon/log"
	"github.com/satriaardiperdana-2020/launlog-be/internal/helper"
	"math/big"
	"net/http"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"

	"github.com/satriaardiperdana-2020/launlog-be/internal/api"
	"github.com/satriaardiperdana-2020/launlog-be/internal/repository/postgresql"
)

type ServiceHandler struct {
	Queries *postgresql.Queries
}

// ==================== SERVICE CATEGORIES ====================

func (h *ServiceHandler) ListServiceCategories(ctx context.Context, req api.ListServiceCategoriesRequestObject) (api.ListServiceCategoriesResponseObject, error) {
	log.Info("🔥🔥 ListServiceCategories HANDLER called") // <- tambah
	categories, err := h.Queries.ListServiceCategories(ctx)
	if err != nil {
		log.Printf("ERROR ListServiceCategories: %v", err) // <- tambah
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	log.Printf("Found %d categories", len(categories))
	resp := make([]api.ServiceCategory, len(categories))
	for i, cat := range categories {
		desc := ""
		if cat.Description.Valid {
			desc = cat.Description.String
		}
		id := int(cat.ID) // ← konversi int64 ke int
		resp[i] = api.ServiceCategory{
			Id:          &id,
			Name:        &cat.Name,
			Description: &desc,
		}
		log.Info(" res:  ", resp)
	}
	return api.ListServiceCategories200JSONResponse(resp), nil
}

func (h *ServiceHandler) GetServiceCategoryById(ctx context.Context, req api.GetServiceCategoryByIdRequestObject) (api.GetServiceCategoryByIdResponseObject, error) {
	cat, err := h.Queries.GetServiceCategoryByID(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}
	desc := ""
	if cat.Description.Valid {
		desc = cat.Description.String
	}
	id := int(cat.ID) // ← konversi int64 ke int
	resp := api.ServiceCategory{
		Id:          &id,
		Name:        &cat.Name,
		Description: &desc,
	}
	return api.GetServiceCategoryById200JSONResponse(resp), nil
}

func (h *ServiceHandler) CreateServiceCategory(ctx context.Context, req api.CreateServiceCategoryRequestObject) (api.CreateServiceCategoryResponseObject, error) {
	desc := pgtype.Text{}
	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}
	cat, err := h.Queries.CreateServiceCategory(ctx, postgresql.CreateServiceCategoryParams{
		Name:        req.Body.Name,
		Description: desc,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	descStr := ""
	if cat.Description.Valid {
		descStr = cat.Description.String
	}
	id := int(cat.ID) // ← konversi int64 ke int
	sortOrder := int(cat.SortOrder)
	resp := api.ServiceCategory{
		Id:          &id,
		Name:        &cat.Name,
		Description: &descStr,
		IsActive:    &cat.IsActive,
		SortOrder:   &sortOrder,
	}
	return api.CreateServiceCategory201JSONResponse(resp), nil
}

func (h *ServiceHandler) UpdateServiceCategory(ctx context.Context, req api.UpdateServiceCategoryRequestObject) (api.UpdateServiceCategoryResponseObject, error) {
	desc := pgtype.Text{}
	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}
	sortOrder := int32(0)
	if req.Body.SortOrder != nil {
		sortOrder = int32(*req.Body.SortOrder)
	}
	cat, err := h.Queries.UpdateServiceCategory(ctx, postgresql.UpdateServiceCategoryParams{
		ID:          int64(req.Id),
		Name:        req.Body.Name,
		Description: desc,
		SortOrder:   sortOrder,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	descStr := ""
	if cat.Description.Valid {
		descStr = cat.Description.String
	}
	id := int(cat.ID)                // ← konversi int64 ke int
	sortOrder2 := int(cat.SortOrder) // konversi ke int
	resp := api.ServiceCategory{
		Id:          &id,
		Name:        &cat.Name,
		Description: &descStr,
		IsActive:    &cat.IsActive,
		SortOrder:   &sortOrder2,
	}
	return api.UpdateServiceCategory200JSONResponse(resp), nil
}

func (h *ServiceHandler) SoftDeleteServiceCategory(ctx context.Context, req api.SoftDeleteServiceCategoryRequestObject) (api.SoftDeleteServiceCategoryResponseObject, error) {
	_, err := h.Queries.SoftDeleteServiceCategory(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Category not found")
	}
	msg := "Deleted successfully"
	/*	resp := api.ServiceCategory{
		Id:          &id,
		Name:        &cat.Name,
		Description: &descStr,
		IsActive:    &cat.IsActive,
		SortOrder:   &sortOrder2,
	}*/
	return api.SoftDeleteServiceCategory200JSONResponse{Message: &msg}, nil
}

// ==================== SERVICES ====================

func (h *ServiceHandler) ListServices(ctx context.Context, req api.ListServicesRequestObject) (api.ListServicesResponseObject, error) {
	var categoryID sql.NullInt64
	if req.Params.CategoryId != nil {
		categoryID.Int64 = int64(*req.Params.CategoryId)
		categoryID.Valid = true
		log.Printf("Filtering by category ID: %d", categoryID.Int64)
	} else {
		log.Printf("No category filter applied")
	}

	var search sql.NullString
	if req.Params.Search != nil && *req.Params.Search != "" {
		search.String = *req.Params.Search
		search.Valid = true
		log.Printf("Searching for: %s", search.String)
	} else {
		log.Printf("No search filter applied")
	}

	services, err := h.Queries.ListServices(ctx, postgresql.ListServicesParams{
		CategoryID: pgtype.Int8(categoryID),
		Search:     pgtype.Text(search),
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	resp := make([]api.ServiceWithCategory, len(services))
	for i, svc := range services {
		desc := svc.Description

		//id := int(svc.ID)
		//categoryId := int(svc.CategoryID)
		price := helper.NumericToFloat32(svc.Price)
		minQty := helper.NumericToFloat32(svc.MinQuantity)
		resp[i] = api.ServiceWithCategory{
			//Id:           &id,
			//CategoryId:   &categoryId,
			CategoryName: &svc.CategoryName,
			Name:         &svc.ServiceName,
			Price:        &price,
			Estimation:   &svc.Estimation,
			MinQuantity:  &minQty,
			Unit:         &svc.Unit,
			Description:  &desc,
			IsActive:     &svc.IsActive,
		}
	}
	return api.ListServices200JSONResponse(resp), nil
}

func (h *ServiceHandler) GetServiceById(ctx context.Context, req api.GetServiceByIdRequestObject) (api.GetServiceByIdResponseObject, error) {
	svc, err := h.Queries.GetServiceById(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Service not found")
	}
	desc := svc.Description
	//id := int(svc.ID)
	//categoryId := int(svc.CategoryID)
	price := helper.NumericToFloat32(svc.Price)
	minQty := helper.NumericToFloat32(svc.MinQuantity)
	resp := api.Service{
		//Id:           &id,
		//CategoryId:   &categoryId,
		//CategoryName: &svc.CategoryName,
		Name:        &svc.ServiceName,
		Price:       &price,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQty,
		Unit:        &svc.Unit,
		Description: &desc,
		IsActive:    &svc.IsActive,
	}
	return api.GetServiceById200JSONResponse(resp), nil
}

func (h *ServiceHandler) GetServiceDetail(ctx context.Context, req api.GetServiceDetailRequestObject) (api.GetServiceDetailResponseObject, error) {
	detail, err := h.Queries.GetServiceDetail(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Service not found")
	}
	id := int(detail.ID)
	categoryId := int(detail.CategoryID)
	price := helper.NumericToFloat32(detail.Price)
	minQty := helper.NumericToFloat32(detail.MinQuantity)
	resp := api.ServiceDetail{
		Id:           &id,
		CategoryId:   &categoryId,
		CategoryName: &detail.CategoryName,
		Name:         &detail.Name,
		Price:        &price,
		Estimation:   &detail.Estimation,
		MinQuantity:  &minQty,
		Unit:         &detail.Unit,
		Description:  &detail.Description,
		IsActive:     &detail.IsActive,
		CreatedAt:    &detail.CreatedAt,
		UpdatedAt:    &detail.UpdatedAt,
	}
	return api.GetServiceDetail200JSONResponse(resp), nil
}

/*func (h *ServiceHandler) CreateService(ctx context.Context, req api.CreateServiceRequestObject) (api.CreateServiceResponseObject, error) {
	desc := pgtype.Text{}
	if req.Body.Price == nil {
		msg := "Price is required"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}
	if *req.Body.Price <= 0 {
		msg := "Price must be greater than 0"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}

	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}
	minQty := float64(1)
	if req.Body.MinQuantity != nil {
		minQty = float64(*req.Body.MinQuantity)
	}
	minQtyNumeric := pgtype.Numeric{}
	if err := minQtyNumeric.Scan(minQty); err != nil {
		msg := "Invalid min quantity format"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}
	// Pastikan valid
	minQtyNumeric.Valid = true

	unit := "kg"
	if req.Body.Unit != nil && *req.Body.Unit != "" {
		unit = *req.Body.Unit
	}

	priceNumeric := pgtype.Numeric{}
	// Cara 1: Konversi ke float64 dulu, lalu Scan
	priceFloat64 := float64(*req.Body.Price)
	if err := priceNumeric.Scan(priceFloat64); err != nil {
		// Cara 2: Jika Scan gagal, konversi manual
		priceNumeric.Int = big.NewInt(int64(priceFloat64))
		priceNumeric.Exp = 0
		priceNumeric.Valid = true
	}

	// Pastikan Valid = true
	if !priceNumeric.Valid {
		msg := "Failed to convert price value"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}

	svc, err := h.Queries.CreateService(ctx, postgresql.CreateServiceParams{
		CategoryID:  int64(req.Body.CategoryId),
		Name:        req.Body.Name,
		Price:       priceNumeric,
		Estimation:  req.Body.Estimation,
		MinQuantity: minQtyNumeric,
		Unit:        unit,
		Description: desc,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	price := helper.NumericToFloat32(svc.Price)
	minQtyResp := helper.NumericToFloat32(svc.MinQuantity)
	descStr := ""
	if svc.Description.Valid {
		descStr = svc.Description.String
	}
	id := int(svc.ID)
	categoryId := int(svc.CategoryID)
	resp := api.Service{
		Id:          &id,
		CategoryId:  &categoryId,
		Name:        &svc.Name,
		Price:       &price,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQtyResp,
		Unit:        &svc.Unit,
		Description: &descStr,
		IsActive:    &svc.IsActive,
		CreatedAt:   &svc.CreatedAt,
		UpdatedAt:   &svc.UpdatedAt,
	}
	return api.CreateService201JSONResponse(resp), nil
}
*/

func (h *ServiceHandler) CreateService(ctx context.Context, req api.CreateServiceRequestObject) (api.CreateServiceResponseObject, error) {
	// ==================== VALIDASI PRICE ====================
	if req.Body.Price == nil {
		msg := "Price is required"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}
	if *req.Body.Price <= 0 {
		msg := "Price must be greater than 0"
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}

	// ==================== KONVERSI PRICE ====================
	priceNumeric := pgtype.Numeric{}

	// Cara manual yang lebih andal
	priceValue := *req.Body.Price
	priceInt := int64(priceValue)

	// Set nilai ke pgtype.Numeric
	priceNumeric.Int = big.NewInt(priceInt)
	priceNumeric.Exp = 0
	priceNumeric.Valid = true

	// ==================== MIN_QUANTITY (WAJIB, DEFAULT 1) ====================
	minQty := float64(1) // ← default 1
	if req.Body.MinQuantity != nil {
		minQty = float64(*req.Body.MinQuantity)
	}
	minQtyNumeric := pgtype.Numeric{}
	minQtyInt := int64(minQty)
	minQtyNumeric.Int = big.NewInt(minQtyInt)
	minQtyNumeric.Exp = 0
	minQtyNumeric.Valid = true

	// ==================== UNIT (DEFAULT 'kg') ====================
	unit := "kg"
	if req.Body.Unit != nil && *req.Body.Unit != "" {
		unit = *req.Body.Unit
	}

	// ==================== DESCRIPTION (OPTIONAL) ====================
	desc := pgtype.Text{}
	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}

	// ==================== SORT_ORDER (DEFAULT 0) ====================
	/*	sortOrder := int32(0)
		if req.Body.SortOrder != nil {
			sortOrder = int32(*req.Body.SortOrder)
		}*/

	// ==================== INSERT DATABASE ====================
	svc, err := h.Queries.CreateService(ctx, postgresql.CreateServiceParams{
		CategoryID:  int64(req.Body.CategoryId),
		Name:        req.Body.Name,
		Price:       priceNumeric,
		Estimation:  req.Body.Estimation,
		MinQuantity: minQtyNumeric, // ← PASTIKAN VALID = true
		Unit:        unit,
		Description: desc,
		//SortOrder:   sortOrder,
	})
	if err != nil {
		msg := "Failed to create service: " + err.Error()
		return api.CreateService400JSONResponse{Message: &msg}, nil
	}

	// ==================== RESPONSE ====================

	priceResp := helper.NumericToFloat32(svc.Price)
	minQtyResp := helper.NumericToFloat32(svc.MinQuantity)

	descStr := ""
	if svc.Description.Valid {
		descStr = svc.Description.String
	}

	id := int(svc.ID)
	categoryId := int(svc.CategoryID)

	resp := api.Service{
		Id:          &id,
		CategoryId:  &categoryId,
		Name:        &svc.Name,
		Price:       &priceResp,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQtyResp,
		Unit:        &svc.Unit,
		Description: &descStr,
		IsActive:    &svc.IsActive,
		//SortOrder:   &svc.SortOrder,
		CreatedAt: &svc.CreatedAt,
		UpdatedAt: &svc.UpdatedAt,
	}
	return api.CreateService201JSONResponse(resp), nil
}

// Helper function untuk konversi pgtype.Numeric ke float32
/*func numericToFloat32(n pgtype.Numeric) float32 {
	if !n.Valid {
		return 0
	}
	f, _ := n.Float64Value()
	return float32(f)
}*/

func (h *ServiceHandler) UpdateService(ctx context.Context, req api.UpdateServiceRequestObject) (api.UpdateServiceResponseObject, error) {
	desc := pgtype.Text{}

	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}
	minQty := float64(1)
	if req.Body.MinQuantity != nil {
		minQty = float64(*req.Body.MinQuantity)
	}
	unit := "kg"
	if req.Body.Unit != nil {
		unit = *req.Body.Unit
	}

	// ✅ Gunakan helper Float64ToNumeric
	priceNumeric := helper.Float32ToNumeric(*req.Body.Price)
	minQtyNumeric := helper.Float32ToNumeric(float32(minQty))
	svc, err := h.Queries.UpdateService(ctx, postgresql.UpdateServiceParams{
		ID:          int64(req.Id),
		CategoryID:  int64(req.Body.CategoryId),
		Name:        req.Body.Name,
		Price:       priceNumeric,
		Estimation:  req.Body.Estimation,
		MinQuantity: minQtyNumeric,
		Unit:        unit,
		Description: desc,
	})
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	price := helper.NumericToFloat32(svc.Price)
	minQtyFloat64 := helper.NumericToFloat32(svc.MinQuantity)
	descStr := ""
	if svc.Description.Valid {
		descStr = svc.Description.String
	}
	id := int(svc.ID)
	categoryId := int(svc.CategoryID)
	resp := api.Service{
		Id:          &id,
		CategoryId:  &categoryId,
		Name:        &svc.Name,
		Price:       &price,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQtyFloat64,
		Unit:        &svc.Unit,
		Description: &descStr,
		IsActive:    &svc.IsActive,
		CreatedAt:   &svc.CreatedAt,
		UpdatedAt:   &svc.UpdatedAt,
	}
	return api.UpdateService200JSONResponse(resp), nil
}

// SoftDeleteService soft deletes a service (sets is_active = false)
func (h *ServiceHandler) SoftDeleteService(ctx context.Context, req api.SoftDeleteServiceRequestObject) (api.SoftDeleteServiceResponseObject, error) {
	log.Printf("🔵 SoftDeleteService called for ID: %d", req.Id)

	// Call database query
	svc, err := h.Queries.SoftDeleteService(ctx, int64(req.Id))
	if err != nil {
		log.Printf("❌ Soft delete error: %v", err)
		msg := "Service not found or already deleted"
		return api.SoftDeleteService404JSONResponse{Message: &msg}, nil
	}

	log.Printf("✅ Service %d soft deleted successfully", req.Id)

	// Convert response
	price := helper.NumericToFloat32(svc.Price)
	minQty := helper.NumericToFloat32(svc.MinQuantity)
	id := int(svc.ID)
	categoryId := int(svc.CategoryID)
	sortOrder := int(svc.SortOrder)
	descStr := ""
	if svc.Description.Valid {
		descStr = svc.Description.String
	}

	resp := api.Service{
		Id:          &id,
		CategoryId:  &categoryId,
		Name:        &svc.Name,
		Price:       &price,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQty,
		Unit:        &svc.Unit,
		Description: &descStr,
		IsActive:    &svc.IsActive,
		SortOrder:   &sortOrder,
		CreatedAt:   &svc.CreatedAt,
		UpdatedAt:   &svc.UpdatedAt,
	}

	return api.SoftDeleteService200JSONResponse(resp), nil
}
