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

type OrderHandler struct {
	service service.OrderService
	logger  *zap.Logger
}

func NewOrderHandler(service service.OrderService, logger *zap.Logger) *OrderHandler {
	return &OrderHandler{
		service: service,
		logger:  logger,
	}
}

// CreateOrder godoc
// @Summary Create a new order
// @Description Create an order with specified items
// @Tags Orders
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.CreateOrderRequest true "Order items"
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /orders [post]
func (h *OrderHandler) CreateOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)
	h.logger.Info("Creating order request")

	var req dto.CreateOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, item := range req.Items {
		items = append(items, domain.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}

	order, err := h.service.CreateOrder(r.Context(), claims.UserID, items)
	if err != nil {
		if errors.Is(err, service.ErrCartEmpty) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) || errors.Is(err, service.ErrProductUnavailable) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		h.logger.Error("Failed to create order", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create order", h.logger)
		return
	}

	response := h.orderToResponse(order)
	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// CreateOrderFromCart godoc
// @Summary Create order from cart
// @Description Create an order from all items in user's cart
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Success 201 {object} dto.OrderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /orders/from-cart [post]
func (h *OrderHandler) CreateOrderFromCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)
	h.logger.Info("Creating order from cart request")

	order, err := h.service.CreateOrderFromCart(r.Context(), claims.UserID)
	if err != nil {
		if errors.Is(err, service.ErrCartEmpty) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		if errors.Is(err, service.ErrInsufficientStock) || errors.Is(err, service.ErrProductUnavailable) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		h.logger.Error("Failed to create order from cart", zap.Error(err))
		RespondWithError(w, http.StatusInternalServerError, "Failed to create order", h.logger)
		return
	}

	response := h.orderToResponse(order)
	RespondWithJSON(w, http.StatusCreated, response, h.logger)
}

// GetOrder godoc
// @Summary Get order by ID
// @Description Get order details by ID (owner or admin)
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Success 200 {object} dto.OrderResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /orders/{id} [get]
func (h *OrderHandler) GetOrder(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID", h.logger)
		return
	}

	order, err := h.service.GetOrderByID(r.Context(), claims.UserID, claims.Role, id)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			RespondWithError(w, http.StatusNotFound, "Order not found", h.logger)
			return
		}
		if errors.Is(err, service.ErrNotOrderOwner) {
			RespondWithError(w, http.StatusForbidden, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to get order", h.logger)
		return
	}

	response := h.orderToResponse(order)
	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetMyOrders godoc
// @Summary Get my orders
// @Description Get list of orders for authenticated user
// @Tags Orders
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} dto.OrderResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /orders [get]
func (h *OrderHandler) GetMyOrders(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	orders, err := h.service.GetUserOrders(r.Context(), claims.UserID, page, pageSize)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get orders", h.logger)
		return
	}

	var responses []dto.OrderResponse
	for _, o := range orders {
		responses = append(responses, *h.orderToResponse(o))
	}

	RespondWithJSON(w, http.StatusOK, responses, h.logger)
}

// GetAllOrders godoc
// @Summary Get all orders (Admin)
// @Description Get list of all orders (admin only)
// @Tags Admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Success 200 {array} dto.OrderResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/orders [get]
func (h *OrderHandler) GetAllOrders(w http.ResponseWriter, r *http.Request) {
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	pageSize, _ := strconv.Atoi(r.URL.Query().Get("page_size"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}

	orders, err := h.service.GetAllOrders(r.Context(), page, pageSize)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get orders", h.logger)
		return
	}

	var responses []dto.OrderResponse
	for _, o := range orders {
		responses = append(responses, *h.orderToResponse(o))
	}

	RespondWithJSON(w, http.StatusOK, responses, h.logger)
}

// UpdateOrderStatus godoc
// @Summary Update order status (Admin)
// @Description Update order status (admin only)
// @Tags Admin
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Order ID"
// @Param request body dto.UpdateOrderStatusRequest true "New status"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /admin/orders/{id}/status [put]
func (h *OrderHandler) UpdateOrderStatus(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid order ID", h.logger)
		return
	}

	var req dto.UpdateOrderStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	status := domain.OrderStatus(req.Status)

	if err := h.service.UpdateOrderStatus(r.Context(), id, status); err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			RespondWithError(w, http.StatusNotFound, "Order not found", h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to update order status", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Order status updated successfully"}, h.logger)
}

// ProcessPayment godoc
// @Summary Process payment
// @Description Process payment for an order
// @Tags Payments
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.PaymentRequest true "Payment details"
// @Success 200 {object} dto.PaymentResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 403 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /payments [post]
func (h *OrderHandler) ProcessPayment(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	var req dto.PaymentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	result, err := h.service.ProcessPayment(r.Context(), claims.UserID, req.OrderID, req.PaymentMethod)
	if err != nil {
		if errors.Is(err, service.ErrOrderNotFound) {
			RespondWithError(w, http.StatusNotFound, "Order not found", h.logger)
			return
		}
		if errors.Is(err, service.ErrNotOrderOwner) {
			RespondWithError(w, http.StatusForbidden, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to process payment", h.logger)
		return
	}

	response := dto.PaymentResponse{
		PaymentID:     result.PaymentID,
		Status:        result.Status,
		Message:       result.Message,
		TransactionID: result.TransactionID,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// AddToCart godoc
// @Summary Add item to cart
// @Description Add a product to user's shopping cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body dto.AddToCartRequest true "Cart item"
// @Success 200 {object} dto.CartItemResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 404 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cart [post]
func (h *OrderHandler) AddToCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	var req dto.AddToCartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	cartItem, err := h.service.AddToCart(r.Context(), claims.UserID, req.ProductID, req.Quantity)
	if err != nil {
		if errors.Is(err, service.ErrProductNotFound) {
			RespondWithError(w, http.StatusNotFound, "Product not found", h.logger)
			return
		}
		if errors.Is(err, service.ErrProductUnavailable) || errors.Is(err, service.ErrInsufficientStock) {
			RespondWithError(w, http.StatusBadRequest, err.Error(), h.logger)
			return
		}
		RespondWithError(w, http.StatusInternalServerError, "Failed to add to cart", h.logger)
		return
	}

	response := dto.CartItemResponse{
		ID:        cartItem.ID,
		ProductID: cartItem.ProductID,
		Quantity:  cartItem.Quantity,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// GetCart godoc
// @Summary Get shopping cart
// @Description Get user's shopping cart with items
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.CartResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cart [get]
func (h *OrderHandler) GetCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	items, totalPrice, err := h.service.GetCart(r.Context(), claims.UserID)
	if err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to get cart", h.logger)
		return
	}

	var itemResponses []dto.CartItemResponse
	totalItems := 0
	for _, item := range items {
		cartItem := dto.CartItemResponse{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
		if item.Product != nil {
			cartItem.Product = dto.ProductResponse{
				ID:          item.Product.ID,
				SellerID:    item.Product.SellerID,
				Title:       item.Product.Title,
				Description: item.Product.Description,
				Price:       item.Product.Price,
				Quantity:    item.Product.Quantity,
				Category:    item.Product.Category,
				Status:      string(item.Product.Status),
			}
		}
		itemResponses = append(itemResponses, cartItem)
		totalItems += item.Quantity
	}

	response := dto.CartResponse{
		Items:      itemResponses,
		TotalItems: totalItems,
		TotalPrice: totalPrice,
	}

	RespondWithJSON(w, http.StatusOK, response, h.logger)
}

// UpdateCartItem godoc
// @Summary Update cart item quantity
// @Description Update quantity of an item in cart
// @Tags Cart
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Param request body dto.UpdateCartItemRequest true "New quantity"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cart/{id} [put]
func (h *OrderHandler) UpdateCartItem(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid cart item ID", h.logger)
		return
	}

	var req dto.UpdateCartItemRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid request body", h.logger)
		return
	}

	if err := h.service.UpdateCartItem(r.Context(), claims.UserID, id, req.Quantity); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to update cart item", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Cart item updated successfully"}, h.logger)
}

// RemoveFromCart godoc
// @Summary Remove item from cart
// @Description Remove an item from shopping cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Param id path int true "Cart item ID"
// @Success 200 {object} dto.SuccessResponse
// @Failure 400 {object} dto.ErrorResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cart/{id} [delete]
func (h *OrderHandler) RemoveFromCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	vars := mux.Vars(r)
	id, err := strconv.ParseInt(vars["id"], 10, 64)
	if err != nil {
		RespondWithError(w, http.StatusBadRequest, "Invalid cart item ID", h.logger)
		return
	}

	if err := h.service.RemoveFromCart(r.Context(), claims.UserID, id); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to remove from cart", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Item removed from cart"}, h.logger)
}

// ClearCart godoc
// @Summary Clear shopping cart
// @Description Remove all items from shopping cart
// @Tags Cart
// @Produce json
// @Security BearerAuth
// @Success 200 {object} dto.SuccessResponse
// @Failure 401 {object} dto.ErrorResponse
// @Failure 500 {object} dto.ErrorResponse
// @Router /cart/clear [delete]
func (h *OrderHandler) ClearCart(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value("claims").(*service.Claims)

	if err := h.service.ClearCart(r.Context(), claims.UserID); err != nil {
		RespondWithError(w, http.StatusInternalServerError, "Failed to clear cart", h.logger)
		return
	}

	RespondWithJSON(w, http.StatusOK, dto.SuccessResponse{Message: "Cart cleared successfully"}, h.logger)
}

func (h *OrderHandler) orderToResponse(order *domain.Order) *dto.OrderResponse {
	response := &dto.OrderResponse{
		ID:            order.ID,
		UserID:        order.UserID,
		Status:        string(order.Status),
		TotalAmount:   order.TotalAmount,
		PaymentStatus: string(order.PaymentStatus),
		PaymentID:     order.PaymentID,
		CreatedAt:     order.CreatedAt,
	}

	for _, item := range order.Items {
		response.Items = append(response.Items, dto.OrderItemResponse{
			ID:              item.ID,
			ProductID:       item.ProductID,
			Quantity:        item.Quantity,
			PriceAtPurchase: item.PriceAtPurchase,
		})
	}

	return response
}
