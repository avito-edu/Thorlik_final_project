package main

import (
	"fmt"
	"os"
)

// накидал тестов для проверки эндпоинтов, решил оставить
func main() {
	fmt.Println()

	if len(os.Args) < 2 {
		return
	}

	switch os.Args[1] {
	case "auth":
		RunAuthTests()
	case "users":
		RunUserTests()
	case "products":
		RunProductTests()
	case "cart":
		RunCartTests()
	case "orders":
		RunOrderTests()
	}
}
