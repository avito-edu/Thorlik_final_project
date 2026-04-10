package main

import (
	"encoding/json"
	"fmt"
)

func RunAuthTests() {
	fmt.Println()

	client := NewClient()

	//регистрация
	testRegister(client)

	//регистрация дубликата
	testRegisterDuplicate(client)

	//логин
	testLogin(client)

	//логин с неправильным паролем
	testLoginWrongPassword(client)
}

func testRegister(client *Client) {
	fmt.Println("\n тест регистрации")

	req := RegisterRequest{
		Email:     "testuser@example.com",
		Password:  "password123",
		FirstName: "Iva",
		LastName:  "Horl",
	}

	body, status, err := client.Post("/auth/register", req)
	Print("User", body, status, err)
}

func testRegisterDuplicate(client *Client) {
	fmt.Println("\n регистрация дубликата")

	req := RegisterRequest{
		Email:     "testuser@example.com",
		Password:  "password123",
		FirstName: "Iva",
		LastName:  "Horl",
	}

	body, status, err := client.Post("/auth/register", req)
	Print("Регистрация дубликата", body, status, err)

	if status == 409 {
		fmt.Println("ОК")
	}
}

func testLogin(client *Client) {
	fmt.Println("\nлогин")

	req := LoginRequest{
		Email:    "testuser@example.com",
		Password: "password123",
	}

	body, status, err := client.Post("/auth/login", req)
	Print("Логин", body, status, err)

	if status == 200 {
		var resp LoginResponse
		if err := json.Unmarshal(body, &resp); err == nil {
			fmt.Printf("Токен: %s...%s\n", resp.Token[:20], resp.Token[len(resp.Token)-10:])
			fmt.Printf("Юзер: %s %s (%s)\n", resp.User.FirstName, resp.User.LastName, resp.User.Role)
		}
	}
}

func testLoginWrongPassword(client *Client) {
	fmt.Println("\nлогин с неправильным паролем")

	req := LoginRequest{
		Email:    "testuser@example.com",
		Password: "wrongpassword",
	}

	body, status, err := client.Post("/auth/login", req)
	Print("Логин с неправильным паролем", body, status, err)

	if status == 401 {
		fmt.Println("ОК")
	}
}
