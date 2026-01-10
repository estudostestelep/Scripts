package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

func TestThemeResetQuick() {
	separator := strings.Repeat("=", 70)
	fmt.Println("\n" + separator)
	fmt.Println("THEME RESET TEST - Quick Validation")
	fmt.Println(separator + "\n")

	// Create config
	config := GetDefaultConfig()
	config.BackendURL = "http://localhost:8080"

	// Create client
	logger := NewLogger(true)
	client := NewAPIClient(config.BackendURL, logger)

	// Step 1: Login
	fmt.Println("[1] Logging in as " + config.TestUser.Email + "...")
	loginResp, err := client.Request("POST", "/login", map[string]interface{}{
		"email":    config.TestUser.Email,
		"password": config.TestUser.Password,
	}, false)
	if err != nil {
		log.Fatalf("❌ Login failed: %v", err)
	}

	// Extract token and headers
	data := loginResp["data"].(map[string]interface{})
	client.token = data["token"].(string)
	client.orgID = data["organization_id"].(string)
	client.projID = data["project_id"].(string)
	fmt.Printf("✅ Login successful (Token: %s...)\n\n", client.token[:20])

	// Step 2: Get current theme
	fmt.Println("[2] Getting current theme configuration...")
	getResp, err := client.Request("GET", "/project/settings/theme", nil, true)
	if err != nil {
		log.Fatalf("❌ GET theme failed: %v", err)
	}
	fmt.Println("✅ Current theme retrieved")
	prettyPrint("Before Reset", getResp)

	// Step 3: Reset theme
	fmt.Println("\n[3] Resetting theme to defaults...")
	resetResp, err := client.Request("POST", "/project/settings/theme/reset", map[string]interface{}{}, true)
	if err != nil {
		log.Fatalf("❌ Reset theme failed: %v", err)
	}
	fmt.Println("✅ Theme reset successful")
	prettyPrint("After Reset", resetResp)

	// Step 4: Verify reset values
	fmt.Println("\n[4] Verifying reset values...")
	verified := true

	// Extract data
	dataObj, ok := resetResp["data"].(map[string]interface{})
	if !ok {
		log.Fatal("❌ Invalid response format - no 'data' field")
	}

	// Check expected defaults
	expectedDefaults := map[string]string{
		"primary_color_light":   "#1E293B",
		"primary_color_dark":    "#F8FAFC",
		"background_color_light": "#FFFFFF",
		"background_color_dark":  "#0F172A",
	}

	for field, expectedValue := range expectedDefaults {
		actualValue := ""
		if val, exists := dataObj[field]; exists && val != nil {
			actualValue = val.(string)
		}

		if actualValue == expectedValue {
			fmt.Printf("  ✅ %s: %s (correct)\n", field, actualValue)
		} else {
			fmt.Printf("  ❌ %s: got '%s', expected '%s'\n", field, actualValue, expectedValue)
			verified = false
		}
	}

	// Final result
	fmt.Println("\n" + strings.Repeat("=", 70))
	if verified {
		fmt.Println("✅ THEME RESET TEST PASSED - All colors reset to defaults!")
	} else {
		fmt.Println("❌ THEME RESET TEST FAILED - Some colors don't match defaults")
	}
	fmt.Println(strings.Repeat("=", 70) + "\n")
}

func prettyPrint(title string, data map[string]interface{}) {
	fmt.Printf("\n--- %s ---\n", title)
	jsonData, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(jsonData))
}
