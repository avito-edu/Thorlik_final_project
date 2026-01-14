package main

import (
	"encoding/json"
	"fmt"
)

func RunCartTests() {
	fmt.Println()

	client := NewClient()

	productIDs := setupProductsForCart(client)
	if len(productIDs) == 0 {
		fmt.Println("Ошибка")
		return
	}

	buyerToken := setupBuyer(client)
	if buyerToken == "" {
		fmt.Println("Ошибка")
		return
	}
	client.SetToken(buyerToken)

	// testGetCart(client)

	// cartItemIDs := testAddToCart(client, productIDs)

	// testGetCart(client)

	// if len(cartItemIDs) > 0 {
	// 	testUpdateCartItem(client, cartItemIDs[0], 5)
	// }

	// testGetCart(client)

	// if len(cartItemIDs) > 1 {
	// 	testRemoveFromCart(client, cartItemIDs[1])
	// }

	testGetCart(client)

	// testAddToCart(client, productIDs)

	testClearCart(client)

	testGetCart(client)
}

func setupProductsForCart(client *Client) []int64 {
	registerReq := RegisterRequest{
		Email:     "cart_seller@example.com",
		Password:  "seller123",
		FirstName: "Cart",
		LastName:  "Seller",
	}
	client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "cart_seller@example.com",
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
		{Title: "Cart Test Product 1", Description: "Test", Price: 100.00, Quantity: 50, Category: "Test"},
		{Title: "Cart Test Product 2", Description: "Test", Price: 200.00, Quantity: 30, Category: "Test"},
		{Title: "Cart Test Product 3", Description: "Test", Price: 50.00, Quantity: 100, Category: "Test"},
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

func setupBuyer(client *Client) string {
	// registerReq := RegisterRequest{
	// 	Email:     "cart_buyer@example.com",
	// 	Password:  "buyer123",
	// 	FirstName: "Cart",
	// 	LastName:  "Buyer",
	// }
	// client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "cart_buyer@example.com",
		Password: "buyer123",
	}
	body, status, err := client.Post("/auth/login", loginReq)
	if err != nil || status != 200 {
		fmt.Printf("Ошибка: %v\n", err)
		return ""
	}

	var resp LoginResponse
	json.Unmarshal(body, &resp)
	return resp.Token
}

func testGetCart(client *Client) {
	body, status, err := client.Get("/cart")
	Print("Get Cart", body, status, err)
}

func testAddToCart(client *Client, productIDs []int64) []int64 {
	var cartItemIDs []int64

	for i, productID := range productIDs {
		req := CartItemRequest{
			ProductID: productID,
			Quantity:  i + 1,
		}

		body, status, err := client.Post("/cart", req)
		Print(fmt.Sprintf("Add Product %d to Cart", productID), body, status, err)

		if status == 200 || status == 201 {
			var created struct {
				ID int64 `json:"id"`
			}
			if json.Unmarshal(body, &created) == nil && created.ID > 0 {
				cartItemIDs = append(cartItemIDs, created.ID)
			}
		}
	}

	return cartItemIDs
}

func testUpdateCartItem(client *Client, cartItemID int64, quantity int) {
	req := UpdateCartRequest{Quantity: quantity}
	path := fmt.Sprintf("/cart/%d", cartItemID)
	body, status, err := client.Put(path, req)
	Print("Update Cart Item", body, status, err)
}

func testRemoveFromCart(client *Client, cartItemID int64) {
	path := fmt.Sprintf("/cart/%d", cartItemID)
	body, status, err := client.Delete(path)
	Print("Remove from Cart", body, status, err)
}

func testClearCart(client *Client) {
	body, status, err := client.Delete("/cart/clear")
	Print("Clear Cart", body, status, err)
}
