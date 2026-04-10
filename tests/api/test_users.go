package main

import (
	"encoding/json"
	"fmt"
)

func RunUserTests() {
	fmt.Println()

	client := NewClient()

	//token := setupUser(client)
	//if token == "" {
	//	return
	//}
	//client.SetToken(token)

	//testGetProfile(client)

	//testUpdateProfile(client)

	//testGetProfile(client)

	//testChangePassword(client)

	//testUnauthorizedAccess()

	//для админа
	adminToken := setupAdmin(client)
	if adminToken != "" {
		adminClient := NewClient()
		adminClient.SetToken(adminToken)

		//testGetAllUsers(adminClient)

		//testGetUserByID(adminClient, 1)

		//testUpdateUserRole(adminClient, 2, "moderator")

		testDeleteUser(adminClient, 99999, false)

		deleteUserID := createUserForDeletion(client)
		if deleteUserID > 0 {
			testDeleteUser(adminClient, deleteUserID, true)
			testGetUserByID(adminClient, deleteUserID)
		}
	}
}

func setupUser(client *Client) string {
	//registerReq := RegisterRequest{
	//	Email:     "user_test@example.com",
	//	Password:  "password123",
	//	FirstName: "Test",
	//	LastName:  "User",
	//}
	//client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "user_test@example.com",
		Password: "password123",
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

func setupAdmin(client *Client) string {
	fmt.Println("поменять вручную роль пользователя на админа")

	//registerReq := RegisterRequest{
	//	Email:     "admin_test@example.com",
	//	Password:  "admin123",
	//	FirstName: "Admin",
	//	LastName:  "User",
	//}
	//client.Post("/auth/register", registerReq)

	loginReq := LoginRequest{
		Email:    "admin_test@example.com",
		Password: "admin123",
	}
	body, status, err := client.Post("/auth/login", loginReq)
	if err != nil || status != 200 {
		return ""
	}

	var resp LoginResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return ""
	}

	if resp.User.Role != "admin" {
		fmt.Printf("не та роль\n")
	}

	return resp.Token
}

func testGetProfile(client *Client) {
	body, status, err := client.Get("/users/me")
	Print("Get Profile", body, status, err)
}

func testUpdateProfile(client *Client) {
	req := UpdateProfileRequest{
		FirstName: "Updated",
		LastName:  "Name",
	}

	body, status, err := client.Put("/users/me", req)
	Print("Update Profile", body, status, err)
}

func testChangePassword(client *Client) {
	req := ChangePasswordRequest{
		OldPassword: "password123",
		NewPassword: "newpassword456",
	}

	body, status, err := client.Put("/users/me/password", req)
	Print("Change Password", body, status, err)

	if status == 200 {
		fmt.Println("ОК")

		loginReq := LoginRequest{
			Email:    "user_test@example.com",
			Password: "newpassword456",
		}
		body, status, err = client.Post("/auth/login", loginReq)
		if status == 200 {
			fmt.Println("ОК")
		}
	}
}

func testUnauthorizedAccess() {
	client := NewClient()
	body, status, err := client.Get("/users/me")
	Print("Unauthorized Access", body, status, err)

	if status == 401 {
		fmt.Println("ОК")
	}
}

func testGetAllUsers(client *Client) {
	body, status, err := client.Get("/users")
	Print("Get All Users", body, status, err)
}

func testGetUserByID(client *Client, userID int64) {
	path := fmt.Sprintf("/users/%d", userID)
	body, status, err := client.Get(path)
	Print("Get User by ID", body, status, err)
}

func testUpdateUserRole(client *Client, userID int64, role string) {
	req := UpdateRoleRequest{Role: role}
	path := fmt.Sprintf("/users/%d/role", userID)
	body, status, err := client.Put(path, req)
	Print("Update User Role", body, status, err)
}

func createUserForDeletion(client *Client) int64 {
	registerReq := RegisterRequest{
		Email:     "user_to_delete@example.com",
		Password:  "delete123",
		FirstName: "Delete",
		LastName:  "Me",
	}

	body, status, _ := client.Post("/auth/register", registerReq)
	if status == 201 || status == 200 {
		var created struct {
			ID int64 `json:"id"`
		}
		if json.Unmarshal(body, &created) == nil {
			fmt.Printf("ОК\n")
			return created.ID
		}
	} else if status == 409 {
		loginReq := LoginRequest{
			Email:    "user_to_delete@example.com",
			Password: "delete123",
		}
		body, status, _ := client.Post("/auth/login", loginReq)
		if status == 200 {
			var resp LoginResponse
			if json.Unmarshal(body, &resp) == nil {
				fmt.Printf("ОК")
				return resp.User.ID
			}
		}
	}

	return 0
}

func testDeleteUser(client *Client, userID int64, expectSuccess bool) {
	path := fmt.Sprintf("/users/%d", userID)
	body, status, err := client.Delete(path)
	Print(fmt.Sprintf("Delete User %d", userID), body, status, err)

	if expectSuccess {
		if status == 200 || status == 204 {
			fmt.Println("ОК")
		} else {
			fmt.Printf("Код ошибки %d\n", status)
		}
	} else {
		if status == 404 {
			fmt.Println("ОК")
		} else {
			fmt.Printf("Код ошибки %d\n", status)
		}
	}
}
