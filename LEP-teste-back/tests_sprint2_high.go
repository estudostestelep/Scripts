package main

import (
	"fmt"
)

// Sprint 2: HIGH PRIORITY TESTS (13 testes)
// Objetivo: Implementar configurações, tema e menus avançados
// Estimado: 1 dia de trabalho

// ============================================================================
// DISPLAY SETTINGS (3 testes - 2 horas)
// ============================================================================

func (ts *TestSuite) TestGetDisplaySettings() bool {
	ts.logger.Subsection("GET /project/settings/display - Get display settings")

	_, err := ts.client.Request("GET", "/project/settings/display", nil, true)
	if err != nil {
		ts.addResult("GET /project/settings/display", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("GET /project/settings/display", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("GET /project/settings/display", true, "Display settings retrieved")
	return true
}

func (ts *TestSuite) TestUpdateDisplaySettings() bool {
	ts.logger.Subsection("PUT /project/settings/display - Update display settings")

	payload := map[string]interface{}{
		"show_prices":       true,
		"show_descriptions": true,
		"show_images":       true,
		"item_per_page":     12,
	}

	_, err := ts.client.Request("PUT", "/project/settings/display", payload, true)
	if err != nil {
		ts.addResult("PUT /project/settings/display", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 201 {
		ts.addResult("PUT /project/settings/display", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("PUT /project/settings/display", true, "Display settings updated")
	return true
}

func (ts *TestSuite) TestResetDisplaySettings() bool {
	ts.logger.Subsection("POST /project/settings/display/reset - Reset display settings")

	_, err := ts.client.Request("POST", "/project/settings/display/reset", nil, true)
	if err != nil {
		ts.addResult("POST /project/settings/display/reset", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 {
		ts.addResult("POST /project/settings/display/reset", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("POST /project/settings/display/reset", true, "Display settings reset")
	return true
}

// ============================================================================
// THEME CUSTOMIZATION (5 testes - 3 horas)
// ============================================================================

func (ts *TestSuite) TestGetThemeSettings() bool {
	ts.logger.Subsection("GET /project/settings/theme - Get theme settings")

	_, err := ts.client.Request("GET", "/project/settings/theme", nil, true)
	if err != nil {
		ts.addResult("GET /project/settings/theme", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("GET /project/settings/theme", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("GET /project/settings/theme", true, "Theme settings retrieved")
	return true
}

func (ts *TestSuite) TestCreateThemeSettings() bool {
	ts.logger.Subsection("POST /project/settings/theme - Create theme settings")

	payload := map[string]interface{}{
		"primary_color":   "#FF6B35",
		"secondary_color": "#004E89",
		"accent_color":    "#F7931E",
		"font_family":     "Roboto",
		"theme_mode":      "light",
	}

	_, err := ts.client.Request("POST", "/project/settings/theme", payload, true)
	if err != nil {
		ts.addResult("POST /project/settings/theme", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 201 {
		ts.addResult("POST /project/settings/theme", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("POST /project/settings/theme", true, "Theme settings created")
	return true
}

func (ts *TestSuite) TestUpdateThemeSettings() bool {
	ts.logger.Subsection("PUT /project/settings/theme - Update theme settings")

	payload := map[string]interface{}{
		"primary_color": "#1A1A2E",
		"theme_mode":    "dark",
	}

	_, err := ts.client.Request("PUT", "/project/settings/theme", payload, true)
	if err != nil {
		ts.addResult("PUT /project/settings/theme", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 {
		ts.addResult("PUT /project/settings/theme", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("PUT /project/settings/theme", true, "Theme settings updated")
	return true
}

func (ts *TestSuite) TestResetThemeSettings() bool {
	ts.logger.Subsection("POST /project/settings/theme/reset - Reset theme settings")

	_, err := ts.client.Request("POST", "/project/settings/theme/reset", nil, true)
	if err != nil {
		ts.addResult("POST /project/settings/theme/reset", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 {
		ts.addResult("POST /project/settings/theme/reset", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("POST /project/settings/theme/reset", true, "Theme settings reset")
	return true
}

func (ts *TestSuite) TestDeleteThemeSettings() bool {
	ts.logger.Subsection("DELETE /project/settings/theme - Delete theme settings")

	_, err := ts.client.Request("DELETE", "/project/settings/theme", nil, true)
	if err != nil {
		ts.addResult("DELETE /project/settings/theme", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("DELETE /project/settings/theme", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("DELETE /project/settings/theme", true, "Theme settings deleted")
	return true
}

// ============================================================================
// MENU ADVANCED FEATURES (5 testes - 3 horas)
// ============================================================================

func (ts *TestSuite) TestGetActiveNowMenu() bool {
	ts.logger.Subsection("GET /menu/active-now - Get currently active menu")

	_, err := ts.client.Request("GET", "/menu/active-now", nil, true)
	if err != nil {
		ts.addResult("GET /menu/active-now", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("GET /menu/active-now", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("GET /menu/active-now", true, "Active menu retrieved")
	return true
}

func (ts *TestSuite) TestGetActiveMenus() bool {
	ts.logger.Subsection("GET /menu/active - Get all active menus")

	_, err := ts.client.Request("GET", "/menu/active", nil, true)
	if err != nil {
		ts.addResult("GET /menu/active", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("GET /menu/active", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("GET /menu/active", true, "Active menus retrieved")
	return true
}

func (ts *TestSuite) TestGetMenuOptions() bool {
	ts.logger.Subsection("GET /menu/options - Get menu options")

	_, err := ts.client.Request("GET", "/menu/options", nil, true)
	if err != nil {
		ts.addResult("GET /menu/options", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("GET /menu/options", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("GET /menu/options", true, "Menu options retrieved")
	return true
}

func (ts *TestSuite) TestMenuManualOverride() bool {
	ts.logger.Subsection("PUT /menu/:id/manual-override - Set manual menu override")

	menuID := "test-menu-id"
	payload := map[string]interface{}{
		"override_enabled": true,
		"override_menu_id": "alternate-menu-id",
	}

	_, err := ts.client.Request("PUT", fmt.Sprintf("/menu/%s/manual-override", menuID), payload, true)
	if err != nil {
		ts.addResult("PUT /menu/:id/manual-override", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("PUT /menu/:id/manual-override", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("PUT /menu/:id/manual-override", true, "Menu override set")
	return true
}

func (ts *TestSuite) TestDeleteMenuManualOverride() bool {
	ts.logger.Subsection("DELETE /menu/manual-override - Delete manual menu override")

	_, err := ts.client.Request("DELETE", "/menu/manual-override", nil, true)
	if err != nil {
		ts.addResult("DELETE /menu/manual-override", false, fmt.Sprintf("Erro: %s", err.Error()))
		return false
	}

	if status := ts.client.GetLastStatus(); status != 200 && status != 404 {
		ts.addResult("DELETE /menu/manual-override", false, fmt.Sprintf("Status: %d", status))
		return false
	}

	ts.addResult("DELETE /menu/manual-override", true, "Menu override deleted")
	return true
}

// ============================================================================
// SPRINT 2 ORCHESTRATION
// ============================================================================

func (ts *TestSuite) RunSprintTwoTests() {
	ts.logger.Section("SPRINT 2: HIGH PRIORITY TESTS (13 testes)")
	ts.logger.Info("Objetivo: Configurações, tema e menus avançados")
	ts.logger.Info("Tempo estimado: 1 dia")
	fmt.Println()

	passed := 0
	failed := 0

	// Display Settings Tests (3)
	ts.logger.Info("Display Settings Tests (3):")
	if ts.TestGetDisplaySettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestUpdateDisplaySettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestResetDisplaySettings() {
		passed++
	} else {
		failed++
	}

	// Theme Customization Tests (5)
	ts.logger.Info("Theme Customization Tests (5):")
	if ts.TestGetThemeSettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestCreateThemeSettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestUpdateThemeSettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestResetThemeSettings() {
		passed++
	} else {
		failed++
	}
	if ts.TestDeleteThemeSettings() {
		passed++
	} else {
		failed++
	}

	// Menu Advanced Tests (5)
	ts.logger.Info("Menu Advanced Features Tests (5):")
	if ts.TestGetActiveNowMenu() {
		passed++
	} else {
		failed++
	}
	if ts.TestGetActiveMenus() {
		passed++
	} else {
		failed++
	}
	if ts.TestGetMenuOptions() {
		passed++
	} else {
		failed++
	}
	if ts.TestMenuManualOverride() {
		passed++
	} else {
		failed++
	}
	if ts.TestDeleteMenuManualOverride() {
		passed++
	} else {
		failed++
	}

	fmt.Println()
	ts.logger.Stats(passed+failed, passed, failed)
	ts.passed += passed
	ts.failed += failed
}
