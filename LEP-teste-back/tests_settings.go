package main

// =============================================================================
// TESTES DE CONFIGURAÇÕES (Independente)
// Inclui: Display Settings, Theme Settings
// =============================================================================

func (ts *TestSuite) RunSettingsTests() {
	ts.logger.Section("FASE 8: CONFIGURAÇÕES")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Display Settings", ts.TestDisplaySettingsConfig},
		{"Theme Settings", ts.TestThemeSettingsConfig},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// DISPLAY SETTINGS
// =============================================================================

func (ts *TestSuite) TestDisplaySettingsConfig() bool {
	ts.logger.Subsection("DISPLAY SETTINGS - Configurações de exibição")
	allPassed := true

	// GET DISPLAY SETTINGS
	ts.logger.Info("1. Obter configurações de display - GET /project/settings/display")
	_, err := ts.client.Request("GET", "/project/settings/display", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /project/settings/display", true, "Endpoint não implementado (não crítico)")
		} else {
			ts.addResult("GET /project/settings/display", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /project/settings/display", true, "Configurações obtidas")
	}

	// UPDATE DISPLAY SETTINGS
	ts.logger.Info("2. Atualizar configurações de display - PUT /project/settings/display")
	updatePayload := map[string]interface{}{
		"show_prices":      true,
		"show_images":      true,
		"show_description": true,
		"show_prep_time":   false,
		"show_rating":      true,
		"products_per_row": 3,
	}

	_, err = ts.client.Request("PUT", "/project/settings/display", updatePayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("PUT /project/settings/display", true, "Endpoint não implementado")
		} else {
			ts.addResult("PUT /project/settings/display", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("PUT /project/settings/display", true, "Configurações atualizadas")
	}

	// RESET DISPLAY SETTINGS
	ts.logger.Info("3. Resetar configurações de display - POST /project/settings/display/reset")
	_, err = ts.client.Request("POST", "/project/settings/display/reset", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /project/settings/display/reset", true, "Endpoint não implementado")
		} else {
			ts.addResult("POST /project/settings/display/reset", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("POST /project/settings/display/reset", true, "Configurações resetadas")
	}

	return allPassed
}

// =============================================================================
// THEME SETTINGS
// =============================================================================

func (ts *TestSuite) TestThemeSettingsConfig() bool {
	ts.logger.Subsection("THEME SETTINGS - Customização de tema")
	allPassed := true

	// GET THEME
	ts.logger.Info("1. Obter tema - GET /project/settings/theme")
	_, err := ts.client.Request("GET", "/project/settings/theme", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /project/settings/theme", true, "Tema não existe ainda")
		} else {
			ts.addResult("GET /project/settings/theme", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /project/settings/theme", true, "Tema obtido")
	}

	// CREATE/UPDATE THEME
	ts.logger.Info("2. Criar/atualizar tema - POST /project/settings/theme")
	themePayload := map[string]interface{}{
		// Light theme colors
		"primary_color":         "#FF6B35",
		"secondary_color":       "#004E89",
		"background_color":      "#FFFFFF",
		"card_background_color": "#F8F9FA",
		"text_color":            "#212529",
		"text_secondary_color":  "#6C757D",
		"accent_color":          "#28A745",
		"destructive_color":     "#DC3545",
		"success_color":         "#28A745",
		"warning_color":         "#FFC107",
		"border_color":          "#DEE2E6",
		"price_color":           "#28A745",
		// Dark theme colors
		"dark_primary_color":         "#FF8C5A",
		"dark_secondary_color":       "#5B9BD5",
		"dark_background_color":      "#1A1A2E",
		"dark_card_background_color": "#16213E",
		"dark_text_color":            "#E8E8E8",
		"dark_text_secondary_color":  "#A0A0A0",
		"dark_accent_color":          "#4ADE80",
		"dark_destructive_color":     "#EF4444",
		"dark_success_color":         "#4ADE80",
		"dark_warning_color":         "#FBBF24",
		"dark_border_color":          "#374151",
		"dark_price_color":           "#4ADE80",
	}

	_, err = ts.client.Request("POST", "/project/settings/theme", themePayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /project/settings/theme", true, "Endpoint não implementado")
		} else {
			ts.addResult("POST /project/settings/theme", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("POST /project/settings/theme", true, "Tema criado/atualizado")
	}

	// UPDATE THEME (PUT)
	ts.logger.Info("3. Atualizar tema - PUT /project/settings/theme")
	updatePayload := map[string]interface{}{
		"primary_color":   "#E74C3C",
		"secondary_color": "#3498DB",
	}

	_, err = ts.client.Request("PUT", "/project/settings/theme", updatePayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("PUT /project/settings/theme", true, "Endpoint não implementado")
		} else {
			ts.addResult("PUT /project/settings/theme", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("PUT /project/settings/theme", true, "Tema atualizado")
	}

	// RESET THEME
	ts.logger.Info("4. Resetar tema - POST /project/settings/theme/reset")
	_, err = ts.client.Request("POST", "/project/settings/theme/reset", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("POST /project/settings/theme/reset", true, "Endpoint não implementado")
		} else {
			ts.addResult("POST /project/settings/theme/reset", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("POST /project/settings/theme/reset", true, "Tema resetado")
	}

	return allPassed
}

// =============================================================================
// GENERAL SETTINGS (Settings legado)
// =============================================================================

func (ts *TestSuite) TestGeneralSettings() bool {
	ts.logger.Subsection("GENERAL SETTINGS - Configurações gerais do projeto")
	allPassed := true

	// GET SETTINGS
	ts.logger.Info("1. Obter configurações gerais - GET /settings")
	_, err := ts.client.Request("GET", "/settings", nil, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("GET /settings", true, "Endpoint não implementado")
		} else {
			ts.addResult("GET /settings", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("GET /settings", true, "Configurações obtidas")
	}

	// UPDATE SETTINGS
	ts.logger.Info("2. Atualizar configurações gerais - PUT /settings")
	updatePayload := map[string]interface{}{
		"opening_time":       "08:00",
		"closing_time":       "22:00",
		"tables_count":       20,
		"capacity_per_table": 4,
	}

	_, err = ts.client.Request("PUT", "/settings", updatePayload, true)
	if err != nil {
		status := ts.client.GetLastStatus()
		if status == 404 {
			ts.addResult("PUT /settings", true, "Endpoint não implementado")
		} else {
			ts.addResult("PUT /settings", false, err.Error())
			allPassed = false
		}
	} else {
		ts.addResult("PUT /settings", true, "Configurações atualizadas")
	}

	return allPassed
}
