package main

import (
	"encoding/json"
	"fmt"
)

func RunOrderTests() {
	fmt.Println()

	client := NewClient()

	productIDs := setupApprovedProducts(client)

	if len(productIDs) == 0 {
		return
	}

	buyerToken := setupOrderBuyer(client)
	if buyerToken == "" {
		return
	}
	client.SetToken(buyerToken)

	testGetMyOrders(client)

	orderID := testCreateOrderDirect(client, productIDs)

	if orderID > 0 {
		testGetOrderByID(client, orderID)
	}

	testAddItemsForOrder(client, productIDs)
	orderFromCartID := testCreateOrderFromCart(client)

	testGetMyOrders(client)

	if orderID > 0 {
		testProcessPayment(client, orderID)
	}

	if orderID > 0 {
		testGetOrderByID(client, orderID)
	}

	//для админа
	adminToken := setupOrderAdmin(client)
	if adminToken != "" {
		adminClient := NewClient()
		adminClient.SetToken(adminToken)

		testGetAllOrders(adminClient)

		if orderID > 0 {
			testUpdateOrderStatus(adminClient, orderID, "shipped")
			testUpdateOrderStatus(adminClient, orderID, "delivered")
		}

		if orderFromCartID > 0 {
			testUpdateOrderStatus(adminClient, orderFromCartID, "cancelled")
		}
	}
}

func setupApprovedProducts(client *Client) []int64 {
	registerReq := RegisterRequest{
		Email:     "order_seller@example.com",
		Password:  "seller123",
		FirstName: "Order",
		LastName:  "Seller",
	}
	client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "order_seller@example.com",
		Password: "seller123",
	}
	body, status, _ := client.Post("/auth/login", loginReq)
	if status != 200 {
		return nil
	}

	var loginResp LoginResponse
	json.Unmarshal(body, &loginResp)
	client.SetToken(loginResp.Token)

	products := []CreateProductRequest{
		{Title: "Order Test Product 1", Description: "For order testing", Price: 150.00, Quantity: 100, Category: "Test"},
		{Title: "Order Test Product 2", Description: "For order testing", Price: 75.50, Quantity: 50, Category: "Test"},
	}

	var productIDs []int64
	for _, p := range products {
		body, status, _ := client.Post("/products", p)
		if status == 200 || status == 201 {
			var created struct {
				ID int64 `json:"id"`
			}
			json.Unmarshal(body, &created)
			productIDs = append(productIDs, created.ID)
		}
	}

	fmt.Println("поменять вручную статус продуктов на approved в БД")

	return productIDs
}

func setupOrderBuyer(client *Client) string {
	registerReq := RegisterRequest{
		Email:     "order_buyer@example.com",
		Password:  "buyer123",
		FirstName: "Order",
		LastName:  "Buyer",
	}
	client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "order_buyer@example.com",
		Password: "buyer123",
	}
	body, status, _ := client.Post("/auth/login", loginReq)
	if status != 200 {
		return ""
	}

	var resp LoginResponse
	json.Unmarshal(body, &resp)
	return resp.Token
}

func setupOrderAdmin(client *Client) string {
	fmt.Println("поменять вручную роль пользователя")

	// registerReq := RegisterRequest{
	// 	Email:     "order_admin@example.com",
	// 	Password:  "admin123",
	// 	FirstName: "Order",
	// 	LastName:  "Admin",
	// }
	// client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "order_admin@example.com",
		Password: "admin123",
	}
	body, status, _ := client.Post("/auth/login", loginReq)
	if status != 200 {
		return ""
	}

	var resp LoginResponse
	json.Unmarshal(body, &resp)
	return resp.Token
}

func testGetMyOrders(client *Client) {
	body, status, err := client.Get("/orders")
	Print("Get My Orders", body, status, err)
}

func testCreateOrderDirect(client *Client, productIDs []int64) int64 {
	items := make([]OrderItemRequest, 0)
	for i, id := range productIDs {
		items = append(items, OrderItemRequest{
			ProductID: id,
			Quantity:  i + 1,
		})
	}

	req := CreateOrderRequest{Items: items}
	body, status, err := client.Post("/orders", req)
	Print("Create Order", body, status, err)

	if status == 200 || status == 201 {
		var created struct {
			ID int64 `json:"id"`
		}
		if json.Unmarshal(body, &created) == nil {
			return created.ID
		}
	}

	return 0
}

func testAddItemsForOrder(client *Client, productIDs []int64) {
	for _, id := range productIDs {
		req := CartItemRequest{ProductID: id, Quantity: 2}
		client.Post("/cart", req)
	}
}

func testCreateOrderFromCart(client *Client) int64 {
	body, status, err := client.Post("/orders/from-cart", nil)
	Print("Create Order From Cart", body, status, err)

	if status == 200 || status == 201 {
		var created struct {
			ID int64 `json:"id"`
		}
		if json.Unmarshal(body, &created) == nil {
			return created.ID
		}
	}

	return 0
}

func testGetOrderByID(client *Client, orderID int64) {
	path := fmt.Sprintf("/orders/%d", orderID)
	body, status, err := client.Get(path)
	Print("Get Order by ID", body, status, err)
}

func testProcessPayment(client *Client, orderID int64) {
	req := PaymentRequest{
		OrderID:       orderID,
		PaymentMethod: "card",
	}

	body, status, err := client.Post("/payments", req)
	Print("Process Payment", body, status, err)

	if status == 200 {
		fmt.Println("ОК")
	}
}

func testGetAllOrders(client *Client) {
	body, status, err := client.Get("/admin/orders")
	Print("Get All Orders", body, status, err)
}

func testUpdateOrderStatus(client *Client, orderID int64, status string) {
	req := UpdateStatusRequest{Status: status}
	path := fmt.Sprintf("/admin/orders/%d/status", orderID)
	body, respStatus, err := client.Put(path, req)
	Print(fmt.Sprintf("Update Order Status to %s", status), body, respStatus, err)
}
