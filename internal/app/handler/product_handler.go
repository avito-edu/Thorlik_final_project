package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/Thorlik/marketplace/internal/app/dto"
	"github.com/Thorlik/marketplace/internal/model/domain"
	"github.com/Thorlik/marketplace/internal/service"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
)

type ProductHandler struct {
	service service.ProductService
	logger  *zap.Logger
}

func NewProductHandler(service service.ProductService, logger *zap.Logger) *ProductHandler {
	return &ProductHandler{
		service: service,
		logger:  logger,
	}
}

func (h *ProductHandler) CreateProduct(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)
	h.logger.Info("Handling create product request")

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	product := &domain.Product{
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Category:    req.Category,
	}

	createdProduct, err := h.service.Create(r.Context(), claims.UserID, product)
	if err != nil {
		if errors.Is(err, service.ErrInvalidProductTitle) || errors.Is(err, service.ErrInvalidProductPrice) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		h.logger.Error("Failed to create product", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create product", h.logger)
		return
	}

	response := dto.ProductResponse{
		ID:          createdProduct.ID,
		SellerID:    createdProduct.SellerID,
		Title:       createdProduct.Title,
		Description: createdProduct.Description,
		Price:       createdProduct.Price,
		Quantity:    createdProduct.Quantity,
		Category:    createdProduct.Category,
		Status:      string(createdProduct.Status),
		CreatedAt:   createdProduct.CreatedAt,
	}

	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

func (h *ProductHandler) GetProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID", h.logger)
		return
	}

	product, err := h.service.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			RespondWithError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get product", h.logger)
		return
	}

	response := dto.ProductResponse{
		ID:          product.ID,
		SellerID:    product.SellerID,
		Title:       product.Title,
		Description: product.Description,
		Price:       product.Price,
		Quantity:    product.Quantity,
		Category:    product.Category,
		Status:      string(product.Status),
		CreatedAt:   product.CreatedAt,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

func (h *ProductHandler) GetAllProducts(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page, _ := strconv.Atoi(query.Get("page"))
	pageSize, _ := strconv.Atoi(query.Get("page_size"))
	minPrice, _ := strconv.ParseFloat(query.Get("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(query.Get("max_price"), 64)
	sellerID, _ := strconv.ParseInt(query.Get("seller_id"), 10, 64)

	filter := &dto.ProductFilter{
		Category:  query.Get("category"),
		MinPrice:  minPrice,
		MaxPrice:  maxPrice,
		SellerID:  sellerID,
		Status:    query.Get("status"),
		Page:      page,
		PageSize:  pageSize,
		SortBy:    query.Get("sort_by"),
		SortOrder: query.Get("sort_order"),
	}

	if filter.Status == "" {
		filter.Status = "approved"
	}

	products, totalCount, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products", h.logger)
		return
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ProductResponse{
			ID:          p.ID,
			SellerID:    p.SellerID,
			Title:       p.Title,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Category:    p.Category,
			Status:      string(p.Status),
			CreatedAt:   p.CreatedAt,
		})
	}

	response := dto.ProductListResponse{
		Products:   responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

func (h *ProductHandler) GetMyProducts(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	products, err := h.service.GetBySeller(r.Context(), claims.UserID, page, pageSize)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products", h.logger)
		return
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ProductResponse{
			ID:          p.ID,
			SellerID:    p.SellerID,
			Title:       p.Title,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Category:    p.Category,
			Status:      string(p.Status),
			CreatedAt:   p.CreatedAt,
		})
	}

	RespondWithJSON(w, http.StatusOK, responses, h.logger)
}

func (h *ProductHandler) UpdateProduct(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID", h.logger)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	product := &domain.Product{
		ID:          id,
		Title:       req.Title,
		Description: req.Description,
		Price:       req.Price,
		Quantity:    req.Quantity,
		Category:    req.Category,
	}

	updatedProduct, err := h.service.Update(r.Context(), claims.UserID, claims.Role, product)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			RespondWithError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}
		if errors.Is(err, service.ErrNotProductOwner) {
			RespondWithError(w, http.StatusForbidden, err.Error(), h.logger)
			return
		}
		if errors.Is(err, service.ErrInvalidProductTitle) || errors.Is(err, service.ErrInvalidProductPrice) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update product", h.logger)
		return
	}

	response := dto.ProductResponse{
		ID:          updatedProduct.ID,
		SellerID:    updatedProduct.SellerID,
		Title:       updatedProduct.Title,
		Description: updatedProduct.Description,
		Price:       updatedProduct.Price,
		Quantity:    updatedProduct.Quantity,
		Category:    updatedProduct.Category,
		Status:      string(updatedProduct.Status),
		CreatedAt:   updatedProduct.CreatedAt,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

func (h *ProductHandler) DeleteProduct(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID", h.logger)
		return
	}

	if err := h.service.Delete(r.Context(), claims.UserID, claims.Role, id); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			RespondWithError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}
		if errors.Is(err, service.ErrNotProductOwner) {
			RespondWithError(w, http.StatusForbidden, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to delete product", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Product deleted successfully"}, h.logger)
}

func (h *ProductHandler) ModerateProduct(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid product ID", h.logger)
		return
	}

	var req dto.ModerateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	status := domain.ProductStatus(req.Status)
	if status != domain.ProductStatusApproved && status != domain.ProductStatusRejected {
		RespondWithError(w, http.StatusBadRequest, "Invalid status", h.logger)
		return
	}

	if err := h.service.UpdateStatus(r.Context(), id, status); err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			RespondWithError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to moderate product", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Product status updated successfully"}, h.logger)
}

func (h *ProductHandler) GetPendingProducts(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	filter := &dto.ProductFilter{
		Status:   "pending",
		Page:     page,
		PageSize: pageSize,
	}

	products, totalCount, err := h.service.GetAll(r.Context(), filter)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get products", h.logger)
		return
	}

	var responses []dto.ProductResponse
	for _, p := range products {
		responses = append(responses, dto.ProductResponse{
			ID:          p.ID,
			SellerID:    p.SellerID,
			Title:       p.Title,
			Description: p.Description,
			Price:       p.Price,
			Quantity:    p.Quantity,
			Category:    p.Category,
			Status:      string(p.Status),
			CreatedAt:   p.CreatedAt,
		})
	}

	response := dto.ProductListResponse{
		Products:   responses,
		TotalCount: totalCount,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}
