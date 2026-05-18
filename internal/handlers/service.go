package handlers

import (
	"context"
	"github.com/satriaardiperdana-2020/launlog-be/internal/helper"
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
	categories, err := h.Queries.ListServiceCategories(ctx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
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
	}
	return api.ListServiceCategories200JSONResponse(resp), nil
}

func (h *ServiceHandler) GetServiceCategory(ctx context.Context, req api.GetServiceCategoryRequestObject) (api.GetServiceCategoryResponseObject, error) {
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
	return api.GetServiceCategory200JSONResponse(resp), nil
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
	services, err := h.Queries.ListServices(ctx)
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	resp := make([]api.ServiceWithCategory, len(services))
	for i, svc := range services {
		desc := ""
		if svc.Description.Valid {
			desc = svc.Description.String
		}

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

func (h *ServiceHandler) GetService(ctx context.Context, req api.GetServiceRequestObject) (api.GetServiceResponseObject, error) {
	svc, err := h.Queries.GetServiceByID(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Service not found")
	}
	desc := ""
	if svc.Description.Valid {
		desc = svc.Description.String
	}
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
	return api.GetService200JSONResponse(resp), nil
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

func (h *ServiceHandler) CreateService(ctx context.Context, req api.CreateServiceRequestObject) (api.CreateServiceResponseObject, error) {
	desc := pgtype.Text{}
	if req.Body.Description != nil {
		desc.String = *req.Body.Description
		desc.Valid = true
	}
	/*minQty := float64(1)
	if req.Body.MinQuantity != nil {
		minQty = float64(*req.Body.MinQuantity)
	}*/
	unit := "kg"
	if req.Body.Unit != nil {
		unit = *req.Body.Unit
	}

	priceNumeric := pgtype.Numeric{}
	priceNumeric.Scan(req.Body.Price)
	minQtyNumeric := pgtype.Numeric{}
	minQtyNumeric.Scan(req.Body.MinQuantity)

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
	priceNumeric := helper.Float32ToNumeric(req.Body.Price)
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

// SoftDeleteService implements strict server interface
func (h *ServiceHandler) SoftDeleteService(ctx context.Context, req api.SoftDeleteServiceRequestObject) (api.SoftDeleteServiceResponseObject, error) {
	// Panggil query soft delete
	svc, err := h.Queries.SoftDeleteService(ctx, int64(req.Id))
	if err != nil {
		return nil, echo.NewHTTPError(http.StatusNotFound, "Service not found or already deleted")
	}

	// Konversi data untuk response
	desc := ""
	if svc.Description.Valid {
		desc = svc.Description.String
	}

	// Konversi pgtype.Numeric ke float32
	price := helper.NumericToFloat32(svc.Price)
	minQty := helper.NumericToFloat32(svc.MinQuantity)

	id := int(svc.ID)
	categoryId := int(svc.CategoryID)

	resp := api.Service{
		Id:          &id,
		CategoryId:  &categoryId,
		Name:        &svc.Name,
		Price:       &price,
		Estimation:  &svc.Estimation,
		MinQuantity: &minQty,
		Unit:        &svc.Unit,
		Description: &desc,
		IsActive:    &svc.IsActive,
		//SortOrder:   &svc.SortOrder,
		CreatedAt: &svc.CreatedAt,
		UpdatedAt: &svc.UpdatedAt,
	}

	return api.SoftDeleteService200JSONResponse(resp), nil
}
