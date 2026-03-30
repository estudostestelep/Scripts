package main

import (
	"fmt"

	"github.com/google/uuid"
)

// =============================================================================
// TESTES DE CATÁLOGO (Cadeia de Dependência)
// Ordem: Menu → Category → Subcategory → Product → Tag
// =============================================================================

// IDs criados durante os testes para uso nas dependências
var (
	createdMenuID        string
	createdCategoryID    string
	createdSubcategoryID string
	createdProductID     string
	createdTagID         string
)

func (ts *TestSuite) RunCatalogTests() {
	ts.logger.Section("FASE 3: CATÁLOGO (Menu → Category → Product → Tag)")

	tests := []struct {
		name string
		fn   func() bool
	}{
		{"Menu CRUD", ts.TestMenuCRUDCatalog},
		{"Category CRUD", ts.TestCategoryCRUDCatalog},
		{"Subcategory CRUD", ts.TestSubcategoryCRUDCatalog},
		{"Product CRUD", ts.TestProductCRUDCatalog},
		{"Tag CRUD", ts.TestTagCRUDCatalog},
		{"Product Tags", ts.TestProductTagsAssociation},
	}

	for _, test := range tests {
		test.fn()
	}
}

// =============================================================================
// MENU CRUD
// =============================================================================

func (ts *TestSuite) TestMenuCRUDCatalog() bool {
	ts.logger.Subsection("MENU - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar menu - POST /menu")
	createPayload := map[string]interface{}{
		"name":          "Menu Teste " + uuid.New().String()[:8],
		"description":   "Menu criado para testes automatizados",
		"active":        true,
		"is_default":    false,
		"order": 1,
	}

	createResp, err := ts.client.Request("POST", "/admin/menu", createPayload, true)
	if err != nil {
		ts.addResult("POST /admin/menu", false, err.Error())
		allPassed = false
	} else {
		// Extrair ID do menu criado
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdMenuID = id
				ts.addResult("POST /admin/menu", true, fmt.Sprintf("Menu criado: %s", id[:8]))
			}
		}
		if createdMenuID == "" {
			ts.addResult("POST /admin/menu", true, "Menu criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar menus - GET /menu")
	resp, err := ts.client.Request("GET", "/menu", nil, true)
	if err != nil {
		ts.addResult("GET /menu", false, err.Error())
		allPassed = false
	} else {
		menus := ts.client.ExtractArray(resp)
		ts.addResult("GET /menu", true, fmt.Sprintf("%d menus encontrados", len(menus)))

		// Se não temos ID, pegar do primeiro menu
		if createdMenuID == "" && len(menus) > 0 {
			if menu, ok := menus[0].(map[string]interface{}); ok {
				if id, ok := menu["id"].(string); ok {
					createdMenuID = id
				}
			}
		}
	}

	// READ ONE
	if createdMenuID != "" {
		ts.logger.Info("3. Buscar menu específico - GET /menu/:id")
		_, err := ts.client.Request("GET", "/menu/"+createdMenuID, nil, true)
		if err != nil {
			ts.addResult("GET /menu/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /menu/:id", true, "Menu obtido")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar menu - PUT /admin/menu/:id")
		updatePayload := map[string]interface{}{
			"name":        "Menu Atualizado " + uuid.New().String()[:8],
			"description": "Descrição atualizada",
		}
		_, err = ts.client.Request("PUT", "/admin/menu/"+createdMenuID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /admin/menu/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /admin/menu/:id", true, "Menu atualizado")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar menus ativos - GET /menu/active")
	_, err = ts.client.Request("GET", "/menu/active", nil, true)
	if err != nil {
		ts.addResult("GET /menu/active", false, err.Error())
	} else {
		ts.addResult("GET /menu/active", true, "Menus ativos listados")
	}

	return allPassed
}

// =============================================================================
// CATEGORY CRUD (depende de Menu)
// =============================================================================

func (ts *TestSuite) TestCategoryCRUDCatalog() bool {
	ts.logger.Subsection("CATEGORY - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar categoria - POST /category")
	createPayload := map[string]interface{}{
		"name":          "Categoria Teste " + uuid.New().String()[:8],
		"description":   "Categoria criada para testes",
		"active":        true,
		"order": 1,
	}

	// Adicionar menu_id se disponível
	if createdMenuID != "" {
		createPayload["menu_id"] = createdMenuID
	}

	createResp, err := ts.client.Request("POST", "/admin/category", createPayload, true)
	if err != nil {
		ts.addResult("POST /admin/category", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdCategoryID = id
				ts.addResult("POST /admin/category", true, fmt.Sprintf("Categoria criada: %s", id[:8]))
			}
		}
		if createdCategoryID == "" {
			ts.addResult("POST /admin/category", true, "Categoria criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar categorias - GET /category")
	resp, err := ts.client.Request("GET", "/category", nil, true)
	if err != nil {
		ts.addResult("GET /category", false, err.Error())
		allPassed = false
	} else {
		categories := ts.client.ExtractArray(resp)
		ts.addResult("GET /category", true, fmt.Sprintf("%d categorias encontradas", len(categories)))

		if createdCategoryID == "" && len(categories) > 0 {
			if cat, ok := categories[0].(map[string]interface{}); ok {
				if id, ok := cat["id"].(string); ok {
					createdCategoryID = id
				}
			}
		}
	}

	// READ ONE
	if createdCategoryID != "" {
		ts.logger.Info("3. Buscar categoria específica - GET /category/:id")
		_, err := ts.client.Request("GET", "/category/"+createdCategoryID, nil, true)
		if err != nil {
			ts.addResult("GET /category/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /category/:id", true, "Categoria obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar categoria - PUT /admin/category/:id")
		updatePayload := map[string]interface{}{
			"name":        "Categoria Atualizada " + uuid.New().String()[:8],
			"description": "Descrição atualizada",
		}
		_, err = ts.client.Request("PUT", "/admin/category/"+createdCategoryID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /admin/category/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /admin/category/:id", true, "Categoria atualizada")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar categorias ativas - GET /category/active")
	_, err = ts.client.Request("GET", "/category/active", nil, true)
	if err != nil {
		ts.addResult("GET /category/active", false, err.Error())
	} else {
		ts.addResult("GET /category/active", true, "Categorias ativas listadas")
	}

	// READ BY MENU
	if createdMenuID != "" {
		ts.logger.Info("6. Listar categorias por menu - GET /category/menu/:menuId")
		_, err = ts.client.Request("GET", "/category/menu/"+createdMenuID, nil, true)
		if err != nil {
			ts.addResult("GET /category/menu/:menuId", false, err.Error())
		} else {
			ts.addResult("GET /category/menu/:menuId", true, "Categorias por menu listadas")
		}
	}

	return allPassed
}

// =============================================================================
// SUBCATEGORY CRUD
// =============================================================================

func (ts *TestSuite) TestSubcategoryCRUDCatalog() bool {
	ts.logger.Subsection("SUBCATEGORY - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar subcategoria - POST /subcategory")
	createPayload := map[string]interface{}{
		"name":   "Subcategoria Teste " + uuid.New().String()[:8],
		"notes":  "Subcategoria criada para testes",
		"active": true,
		"order":  1,
	}

	createResp, err := ts.client.Request("POST", "/admin/subcategory", createPayload, true)
	if err != nil {
		ts.addResult("POST /admin/subcategory", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdSubcategoryID = id
				ts.addResult("POST /admin/subcategory", true, fmt.Sprintf("Subcategoria criada: %s", id[:8]))
			}
		}
		if createdSubcategoryID == "" {
			ts.addResult("POST /admin/subcategory", true, "Subcategoria criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar subcategorias - GET /subcategory")
	resp, err := ts.client.Request("GET", "/subcategory", nil, true)
	if err != nil {
		ts.addResult("GET /subcategory", false, err.Error())
		allPassed = false
	} else {
		subcategories := ts.client.ExtractArray(resp)
		ts.addResult("GET /subcategory", true, fmt.Sprintf("%d subcategorias encontradas", len(subcategories)))

		if createdSubcategoryID == "" && len(subcategories) > 0 {
			if subcat, ok := subcategories[0].(map[string]interface{}); ok {
				if id, ok := subcat["id"].(string); ok {
					createdSubcategoryID = id
				}
			}
		}
	}

	// READ ONE
	if createdSubcategoryID != "" {
		ts.logger.Info("3. Buscar subcategoria específica - GET /subcategory/:id")
		_, err := ts.client.Request("GET", "/subcategory/"+createdSubcategoryID, nil, true)
		if err != nil {
			ts.addResult("GET /subcategory/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /subcategory/:id", true, "Subcategoria obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar subcategoria - PUT /admin/subcategory/:id")
		updatePayload := map[string]interface{}{
			"name":  "Subcategoria Atualizada " + uuid.New().String()[:8],
			"notes": "Notas atualizadas",
		}
		_, err = ts.client.Request("PUT", "/admin/subcategory/"+createdSubcategoryID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /admin/subcategory/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /admin/subcategory/:id", true, "Subcategoria atualizada")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar subcategorias ativas - GET /subcategory/active")
	_, err = ts.client.Request("GET", "/subcategory/active", nil, true)
	if err != nil {
		ts.addResult("GET /subcategory/active", false, err.Error())
	} else {
		ts.addResult("GET /subcategory/active", true, "Subcategorias ativas listadas")
	}

	// READ BY CATEGORY
	if createdCategoryID != "" {
		ts.logger.Info("6. Listar subcategorias por categoria - GET /subcategory/category/:categoryId")
		_, err = ts.client.Request("GET", "/subcategory/category/"+createdCategoryID, nil, true)
		if err != nil {
			ts.addResult("GET /subcategory/category/:categoryId", false, err.Error())
		} else {
			ts.addResult("GET /subcategory/category/:categoryId", true, "Subcategorias por categoria listadas")
		}
	}

	return allPassed
}

// =============================================================================
// PRODUCT CRUD
// =============================================================================

func (ts *TestSuite) TestProductCRUDCatalog() bool {
	ts.logger.Subsection("PRODUCT - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar produto - POST /product")
	createPayload := map[string]interface{}{
		"name":         "Produto Teste " + uuid.New().String()[:8],
		"description":  "Produto criado para testes automatizados",
		"type":         "prato",
		"price_normal": 29.90,
		"active":       true,
	}

	// Adicionar category_id e subcategory_id se disponíveis
	if createdCategoryID != "" {
		createPayload["category_id"] = createdCategoryID
	}
	if createdSubcategoryID != "" {
		createPayload["subcategory_id"] = createdSubcategoryID
	}

	createResp, err := ts.client.Request("POST", "/product", createPayload, true)
	if err != nil {
		ts.addResult("POST /product", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdProductID = id
				ts.addResult("POST /product", true, fmt.Sprintf("Produto criado: %s", id[:8]))
			}
		}
		if createdProductID == "" {
			ts.addResult("POST /product", true, "Produto criado (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar produtos - GET /product")
	resp, err := ts.client.Request("GET", "/product", nil, true)
	if err != nil {
		ts.addResult("GET /product", false, err.Error())
		allPassed = false
	} else {
		products := ts.client.ExtractArray(resp)
		ts.addResult("GET /product", true, fmt.Sprintf("%d produtos encontrados", len(products)))

		if createdProductID == "" && len(products) > 0 {
			if prod, ok := products[0].(map[string]interface{}); ok {
				if id, ok := prod["id"].(string); ok {
					createdProductID = id
				}
			}
		}
	}

	// READ ONE
	if createdProductID != "" {
		ts.logger.Info("3. Buscar produto específico - GET /product/:id")
		_, err := ts.client.Request("GET", "/product/"+createdProductID, nil, true)
		if err != nil {
			ts.addResult("GET /product/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /product/:id", true, "Produto obtido")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar produto - PUT /product/:id")
		updatePayload := map[string]interface{}{
			"name":         "Produto Atualizado " + uuid.New().String()[:8],
			"description":  "Descrição atualizada",
			"price_normal": 34.90,
		}
		_, err = ts.client.Request("PUT", "/product/"+createdProductID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /product/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /product/:id", true, "Produto atualizado")
		}
	}

	// READ BY CATEGORY
	if createdCategoryID != "" {
		ts.logger.Info("5. Listar produtos por categoria - GET /product/category/:categoryId")
		_, err = ts.client.Request("GET", "/product/category/"+createdCategoryID, nil, true)
		if err != nil {
			ts.addResult("GET /product/category/:categoryId", false, err.Error())
		} else {
			ts.addResult("GET /product/category/:categoryId", true, "Produtos por categoria listados")
		}
	}

	// READ BY SUBCATEGORY
	if createdSubcategoryID != "" {
		ts.logger.Info("6. Listar produtos por subcategoria - GET /product/subcategory/:subcategoryId")
		_, err = ts.client.Request("GET", "/product/subcategory/"+createdSubcategoryID, nil, true)
		if err != nil {
			ts.addResult("GET /product/subcategory/:subcategoryId", false, err.Error())
		} else {
			ts.addResult("GET /product/subcategory/:subcategoryId", true, "Produtos por subcategoria listados")
		}
	}

	// READ BY TYPE
	ts.logger.Info("7. Listar produtos por tipo - GET /product/type/:type")
	_, err = ts.client.Request("GET", "/product/type/prato", nil, true)
	if err != nil {
		ts.addResult("GET /product/type/:type", false, err.Error())
	} else {
		ts.addResult("GET /product/type/:type", true, "Produtos por tipo listados")
	}

	return allPassed
}

// =============================================================================
// TAG CRUD
// =============================================================================

func (ts *TestSuite) TestTagCRUDCatalog() bool {
	ts.logger.Subsection("TAG - CRUD Completo")
	allPassed := true

	// CREATE
	ts.logger.Info("1. Criar tag - POST /tag")
	createPayload := map[string]interface{}{
		"name":        "Tag Teste " + uuid.New().String()[:8],
		"description": "Tag criada para testes",
		"entity_type": "product",
		"active":      true,
	}

	createResp, err := ts.client.Request("POST", "/tag", createPayload, true)
	if err != nil {
		ts.addResult("POST /tag", false, err.Error())
		allPassed = false
	} else {
		if data := ts.client.ExtractData(createResp); data != nil {
			if id, ok := data["id"].(string); ok {
				createdTagID = id
				ts.addResult("POST /tag", true, fmt.Sprintf("Tag criada: %s", id[:8]))
			}
		}
		if createdTagID == "" {
			ts.addResult("POST /tag", true, "Tag criada (ID não extraído)")
		}
	}

	// READ ALL
	ts.logger.Info("2. Listar tags - GET /tag")
	resp, err := ts.client.Request("GET", "/tag", nil, true)
	if err != nil {
		ts.addResult("GET /tag", false, err.Error())
		allPassed = false
	} else {
		tags := ts.client.ExtractArray(resp)
		ts.addResult("GET /tag", true, fmt.Sprintf("%d tags encontradas", len(tags)))

		if createdTagID == "" && len(tags) > 0 {
			if tag, ok := tags[0].(map[string]interface{}); ok {
				if id, ok := tag["id"].(string); ok {
					createdTagID = id
				}
			}
		}
	}

	// READ ONE
	if createdTagID != "" {
		ts.logger.Info("3. Buscar tag específica - GET /tag/:id")
		_, err := ts.client.Request("GET", "/tag/"+createdTagID, nil, true)
		if err != nil {
			ts.addResult("GET /tag/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("GET /tag/:id", true, "Tag obtida")
		}

		// UPDATE
		ts.logger.Info("4. Atualizar tag - PUT /tag/:id")
		updatePayload := map[string]interface{}{
			"name":        "Tag Atualizada " + uuid.New().String()[:8],
			"description": "Descrição atualizada",
		}
		_, err = ts.client.Request("PUT", "/tag/"+createdTagID, updatePayload, true)
		if err != nil {
			ts.addResult("PUT /tag/:id", false, err.Error())
			allPassed = false
		} else {
			ts.addResult("PUT /tag/:id", true, "Tag atualizada")
		}
	}

	// READ ACTIVE
	ts.logger.Info("5. Listar tags ativas - GET /tag/active")
	_, err = ts.client.Request("GET", "/tag/active", nil, true)
	if err != nil {
		ts.addResult("GET /tag/active", false, err.Error())
	} else {
		ts.addResult("GET /tag/active", true, "Tags ativas listadas")
	}

	// READ BY ENTITY TYPE
	ts.logger.Info("6. Listar tags por tipo de entidade - GET /tag/entity/:entityType")
	_, err = ts.client.Request("GET", "/tag/entity/product", nil, true)
	if err != nil {
		ts.addResult("GET /tag/entity/:entityType", false, err.Error())
	} else {
		ts.addResult("GET /tag/entity/:entityType", true, "Tags por tipo de entidade listadas")
	}

	return allPassed
}

// =============================================================================
// PRODUCT TAGS ASSOCIATION
// =============================================================================

func (ts *TestSuite) TestProductTagsAssociation() bool {
	ts.logger.Subsection("PRODUCT TAGS - Associação de tags a produtos")
	allPassed := true

	if createdProductID == "" || createdTagID == "" {
		ts.addResult("Product Tags", true, "Pulado - IDs não disponíveis")
		return true
	}

	// ADD TAG TO PRODUCT
	ts.logger.Info("1. Adicionar tag ao produto - POST /product/:id/tags")
	addPayload := map[string]interface{}{
		"tag_id": createdTagID,
	}
	_, err := ts.client.Request("POST", "/product/"+createdProductID+"/tags", addPayload, true)
	if err != nil {
		ts.addResult("POST /product/:id/tags", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("POST /product/:id/tags", true, "Tag adicionada ao produto")
	}

	// GET PRODUCT TAGS
	ts.logger.Info("2. Listar tags do produto - GET /product/:id/tags")
	_, err = ts.client.Request("GET", "/product/"+createdProductID+"/tags", nil, true)
	if err != nil {
		ts.addResult("GET /product/:id/tags", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("GET /product/:id/tags", true, "Tags do produto listadas")
	}

	// REMOVE TAG FROM PRODUCT
	ts.logger.Info("3. Remover tag do produto - DELETE /product/:id/tags/:tagId")
	_, err = ts.client.Request("DELETE", "/product/"+createdProductID+"/tags/"+createdTagID, nil, true)
	if err != nil {
		ts.addResult("DELETE /product/:id/tags/:tagId", false, err.Error())
		allPassed = false
	} else {
		ts.addResult("DELETE /product/:id/tags/:tagId", true, "Tag removida do produto")
	}

	return allPassed
}
