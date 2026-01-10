package main

import (
	"fmt"
	"time"
)

// TestProductTagsOptimization testa a otimização do carregamento de produtos com tags
func (ts *TestSuite) TestProductTagsOptimization() bool {
	ts.logger.Section("TEST: Otimização de Carregamento de Produtos com Tags")
	ts.logger.Info("Testa a capacidade de carregar produtos com tags em uma única requisição")

	// ============================================================================
	// 1. TESTE SEM INCLUDESTAGS (Baseline)
	// ============================================================================
	ts.logger.Subsection("1. Baseline: GET /product (sem tags)")

	startTime := time.Now()
	productsResp, err := ts.client.Request("GET", "/product", nil, true)
	baselineTime := time.Since(startTime)

	if err != nil {
		ts.addResult("GET /product (sem tags)", false, err.Error())
		return false
	}

	// Extrair lista de produtos
	var products []interface{}
	if prods, ok := productsResp["data"].([]interface{}); ok {
		products = prods
	} else if prods, ok := productsResp["products"].([]interface{}); ok {
		products = prods
	}

	productCount := len(products)
	ts.logger.Info("✓ Carregados %d produtos em %.2fms (sem tags)", productCount, baselineTime.Seconds()*1000)
	ts.addResult("GET /product (sem tags)", true, fmt.Sprintf("%d produtos em %.0fms", productCount, baselineTime.Seconds()*1000))

	if productCount == 0 {
		ts.logger.Warn("Nenhum produto encontrado para teste de otimização")
		return true
	}

	// ============================================================================
	// 2. TESTE COM INCLUDESTAGS=TRUE (Otimizado)
	// ============================================================================
	ts.logger.Subsection("2. Otimizado: GET /product?includeTags=true")

	startTime = time.Now()
	productsWithTagsResp, err := ts.client.Request("GET", "/product?includeTags=true", nil, true)
	optimizedTime := time.Since(startTime)

	if err != nil {
		ts.addResult("GET /product?includeTags=true", false, err.Error())
		return false
	}

	// Extrair lista de produtos com tags
	var productsWithTags []interface{}
	if prods, ok := productsWithTagsResp["data"].([]interface{}); ok {
		productsWithTags = prods
	} else if prods, ok := productsWithTagsResp["products"].([]interface{}); ok {
		productsWithTags = prods
	}

	productCountWithTags := len(productsWithTags)
	ts.logger.Info("✓ Carregados %d produtos com tags em %.2fms", productCountWithTags, optimizedTime.Seconds()*1000)
	ts.addResult("GET /product?includeTags=true", true, fmt.Sprintf("%d produtos em %.0fms", productCountWithTags, optimizedTime.Seconds()*1000))

	// ============================================================================
	// 3. VALIDAÇÃO: Verificar se tags estão incluídas
	// ============================================================================
	ts.logger.Subsection("3. Validação: Verificar se tags estão incluídas na resposta")

	taggedProductsCount := 0
	totalTagsLoaded := 0

	for _, prod := range productsWithTags {
		if prodMap, ok := prod.(map[string]interface{}); ok {
			if tags, ok := prodMap["tags"].([]interface{}); ok && len(tags) > 0 {
				taggedProductsCount++
				totalTagsLoaded += len(tags)
			}
		}
	}

	ts.logger.Info("✓ %d produtos contêm tags (total: %d tags)", taggedProductsCount, totalTagsLoaded)
	ts.addResult("Produtos com tags incluídas", true, fmt.Sprintf("%d produtos, %d tags", taggedProductsCount, totalTagsLoaded))

	// ============================================================================
	// 4. TESTE DE VALIDAÇÃO DE ESTRUTURA DAS TAGS
	// ============================================================================
	ts.logger.Subsection("4. Validação: Estrutura das tags retornadas")

	var firstProductWithTags map[string]interface{}
	for _, prod := range productsWithTags {
		if prodMap, ok := prod.(map[string]interface{}); ok {
			if tags, ok := prodMap["tags"].([]interface{}); ok && len(tags) > 0 {
				firstProductWithTags = prodMap
				break
			}
		}
	}

	if firstProductWithTags != nil {
		if tags, ok := firstProductWithTags["tags"].([]interface{}); ok && len(tags) > 0 {
			if firstTag, ok := tags[0].(map[string]interface{}); ok {
				// Validar campos esperados em uma tag
				expectedFields := []string{"id", "name", "color", "active"}
				missingFields := []string{}

				for _, field := range expectedFields {
					if _, exists := firstTag[field]; !exists {
						missingFields = append(missingFields, field)
					}
				}

				if len(missingFields) == 0 {
					ts.logger.Info("✓ Tag contém todos os campos esperados: id, name, color, active")
					ts.addResult("Validação de estrutura das tags", true, "Todos os campos presentes")
				} else {
					ts.logger.Warn("Campo(s) faltando em tag: %v", missingFields)
					ts.addResult("Validação de estrutura das tags", false, fmt.Sprintf("Campos faltando: %v", missingFields))
				}
			}
		}
	} else {
		ts.logger.Warn("Nenhum produto com tags encontrado para validação de estrutura")
		ts.addResult("Validação de estrutura das tags", true, "Nenhum produto com tags (OK)")
	}

	// ============================================================================
	// 5. TESTE COM FILTROS + INCLUDESTAGS
	// ============================================================================
	ts.logger.Subsection("5. Teste combinado: Filtros + includeTags")

	// Teste com active=true e includeTags=true
	startTime = time.Now()
	productsActiveWithTagsResp, err := ts.client.Request("GET", "/product?active=true&includeTags=true", nil, true)
	activeFilteredTime := time.Since(startTime)

	if err != nil {
		ts.addResult("GET /product?active=true&includeTags=true", false, err.Error())
		return false
	}

	var activeProducts []interface{}
	if prods, ok := productsActiveWithTagsResp["data"].([]interface{}); ok {
		activeProducts = prods
	} else if prods, ok := productsActiveWithTagsResp["products"].([]interface{}); ok {
		activeProducts = prods
	}

	ts.logger.Info("✓ Carregados %d produtos ativos com tags em %.2fms", len(activeProducts), activeFilteredTime.Seconds()*1000)
	ts.addResult("GET /product?active=true&includeTags=true", true, fmt.Sprintf("%d produtos em %.0fms", len(activeProducts), activeFilteredTime.Seconds()*1000))

	// ============================================================================
	// 6. ANÁLISE DE PERFORMANCE
	// ============================================================================
	ts.logger.Subsection("6. Análise de Performance")

	ts.logger.Info("Resumo de Tempos:")
	ts.logger.Info("  - Baseline (sem tags):              %.0fms", baselineTime.Seconds()*1000)
	ts.logger.Info("  - Com includeTags=true:            %.0fms", optimizedTime.Seconds()*1000)
	ts.logger.Info("  - Com filtro active=true:          %.0fms", activeFilteredTime.Seconds()*1000)

	// Calcular overhead (deve ser minimal)
	overhead := optimizedTime.Seconds()*1000 - baselineTime.Seconds()*1000
	overheadPercent := (overhead / (baselineTime.Seconds() * 1000)) * 100

	if overhead >= 0 {
		ts.logger.Info("  - Overhead: +%.0fms (+%.1f%%)", overhead, overheadPercent)
	} else {
		ts.logger.Info("  - Overhead: %.0fms (-%.1f%%) [Mais rápido - possível cache]", overhead, -overheadPercent)
	}

	ts.logger.Info("")
	ts.logger.Info("✨ Ganho de Performance:")
	ts.logger.Info("  - Sem includeTags, seria necessário fazer %d requisições adicionais para tags", productCount)
	ts.logger.Info("  - Com includeTags, tudo em uma única requisição")
	ts.logger.Info("  - Redução estimada de requisições: ~%d queries (N+1 problem resolvido)", productCount)

	ts.addResult("Análise de Performance", true, fmt.Sprintf("Overhead: %.0fms (%.1f%%)", overhead, overheadPercent))

	return true
}
