package main

import (
	"fmt"
	"sort"
	"strings"
)

// =============================================================================
// VERIFICAÇÃO DE COBERTURA DE ROTAS
// Lista todas as rotas testadas e identifica rotas faltantes
// =============================================================================

// RouteInfo informações sobre uma rota
type RouteInfo struct {
	Method   string
	Path     string
	Tested   bool
	Category string
}

// RouteCoverage resultado da análise de cobertura
type RouteCoverage struct {
	TestedRoutes   []RouteInfo
	MissingRoutes  []RouteInfo
	Coverage       float64
	TotalRoutes    int
	TestedCount    int
}

// GetExpectedRoutes retorna todas as rotas esperadas do backend
func GetExpectedRoutes() []RouteInfo {
	return []RouteInfo{
		// Auth
		{Method: "POST", Path: "/login", Category: "Auth"},
		{Method: "POST", Path: "/logout", Category: "Auth"},
		{Method: "POST", Path: "/register", Category: "Auth"},
		{Method: "POST", Path: "/checkToken", Category: "Auth"},
		{Method: "POST", Path: "/refresh-token", Category: "Auth"},
		{Method: "POST", Path: "/forgot-password", Category: "Auth"},
		{Method: "POST", Path: "/reset-password", Category: "Auth"},
		{Method: "POST", Path: "/admin/login", Category: "Auth"},
		{Method: "POST", Path: "/client/login", Category: "Auth"},
		{Method: "POST", Path: "/create-organization", Category: "Auth"},

		// Health
		{Method: "GET", Path: "/ping", Category: "Health"},

		// User
		{Method: "GET", Path: "/user", Category: "User"},
		{Method: "GET", Path: "/user/:id", Category: "User"},
		{Method: "POST", Path: "/user", Category: "User"},
		{Method: "PUT", Path: "/user/:id", Category: "User"},
		{Method: "DELETE", Path: "/user/:id", Category: "User"},
		{Method: "GET", Path: "/user/:id/organizations-projects", Category: "User"},

		// Admin User
		{Method: "POST", Path: "/admin/admin-user", Category: "Admin User"},
		{Method: "GET", Path: "/admin/admin-user", Category: "Admin User"},
		{Method: "DELETE", Path: "/admin/admin-user/:id", Category: "Admin User"},

		// Client User
		{Method: "POST", Path: "/admin/client-user", Category: "Client User"},
		{Method: "GET", Path: "/admin/client-user", Category: "Client User"},
		{Method: "DELETE", Path: "/admin/client-user/:id", Category: "Client User"},

		// Organization
		{Method: "GET", Path: "/organization", Category: "Organization"},
		{Method: "GET", Path: "/organization/:id", Category: "Organization"},
		{Method: "POST", Path: "/organization", Category: "Organization"},
		{Method: "PUT", Path: "/organization/:id", Category: "Organization"},
		{Method: "DELETE", Path: "/organization/:id", Category: "Organization"},
		{Method: "GET", Path: "/organization/active", Category: "Organization"},

		// Project
		{Method: "GET", Path: "/project", Category: "Project"},
		{Method: "GET", Path: "/project/:id", Category: "Project"},
		{Method: "POST", Path: "/project", Category: "Project"},
		{Method: "PUT", Path: "/project/:id", Category: "Project"},
		{Method: "DELETE", Path: "/project/:id", Category: "Project"},
		{Method: "GET", Path: "/project/active", Category: "Project"},
		{Method: "GET", Path: "/project/organization/:orgId", Category: "Project"},

		// User-Organization
		{Method: "GET", Path: "/user-organization/user/:userId", Category: "Association"},
		{Method: "GET", Path: "/user-organization/org/:orgId", Category: "Association"},

		// User-Project
		{Method: "GET", Path: "/user-project/user/:userId", Category: "Association"},
		{Method: "GET", Path: "/user-project/proj/:projectId", Category: "Association"},
		{Method: "GET", Path: "/user-project/user/:userId/org/:orgId", Category: "Association"},

		// Roles
		{Method: "GET", Path: "/role", Category: "Roles"},
		{Method: "GET", Path: "/role/system", Category: "Roles"},
		{Method: "GET", Path: "/role/my-permissions", Category: "Roles"},

		// Menu
		{Method: "GET", Path: "/menu", Category: "Menu"},
		{Method: "GET", Path: "/menu/:id", Category: "Menu"},
		{Method: "POST", Path: "/menu", Category: "Menu"},
		{Method: "PUT", Path: "/menu/:id", Category: "Menu"},
		{Method: "DELETE", Path: "/menu/:id", Category: "Menu"},
		{Method: "GET", Path: "/menu/active", Category: "Menu"},

		// Category
		{Method: "GET", Path: "/category", Category: "Category"},
		{Method: "GET", Path: "/category/:id", Category: "Category"},
		{Method: "POST", Path: "/category", Category: "Category"},
		{Method: "PUT", Path: "/category/:id", Category: "Category"},
		{Method: "DELETE", Path: "/category/:id", Category: "Category"},

		// Product
		{Method: "GET", Path: "/product", Category: "Product"},
		{Method: "GET", Path: "/product/:id", Category: "Product"},
		{Method: "POST", Path: "/product", Category: "Product"},
		{Method: "PUT", Path: "/product/:id", Category: "Product"},
		{Method: "DELETE", Path: "/product/:id", Category: "Product"},
		{Method: "GET", Path: "/product/category/:id", Category: "Product"},

		// Tag
		{Method: "GET", Path: "/tag", Category: "Tag"},
		{Method: "GET", Path: "/tag/:id", Category: "Tag"},
		{Method: "POST", Path: "/tag", Category: "Tag"},
		{Method: "PUT", Path: "/tag/:id", Category: "Tag"},
		{Method: "DELETE", Path: "/tag/:id", Category: "Tag"},

		// Environment
		{Method: "GET", Path: "/environment", Category: "Environment"},
		{Method: "GET", Path: "/environment/:id", Category: "Environment"},
		{Method: "POST", Path: "/environment", Category: "Environment"},
		{Method: "PUT", Path: "/environment/:id", Category: "Environment"},
		{Method: "DELETE", Path: "/environment/:id", Category: "Environment"},
		{Method: "GET", Path: "/environment/active", Category: "Environment"},

		// Table
		{Method: "GET", Path: "/table", Category: "Table"},
		{Method: "GET", Path: "/table/:id", Category: "Table"},
		{Method: "POST", Path: "/table", Category: "Table"},
		{Method: "PUT", Path: "/table/:id", Category: "Table"},
		{Method: "DELETE", Path: "/table/:id", Category: "Table"},

		// Customer
		{Method: "GET", Path: "/customer", Category: "Customer"},
		{Method: "GET", Path: "/customer/:id", Category: "Customer"},
		{Method: "POST", Path: "/customer", Category: "Customer"},
		{Method: "PUT", Path: "/customer/:id", Category: "Customer"},
		{Method: "DELETE", Path: "/customer/:id", Category: "Customer"},

		// Reservation
		{Method: "GET", Path: "/reservation", Category: "Reservation"},
		{Method: "GET", Path: "/reservation/:id", Category: "Reservation"},
		{Method: "POST", Path: "/reservation", Category: "Reservation"},
		{Method: "PUT", Path: "/reservation/:id", Category: "Reservation"},
		{Method: "DELETE", Path: "/reservation/:id", Category: "Reservation"},

		// Order
		{Method: "GET", Path: "/order", Category: "Order"},
		{Method: "GET", Path: "/order/:id", Category: "Order"},
		{Method: "POST", Path: "/order", Category: "Order"},
		{Method: "PUT", Path: "/order/:id", Category: "Order"},
		{Method: "DELETE", Path: "/order/:id", Category: "Order"},
		{Method: "PUT", Path: "/order/:id/status", Category: "Order"},

		// OrderItem
		{Method: "GET", Path: "/order-item", Category: "OrderItem"},
		{Method: "POST", Path: "/order-item", Category: "OrderItem"},
		{Method: "PUT", Path: "/order-item/:id", Category: "OrderItem"},
		{Method: "DELETE", Path: "/order-item/:id", Category: "OrderItem"},

		// Notification
		{Method: "GET", Path: "/notification", Category: "Notification"},
		{Method: "GET", Path: "/notification/:id", Category: "Notification"},
		{Method: "POST", Path: "/notification", Category: "Notification"},
		{Method: "PUT", Path: "/notification/:id/read", Category: "Notification"},
		{Method: "PUT", Path: "/notification/read-all", Category: "Notification"},
		{Method: "DELETE", Path: "/notification/:id", Category: "Notification"},

		// Settings
		{Method: "GET", Path: "/settings", Category: "Settings"},
		{Method: "PUT", Path: "/settings", Category: "Settings"},
		{Method: "GET", Path: "/project-settings", Category: "Settings"},
		{Method: "PUT", Path: "/project-settings", Category: "Settings"},

		// Theme
		{Method: "GET", Path: "/theme", Category: "Theme"},
		{Method: "PUT", Path: "/theme", Category: "Theme"},

		// Upload
		{Method: "POST", Path: "/upload", Category: "Upload"},
		{Method: "POST", Path: "/upload/image", Category: "Upload"},
		{Method: "DELETE", Path: "/upload/:id", Category: "Upload"},

		// Admin Images
		{Method: "GET", Path: "/admin/images/stats", Category: "Admin"},
		{Method: "POST", Path: "/admin/images/cleanup", Category: "Admin"},

		// Public
		{Method: "GET", Path: "/public/menu", Category: "Public"},
		{Method: "GET", Path: "/public/menu/:slug", Category: "Public"},
		{Method: "GET", Path: "/public/categories", Category: "Public"},
	}
}

// GetTestedRoutes retorna as rotas que são testadas pelos testes atuais
func GetTestedRoutes() map[string]bool {
	return map[string]bool{
		// Auth
		"POST /login":               true,
		"POST /logout":              true,
		"POST /register":            true,
		"POST /checkToken":          true,
		"POST /admin/login":         true,
		"POST /client/login":        true,
		"POST /create-organization": true,

		// Health
		"GET /ping": true,

		// User
		"GET /user":                            true,
		"GET /user/:id":                        true,
		"POST /user":                           true,
		"PUT /user/:id":                        true,
		"DELETE /user/:id":                     true,
		"GET /user/:id/organizations-projects": true,

		// Admin User
		"POST /admin/admin-user":      true,
		"GET /admin/admin-user":       true,
		"DELETE /admin/admin-user/:id": true,

		// Client User
		"POST /admin/client-user":      true,
		"GET /admin/client-user":       true,
		"DELETE /admin/client-user/:id": true,

		// Organization
		"GET /organization":        true,
		"GET /organization/:id":    true,
		"POST /organization":       true,
		"PUT /organization/:id":    true,
		"DELETE /organization/:id": true,
		"GET /organization/active": true,

		// Project
		"GET /project":                      true,
		"GET /project/:id":                  true,
		"POST /project":                     true,
		"PUT /project/:id":                  true,
		"DELETE /project/:id":               true,
		"GET /project/active":               true,
		"GET /project/organization/:orgId":  true,

		// Associations
		"GET /user-organization/user/:userId":         true,
		"GET /user-organization/org/:orgId":            true,
		"GET /user-project/user/:userId":               true,
		"GET /user-project/proj/:projectId":            true,
		"GET /user-project/user/:userId/org/:orgId":    true,

		// Roles
		"GET /role":                true,
		"GET /role/system":         true,
		"GET /role/my-permissions": true,

		// Menu
		"GET /menu":        true,
		"GET /menu/:id":    true,
		"POST /menu":       true,
		"PUT /menu/:id":    true,
		"DELETE /menu/:id": true,
		"GET /menu/active": true,

		// Category
		"GET /category":        true,
		"GET /category/:id":    true,
		"POST /category":       true,
		"PUT /category/:id":    true,
		"DELETE /category/:id": true,

		// Product
		"GET /product":              true,
		"GET /product/:id":          true,
		"POST /product":             true,
		"PUT /product/:id":          true,
		"DELETE /product/:id":       true,

		// Tag
		"GET /tag":        true,
		"GET /tag/:id":    true,
		"POST /tag":       true,
		"PUT /tag/:id":    true,
		"DELETE /tag/:id": true,

		// Environment
		"GET /environment":        true,
		"GET /environment/:id":    true,
		"POST /environment":       true,
		"PUT /environment/:id":    true,
		"DELETE /environment/:id": true,
		"GET /environment/active": true,

		// Table
		"GET /table":        true,
		"GET /table/:id":    true,
		"POST /table":       true,
		"PUT /table/:id":    true,
		"DELETE /table/:id": true,

		// Customer
		"GET /customer":        true,
		"GET /customer/:id":    true,
		"POST /customer":       true,
		"PUT /customer/:id":    true,
		"DELETE /customer/:id": true,

		// Reservation
		"GET /reservation":        true,
		"GET /reservation/:id":    true,
		"POST /reservation":       true,
		"PUT /reservation/:id":    true,
		"DELETE /reservation/:id": true,

		// Order
		"GET /order":            true,
		"GET /order/:id":        true,
		"POST /order":           true,
		"PUT /order/:id":        true,
		"DELETE /order/:id":     true,
		"PUT /order/:id/status": true,

		// OrderItem
		"POST /order-item":       true,
		"PUT /order-item/:id":    true,
		"DELETE /order-item/:id": true,

		// Notification
		"GET /notification":           true,
		"POST /notification":          true,
		"PUT /notification/:id/read":  true,
		"PUT /notification/read-all":  true,
		"DELETE /notification/:id":    true,

		// Settings
		"GET /settings":         true,
		"PUT /settings":         true,
		"GET /project-settings": true,
		"PUT /project-settings": true,

		// Theme
		"GET /theme": true,
		"PUT /theme": true,

		// Upload
		"POST /upload":       true,
		"POST /upload/image": true,

		// Admin
		"GET /admin/images/stats":   true,
		"POST /admin/images/cleanup": true,

		// Public
		"GET /public/menu":       true,
		"GET /public/categories": true,
	}
}

// AnalyzeRouteCoverage analisa a cobertura de rotas
func AnalyzeRouteCoverage() RouteCoverage {
	expectedRoutes := GetExpectedRoutes()
	testedRoutes := GetTestedRoutes()

	var coverage RouteCoverage
	coverage.TotalRoutes = len(expectedRoutes)

	for _, route := range expectedRoutes {
		key := fmt.Sprintf("%s %s", route.Method, route.Path)
		if testedRoutes[key] {
			route.Tested = true
			coverage.TestedRoutes = append(coverage.TestedRoutes, route)
			coverage.TestedCount++
		} else {
			route.Tested = false
			coverage.MissingRoutes = append(coverage.MissingRoutes, route)
		}
	}

	if coverage.TotalRoutes > 0 {
		coverage.Coverage = float64(coverage.TestedCount) / float64(coverage.TotalRoutes) * 100
	}

	return coverage
}

// RunRouteCoverageReport executa e imprime o relatório de cobertura
func (ts *TestSuite) RunRouteCoverageReport() {
	ts.logger.Section("RELATÓRIO DE COBERTURA DE ROTAS")

	coverage := AnalyzeRouteCoverage()

	ts.logger.Info("Cobertura: %.1f%% (%d/%d rotas)",
		coverage.Coverage, coverage.TestedCount, coverage.TotalRoutes)
	ts.logger.Info("")

	if len(coverage.MissingRoutes) > 0 {
		ts.logger.Warn("ROTAS NÃO TESTADAS:")

		// Agrupar por categoria
		byCategory := make(map[string][]RouteInfo)
		for _, route := range coverage.MissingRoutes {
			byCategory[route.Category] = append(byCategory[route.Category], route)
		}

		// Ordenar categorias
		var categories []string
		for cat := range byCategory {
			categories = append(categories, cat)
		}
		sort.Strings(categories)

		for _, cat := range categories {
			routes := byCategory[cat]
			ts.logger.Warn("  [%s]", cat)
			for _, route := range routes {
				ts.logger.Warn("    - %s %s", route.Method, route.Path)
			}
		}
		ts.logger.Info("")
	}

	if len(coverage.MissingRoutes) == 0 {
		ts.logger.Success("Todas as rotas conhecidas estão cobertas pelos testes!")
		ts.addResult("Route Coverage", true, "100% - Todas as rotas cobertas")
	} else if coverage.Coverage >= 85.0 {
		// Cobertura aceitável (>= 85%)
		ts.addResult("Route Coverage", true,
			fmt.Sprintf("%.1f%% - %d rotas faltando (aceitável)", coverage.Coverage, len(coverage.MissingRoutes)))
	} else {
		ts.addResult("Route Coverage", false,
			fmt.Sprintf("%.1f%% - %d rotas faltando", coverage.Coverage, len(coverage.MissingRoutes)))
	}
}

// GenerateMissingTestsPrompt gera um prompt para criar testes faltantes
func GenerateMissingTestsPrompt() string {
	coverage := AnalyzeRouteCoverage()

	if len(coverage.MissingRoutes) == 0 {
		return "Todas as rotas estão cobertas pelos testes."
	}

	var sb strings.Builder
	sb.WriteString("## Rotas que Precisam de Testes\n\n")
	sb.WriteString(fmt.Sprintf("Cobertura atual: %.1f%%\n\n", coverage.Coverage))

	// Agrupar por categoria
	byCategory := make(map[string][]RouteInfo)
	for _, route := range coverage.MissingRoutes {
		byCategory[route.Category] = append(byCategory[route.Category], route)
	}

	for cat, routes := range byCategory {
		sb.WriteString(fmt.Sprintf("### %s\n", cat))
		for _, route := range routes {
			sb.WriteString(fmt.Sprintf("- `%s %s`\n", route.Method, route.Path))
		}
		sb.WriteString("\n")
	}

	sb.WriteString("### Ação Recomendada\n")
	sb.WriteString("Criar testes para as rotas listadas acima para aumentar a cobertura.\n")

	return sb.String()
}
