package main

import (
	"flag"
	"fmt"
	"os"
)

func main() {
	// Parse optional flags
	verbose := flag.Bool("verbose", false, "Enable verbose logging")
	flag.Parse()

	// Create logger
	logger := NewLogger(*verbose)

	// Load configuration (hardcoded no config.go)
	config := GetDefaultConfig()
	config.Verbose = *verbose

	// Print header
	fmt.Println()
	logger.Section("LEP BACKEND TEST SUITE")
	logger.Info("Backend: %s", config.BackendURL)
	logger.Info("Test User: %s", config.TestUser.Email)
	fmt.Println()

	// Create API client
	client := NewAPIClient(config.BackendURL, logger)

	// Create test suite
	suite := NewTestSuite(client, logger, config)

	// Run all tests
	suite.RunAll()

	// Print final summary
	fmt.Println()
	logger.Info("═══════════════════════════════════════════════════════════════")
	logger.Info("Test execution completed!")
	logger.Info("═══════════════════════════════════════════════════════════════")
	fmt.Println()

	// Exit with appropriate code
	if suite.failed > 0 {
		os.Exit(1)
	}
}
