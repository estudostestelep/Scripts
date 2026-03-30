package main

import (
	"fmt"
	"log"
)

func QuickImageUploadTest() {
	log.Println("🧪 Starting quick image upload test...")

	// Criar config
	config := GetDefaultConfig()
	if config.BackendURL == "" {
		config.BackendURL = "http://localhost:8080"
	}

	// Criar client
	logger := NewLogger(false)
	client := NewAPIClient(config.BackendURL, logger)

	// Login
	fmt.Println("[1] Logging in...")
	loginResp, err := client.Request("POST", "/login", map[string]interface{}{
		"email":    config.TestUser.Email,
		"password": config.TestUser.Password,
	}, false)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}

	// Extract token
	data := loginResp["data"].(map[string]interface{})
	client.token = data["token"].(string)
	client.orgID = data["organization_id"].(string)
	client.projID = data["project_id"].(string)
	fmt.Println("✓ Login successful")

	// Create test image
	fmt.Println("[2] Creating test PNG...")
	pngData := generateTestPNG()
	fmt.Printf("✓ PNG created (%d bytes)\n", len(pngData))

	// Upload
	fmt.Println("[3] Uploading image...")
	uploadResp, err := client.RequestWithFile("POST", "/upload/categories/image", pngData, "test.png", "image/png", true)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	// Extract result
	success := uploadResp["success"].(bool)
	uploadData := uploadResp["data"].(map[string]interface{})
	imageURL := uploadData["image_url"].(string)

	fmt.Printf("✓ Upload successful: %v\n", success)
	fmt.Printf("✓ Image URL: %s\n", imageURL)

	// Validate URL
	if len(imageURL) > 0 {
		if len(imageURL) > 150 {
			imageURL = imageURL[:150] + "..."
		}
		fmt.Println("\n✅ TEST PASSED - Image URL is valid!")
	} else {
		log.Fatal("❌ TEST FAILED - No image URL returned")
	}
}
