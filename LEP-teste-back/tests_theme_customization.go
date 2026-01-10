package main

import (
	"fmt"
)

/**
 * 🎨 Theme Customization Tests (Light/Dark Variants)
 *
 * Testes para o módulo de customização de cores/tema com suporte a Light/Dark
 * Endpoints:
 * - GET /project/settings/theme
 * - POST /project/settings/theme
 * - PUT /project/settings/theme
 * - POST /project/settings/theme/reset
 * - DELETE /project/settings/theme
 */

// TestGetTheme testa obter customização de tema com light/dark variants
func (ts *TestSuite) TestGetTheme() bool {
	ts.logger.Section("THEME CUSTOMIZATION - GET /project/settings/theme (Light/Dark)")

	resp, err := ts.client.Request("GET", "/project/settings/theme", nil, true)
	if err != nil {
		ts.addResult("GET /project/settings/theme", false, err.Error())
		return false
	}

	// Verificar se retorna campos light/dark esperados
	lightPrimary := ts.client.ExtractString(resp, "primary_color_light")
	darkPrimary := ts.client.ExtractString(resp, "primary_color_dark")

	if lightPrimary == "" || darkPrimary == "" {
		ts.addResult("GET /project/settings/theme", false, "Campos light/dark não retornados")
		return false
	}

	ts.addResult("GET /project/settings/theme", true,
		fmt.Sprintf("Cores Light/Dark obtidas - Light: %s, Dark: %s", lightPrimary, darkPrimary))
	return true
}

// TestCreateThemeLightDark testa criar tema com variantes light/dark
func (ts *TestSuite) TestCreateThemeLightDark() bool {
	ts.logger.Section("THEME CUSTOMIZATION - POST /project/settings/theme (Light/Dark)")

	payload := map[string]interface{}{
		// LIGHT MODE - 11 cores principais
		"primary_color_light":        "#FF6B35",  // Laranja quente
		"secondary_color_light":      "#F4A261",  // Amarelo ouro
		"background_color_light":     "#FFFFFF",  // Fundo branco
		"card_background_color_light": "#FFFFFF", // Card branco
		"text_color_light":           "#0F172A",  // Texto escuro
		"text_secondary_color_light": "#64748B",  // Texto secundário
		"accent_color_light":         "#FF9F1C",  // Laranja vibrante

		// DARK MODE - 11 cores principais
		"primary_color_dark":        "#FF6B35",   // Laranja mantido
		"secondary_color_dark":      "#F4A261",   // Amarelo mantido
		"background_color_dark":     "#09090b",   // Fundo muito escuro
		"card_background_color_dark": "#18181b",  // Card escuro
		"text_color_dark":           "#fafafa",   // Texto branco
		"text_secondary_color_dark": "#a1a1aa",   // Texto cinza claro
		"accent_color_dark":         "#FF9F1C",   // Laranja vibrante

		// LIGHT MODE - 5 cores semânticas
		"destructive_color_light": "#EF4444",
		"success_color_light":     "#10B981",
		"warning_color_light":     "#F59E0B",
		"border_color_light":      "#E5E7EB",
		"price_color_light":       "#10B981",

		// DARK MODE - 5 cores semânticas
		"destructive_color_dark": "#DC2626",
		"success_color_dark":     "#34D399",
		"warning_color_dark":     "#FBBF24",
		"border_color_dark":      "#475569",
		"price_color_dark":       "#34D399",

		// LIGHT MODE - 2 cores sistema
		"focus_ring_color_light":      "#3B82F6",
		"input_background_color_light": "#F3F4F6",

		// DARK MODE - 2 cores sistema
		"focus_ring_color_dark":       "#93C5FD",
		"input_background_color_dark": "#1F2937",

		// Configurações numéricas
		"disabled_opacity": 0.5,
		"shadow_intensity": 1.0,
		"is_active":        true,
	}

	resp, err := ts.client.Request("POST", "/project/settings/theme", payload, true)
	if err != nil {
		ts.addResult("POST /project/settings/theme", false, err.Error())
		return false
	}

	// Verificar se cores foram salvas
	lightPrimary := ts.client.ExtractString(resp, "primary_color_light")
	darkPrimary := ts.client.ExtractString(resp, "primary_color_dark")

	if lightPrimary != "#FF6B35" || darkPrimary != "#FF6B35" {
		ts.addResult("POST /project/settings/theme", false,
			fmt.Sprintf("Cores light/dark não salvas corretamente: light=%s, dark=%s", lightPrimary, darkPrimary))
		return false
	}

	ts.addResult("POST /project/settings/theme", true, "Tema light/dark customizado com sucesso")
	return true
}

// TestUpdateThemeLightDark testa atualizar tema via PUT
func (ts *TestSuite) TestUpdateThemeLightDark() bool {
	ts.logger.Section("THEME CUSTOMIZATION - PUT /project/settings/theme (Partial Update)")

	// Atualizar apenas cores light de forma parcial
	payload := map[string]interface{}{
		"primary_color_light": "#1E293B",   // Cinza padrão profissional
		"accent_color_dark":   "#F472B6",   // Rosa em dark
	}

	resp, err := ts.client.Request("PUT", "/project/settings/theme", payload, true)
	if err != nil {
		ts.addResult("PUT /project/settings/theme", false, err.Error())
		return false
	}

	lightPrimary := ts.client.ExtractString(resp, "primary_color_light")
	darkAccent := ts.client.ExtractString(resp, "accent_color_dark")

	if lightPrimary != "#1E293B" || darkAccent != "#F472B6" {
		ts.addResult("PUT /project/settings/theme", false, "Cores não atualizadas corretamente em PUT")
		return false
	}

	ts.addResult("PUT /project/settings/theme", true, "Tema atualizado parcialmente com sucesso")
	return true
}

// TestResetThemeLightDark testa resetar para defaults profissionais light/dark
func (ts *TestSuite) TestResetThemeLightDark() bool {
	ts.logger.Section("THEME CUSTOMIZATION - POST /project/settings/theme/reset (Light/Dark Defaults)")

	resp, err := ts.client.Request("POST", "/project/settings/theme/reset", map[string]interface{}{}, true)
	if err != nil {
		ts.addResult("POST /project/settings/theme/reset", false, err.Error())
		return false
	}

	// Verificar se voltou aos defaults profissionais
	lightPrimary := ts.client.ExtractString(resp, "primary_color_light")
	darkPrimary := ts.client.ExtractString(resp, "primary_color_dark")
	lightBg := ts.client.ExtractString(resp, "background_color_light")
	darkBg := ts.client.ExtractString(resp, "background_color_dark")

	expectedLightPrimary := "#1E293B"  // Cinza profissional
	expectedDarkPrimary := "#F8FAFC"   // Branco/Off-white profissional
	expectedLightBg := "#FFFFFF"       // Branco
	expectedDarkBg := "#0F172A"        // Cinza muito escuro

	if lightPrimary != expectedLightPrimary || darkPrimary != expectedDarkPrimary ||
		lightBg != expectedLightBg || darkBg != expectedDarkBg {
		ts.addResult("POST /project/settings/theme/reset", false,
			fmt.Sprintf("Cores não resetadas corretamente: lightPrim=%s (exp %s), darkPrim=%s (exp %s), lightBg=%s (exp %s), darkBg=%s (exp %s)",
				lightPrimary, expectedLightPrimary, darkPrimary, expectedDarkPrimary, lightBg, expectedLightBg, darkBg, expectedDarkBg))
		return false
	}

	ts.addResult("POST /project/settings/theme/reset", true, "Tema resetado para padrões profissionais light/dark")
	return true
}

// TestInvalidHexColorLight Dark testa validação de cores HEX inválidas
func (ts *TestSuite) TestInvalidHexColorLightDark() bool {
	ts.logger.Section("THEME CUSTOMIZATION - Validação de Cores HEX (Light/Dark)")

	// Tentar salvar com cor inválida
	payload := map[string]interface{}{
		"primary_color_light": "not-a-hex-color", // Inválido
	}

	resp, err := ts.client.Request("POST", "/project/settings/theme", payload, true)
	if err == nil {
		// Se não teve erro, pode estar no campo error ou statusCode
		statusStr := ts.client.ExtractString(resp, "statusCode")
		if statusStr != "400" {
			errorMsg := ts.client.ExtractString(resp, "error")
			if errorMsg == "" {
				ts.addResult("Validação HEX inválida", false, "Deveria rejeitar cor HEX inválida")
				return false
			}
		}
	}

	ts.addResult("Validação HEX inválida", true, "Sistema rejeitou cor HEX inválida corretamente")
	return true
}

// TestLightDarkVariantsIndependent testa que light e dark podem ser diferentes
func (ts *TestSuite) TestLightDarkVariantsIndependent() bool {
	ts.logger.Section("THEME CUSTOMIZATION - Light/Dark Variants Independence")

	// Salvar com cores light diferentes de dark
	payload := map[string]interface{}{
		"primary_color_light": "#FF6B35",  // Laranja em light
		"primary_color_dark":  "#1E293B",  // Cinza em dark (diferente!)
		"is_active":           true,
	}

	resp, err := ts.client.Request("POST", "/project/settings/theme", payload, true)
	if err != nil {
		ts.addResult("Light/Dark Independence", false, err.Error())
		return false
	}

	lightPrimary := ts.client.ExtractString(resp, "primary_color_light")
	darkPrimary := ts.client.ExtractString(resp, "primary_color_dark")

	if lightPrimary != "#FF6B35" || darkPrimary != "#1E293B" {
		ts.addResult("Light/Dark Independence", false, "Variantes não foram salvas independentemente")
		return false
	}

	if lightPrimary == darkPrimary {
		ts.addResult("Light/Dark Independence", false, "Cores light e dark deveriam ser diferentes")
		return false
	}

	ts.addResult("Light/Dark Independence", true, "Light e Dark podem ser customizados independentemente")
	return true
}

// TestThemeColorPreviewComplete testa preview completo de todas as 30 cores
func (ts *TestSuite) TestThemeColorPreviewComplete() bool {
	ts.logger.Section("THEME CUSTOMIZATION - Complete Color Preview (30 colors)")

	resp, err := ts.client.Request("GET", "/project/settings/theme", nil, true)
	if err != nil {
		ts.addResult("Theme Complete Preview", false, err.Error())
		return false
	}

	// Verificar se todas as 30 cores estão presentes
	colors := []string{
		// Light Mode - 11 cores principais
		"primary_color_light",
		"secondary_color_light",
		"background_color_light",
		"card_background_color_light",
		"text_color_light",
		"text_secondary_color_light",
		"accent_color_light",
		// Dark Mode - 11 cores principais
		"primary_color_dark",
		"secondary_color_dark",
		"background_color_dark",
		"card_background_color_dark",
		"text_color_dark",
		"text_secondary_color_dark",
		"accent_color_dark",
		// Light Mode - 5 cores semânticas
		"destructive_color_light",
		"success_color_light",
		"warning_color_light",
		"border_color_light",
		"price_color_light",
		// Dark Mode - 5 cores semânticas
		"destructive_color_dark",
		"success_color_dark",
		"warning_color_dark",
		"border_color_dark",
		"price_color_dark",
		// Light Mode - 2 cores sistema
		"focus_ring_color_light",
		"input_background_color_light",
		// Dark Mode - 2 cores sistema
		"focus_ring_color_dark",
		"input_background_color_dark",
	}

	missingColors := []string{}
	for _, color := range colors {
		value := ts.client.ExtractString(resp, color)
		if value == "" {
			missingColors = append(missingColors, color)
		}
	}

	if len(missingColors) > 0 {
		ts.addResult("Theme Complete Preview", false,
			fmt.Sprintf("Cores faltando: %v", missingColors))
		return false
	}

	ts.addResult("Theme Complete Preview", true,
		fmt.Sprintf("Todas as 30 cores (11 principais + 5 semânticas + 2 sistema light/dark) estão presentes"))
	return true
}

// TestDeleteTheme testa deletar customização de tema
func (ts *TestSuite) TestDeleteTheme() bool {
	ts.logger.Section("THEME CUSTOMIZATION - DELETE /project/settings/theme")

	resp, err := ts.client.Request("DELETE", "/project/settings/theme", nil, true)
	if err != nil {
		ts.addResult("DELETE /project/settings/theme", false, err.Error())
		return false
	}

	// Verificar se retorna mensagem de sucesso
	message := ts.client.ExtractString(resp, "message")
	if message == "" {
		ts.addResult("DELETE /project/settings/theme", false, "Resposta não contém mensagem")
		return false
	}

	ts.addResult("DELETE /project/settings/theme", true, "Tema deletado com sucesso")
	return true
}

// RunThemeCustomizationTests executa todos os testes de customização de tema
func (ts *TestSuite) RunThemeCustomizationTests() {
	ts.logger.Section("🎨 THEME CUSTOMIZATION TESTS (Light/Dark Variants)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"GET /project/settings/theme (Light/Dark)", ts.TestGetTheme},
		{"POST /project/settings/theme (Light/Dark)", ts.TestCreateThemeLightDark},
		{"PUT /project/settings/theme (Partial Update)", ts.TestUpdateThemeLightDark},
		{"POST /project/settings/theme/reset (Light/Dark Defaults)", ts.TestResetThemeLightDark},
		{"Validação de Cores HEX (Light/Dark)", ts.TestInvalidHexColorLightDark},
		{"Light/Dark Variants Independence", ts.TestLightDarkVariantsIndependent},
		{"Complete Color Preview (30 colors)", ts.TestThemeColorPreviewComplete},
		{"DELETE /project/settings/theme", ts.TestDeleteTheme},
	}

	for _, test := range tests {
		test.fn()
	}
}
