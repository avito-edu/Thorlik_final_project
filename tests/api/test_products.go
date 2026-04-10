package main

import (
	"encoding/json"
	"fmt"
)

func RunProductTests() {
	fmt.Println()

	client := NewClient()

	//sellerToken := setupSeller(client)
	//if sellerToken == "" {
	//	return
	//}
	//client.SetToken(sellerToken)

	//productIDs := testCreateProducts(client)

	//testGetAllProducts()

	//if len(productIDs) > 0 {
	//	testGetProductByID(productIDs[0])
	//}

	//testGetMyProducts(client)

	//if len(productIDs) > 0 {
	//	testUpdateProduct(client, productIDs[0])
	//}

	//testCreateInvalidProduct(client)

	//для модера
	modToken := setupModerator(client)
	if modToken != "" {
		modClient := NewClient()
		modClient.SetToken(modToken)

		testGetPendingProducts(modClient)

		testModerateProduct(modClient, 1, "approved")
		testModerateProduct(modClient, 2, "approved")
		testGetAllProducts()
	}

	//if len(productIDs) > 2 {
	//	testDeleteProduct(client, productIDs[2])
	//}
}

func setupSeller(client *Client) string {
	registerReq := RegisterRequest{
		Email:     "seller@example.com",
		Password:  "seller123",
		FirstName: "Seller",
		LastName:  "User",
	}
	client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "seller@example.com",
		Password: "seller123",
	}
	body, status, err := client.Post("/auth/login", loginReq)
	if err != nil || status != 200 {
		return ""
	}

	var resp LoginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}

	return resp.Token
}

func setupModerator(client *Client) string {
	fmt.Println(" поменять вручную роль пользователя на модера")

	//registerReq := RegisterRequest{
	//	Email:     "moderator@example.com",
	//	Password:  "mod123",
	//	FirstName: "Moderator",
	//	LastName:  "User",
	//}
	//client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "moderator@example.com",
		Password: "mod123",
	}
	body, status, err := client.Post("/auth/login", loginReq)
	if err != nil || status != 200 {
		return ""
	}

	var resp LoginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}

	if resp.User.Role != "moderator" && resp.User.Role != "admin" {
		fmt.Printf("роль не поменяна")
	}

	return resp.Token
}

func testCreateProducts(client *Client) []int64 {
	products := []CreateProductRequest{
		{
			Title:       "iPhone 15 Pro",
			Description: "Latest Apple smartphone with A17 Pro chip",
			Price:       999.99,
			Quantity:    20,
			Category:    "Electronics",
		},
		{
			Title:       "AirPods Pro 2",
			Description: "Wireless earbuds with Active Noise Cancellation",
			Price:       249.99,
			Quantity:    50,
			Category:    "Electronics",
		},
		{
			Title:       "MacBook Charger 96W",
			Description: "USB-C Power Adapter for MacBook Pro",
			Price:       79.99,
			Quantity:    100,
			Category:    "Accessories",
		},
	}

	var createdIDs []int64

	for i, product := range products {
		body, status, err := client.Post("/products", product)
		Print(fmt.Sprintf("Create Product %d", i+1), body, status, err)

		if status == 201 || status == 200 {
			var created struct {
				ID int64 `json:"id"`
			}
			if err := json.Unmarshal(body, &created); err == nil {
				createdIDs = append(createdIDs, created.ID)
				fmt.Printf("ОК\n")
			}
		}
	}

	return createdIDs
}

func testGetAllProducts() {
	client := NewClient()
	body, status, err := client.Get("/products")
	Print("Get All Products", body, status, err)
}

func testGetProductByID(productID int64) {
	client := NewClient()
	path := fmt.Sprintf("/products/%d", productID)
	body, status, err := client.Get(path)
	Print("Get Product by ID", body, status, err)
}

func testGetMyProducts(client *Client) {
	body, status, err := client.Get("/products/my")
	Print("Get My Products", body, status, err)
}

func testUpdateProduct(client *Client, productID int64) {
	req := UpdateProductRequest{
		Title:    "iPhone 15 Pro (Updated)",
		Price:    899.99,
		Quantity: 15,
	}

	path := fmt.Sprintf("/products/%d", productID)
	body, status, err := client.Put(path, req)
	Print("Update Product", body, status, err)
}

func testCreateInvalidProduct(client *Client) {
	req := CreateProductRequest{
		Title:    "",
		Price:    -10,
		Quantity: -5,
	}

	body, status, err := client.Post("/products", req)
	Print("Create Invalid Product", body, status, err)

	if status == 400 {
		fmt.Println("ОК")
	}
}

func testGetPendingProducts(client *Client) {
	body, status, err := client.Get("/moderate/products/pending")
	Print("Get Pending Products", body, status, err)
}

func testModerateProduct(client *Client, productID int64, status string) {
	req := ModerateRequest{Status: status}
	path := fmt.Sprintf("/moderate/products/%d", productID)
	body, respStatus, err := client.Post(path, req)
	Print(fmt.Sprintf("Moderate Product %d", productID), body, respStatus, err)
}

func testDeleteProduct(client *Client, productID int64) {
	path := fmt.Sprintf("/products/%d", productID)
	body, status, err := client.Delete(path)
	Print("Delete Product", body, status, err)
}
