# 🗺️ Roadmap Priorizado de Testes - Plano de Ação

**Objetivo**: Aumentar cobertura de testes de 44.5% para >80% em 3 sprints
**Testes Atuais**: 68/153 (44.5%)
**Testes Alvo**: 115+/153 (75%+)
**Testes a Adicionar**: ~47 testes

---

## 📅 CRONOGRAMA RECOMENDADO

```
Sprint 1 (Crítico): 2-3 dias    → 34 testes
Sprint 2 (Alto):    3-4 dias    → 13 testes
Sprint 3 (Médio):   3-4 dias    → 38 testes

Total: ~8-11 dias de trabalho (~1-2 semanas)
Meta: 75%+ de cobertura ao final do Sprint 3
```

---

## 🔴 SPRINT 1: CRÍTICO (2-3 Dias)

**Objetivo**: Implementar 34 testes críticos
**Responsável**: Backend Lead + QA Senior
**Verificação**: Cobertura deve ir de 44.5% para ~66%

### Sprint 1.1: Autenticação & Segurança (2 testes - 2 horas)

```go
// File: tests.go
// Add after TestLogout() around line 2310

func (ts *TestSuite) TestCheckToken() {
    ts.logger.Subsection("POST /checkToken - Validar token JWT")

    // Usar token válido do login
    payload := map[string]interface{}{
        "token": ts.config.Headers.OrgID, // usar token real
    }

    resp, err := ts.client.Request("POST", "/checkToken", payload, true)
    if err != nil {
        ts.addResult("POST /checkToken", false, err.Error())
        return
    }

    if status := ts.client.GetLastStatus(); status != 200 {
        ts.addResult("POST /checkToken", false, fmt.Sprintf("Status: %d", status))
        return
    }

    ts.addResult("POST /checkToken", true, "Token validado")
}

func (ts *TestSuite) TestLogoutFull() {
    ts.logger.Subsection("POST /logout - Full logout flow")

    // Fazer login
    loginPayload := map[string]interface{}{
        "email":    ts.config.TestUser.Email,
        "password": ts.config.TestUser.Password,
    }

    loginResp, err := ts.client.Request("POST", "/login", loginPayload, false)
    if err != nil {
        ts.addResult("POST /logout", false, "Login failed")
        return
    }

    // Fazer logout
    resp, err := ts.client.Request("POST", "/logout", nil, true)
    if err != nil {
        ts.addResult("POST /logout", false, err.Error())
        return
    }

    ts.addResult("POST /logout", true, "Logout successful")
}
```

**Checklist**:
- [ ] TestCheckToken() implementado
- [ ] TestLogoutFull() implementado
- [ ] Ambos passando
- [ ] Build sem erros

---

### Sprint 1.2: Webhooks Twilio (2 testes - 3 horas)

```go
// Adicionar depois de TestLogoutFull()

func (ts *TestSuite) TestTwilioStatusWebhook() {
    ts.logger.Subsection("POST /webhook/twilio/status - Status callback")

    // Simular callback do Twilio
    payload := map[string]interface{}{
        "MessageSid":   "SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "MessageStatus": "delivered",
        "ErrorCode":    nil,
        "ErrorMessage": "",
    }

    resp, err := ts.client.Request("POST", "/webhook/twilio/status", payload, false)
    if err != nil {
        ts.addResult("POST /webhook/twilio/status", false, err.Error())
        return
    }

    ts.addResult("POST /webhook/twilio/status", true, "Webhook processed")
}

func (ts *TestSuite) TestTwilioInboundWebhook() {
    ts.logger.Subsection("POST /webhook/twilio/inbound - Inbound message")

    orgID := ts.config.Headers.OrgID
    projID := ts.config.Headers.ProjID

    payload := map[string]interface{}{
        "MessageSid":    "SMxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "AccountSid":    "ACxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx",
        "From":          "+5511999999999",
        "To":            "+5511888888888",
        "Body":          "Test message",
        "NumMedia":      "0",
    }

    path := fmt.Sprintf("/webhook/twilio/inbound/%s/%s", orgID, projID)
    resp, err := ts.client.Request("POST", path, payload, false)
    if err != nil {
        ts.addResult("POST /webhook/twilio/inbound", false, err.Error())
        return
    }

    ts.addResult("POST /webhook/twilio/inbound", true, "Inbound processed")
}
```

**Checklist**:
- [ ] TestTwilioStatusWebhook() implementado
- [ ] TestTwilioInboundWebhook() implementado
- [ ] Webhooks respondendo corretamente
- [ ] Build sem erros

---

### Sprint 1.3: Notificações (7 testes - 5 horas)

```go
// Adicionar depois dos webhooks

func (ts *TestSuite) TestCreateNotificationTemplate() {
    ts.logger.Subsection("POST /notification/template - Criar template")

    payload := map[string]interface{}{
        "name":       "Welcome Template",
        "type":       "sms",
        "content":    "Bem-vindo ao {restaurant_name}!",
        "variables": []string{"restaurant_name"},
    }

    resp, err := ts.client.Request("POST", "/notification/template", payload, true)
    if err != nil {
        ts.addResult("POST /notification/template", false, err.Error())
        return
    }

    ts.addResult("POST /notification/template", true, "Template criado")
}

func (ts *TestSuite) TestSendNotification() {
    ts.logger.Subsection("POST /notification/send - Enviar notificação")

    payload := map[string]interface{}{
        "type":    "sms",
        "to":      "+5511999999999",
        "message": "Sua reserva foi confirmada!",
    }

    resp, err := ts.client.Request("POST", "/notification/send", payload, true)
    if err != nil {
        ts.addResult("POST /notification/send", false, err.Error())
        return
    }

    ts.addResult("POST /notification/send", true, "Notificação enviada")
}

// ... (adicionar mais 5 testes de notificação)
```

**Checklist**:
- [ ] Todos os 7 testes de notificação implementados
- [ ] Templates funcionando
- [ ] Eventos registrando corretamente
- [ ] Build sem erros

---

### Sprint 1.4: Order Status & Progress (2 testes - 2 horas)

```go
// Adicionar após testes de pedido

func (ts *TestSuite) TestGetOrderProgress() {
    ts.logger.Subsection("GET /order/:id/progress - Status de progresso")

    // Criar pedido primeiro
    orderPayload := map[string]interface{}{
        "customer_id": "123e4567-e89b-12d3-a456-426614174010",
        "items": []map[string]interface{}{
            {
                "product_id": "abc123",
                "quantity":   1,
                "price":      50.0,
            },
        },
        "status":        "pending",
        "total_amount":  50.0,
        "organization_id": ts.config.Headers.OrgID,
        "project_id":    ts.config.Headers.ProjID,
    }

    createResp, _ := ts.client.Request("POST", "/order", orderPayload, true)
    orderID := ts.client.ExtractString(createResp["data"].(map[string]interface{}), "id")

    // Verificar progresso
    path := fmt.Sprintf("/order/%s/progress", orderID)
    resp, err := ts.client.Request("GET", path, nil, true)
    if err != nil {
        ts.addResult(fmt.Sprintf("GET %s", path), false, err.Error())
        return
    }

    ts.addResult("GET /order/:id/progress", true, "Progress retrieved")
}

func (ts *TestSuite) TestUpdateOrderStatus() {
    ts.logger.Subsection("PUT /order/:id/status - Atualizar status")

    // Usar pedido existente
    payload := map[string]interface{}{
        "status": "completed",
    }

    resp, err := ts.client.Request("PUT", "/order/test-id/status", payload, true)
    if err != nil {
        ts.addResult("PUT /order/:id/status", false, err.Error())
        return
    }

    ts.addResult("PUT /order/:id/status", true, "Status atualizado")
}
```

**Checklist**:
- [ ] TestGetOrderProgress() implementado
- [ ] TestUpdateOrderStatus() implementado
- [ ] Status transitions funcionando
- [ ] Build sem erros

---

### Sprint 1.5: Seeding & Bootstrap (5 testes - 4 horas)

```go
// Adicionar ao final do arquivo

func (ts *TestSuite) TestCreateOrganizationWithSetup() {
    ts.logger.Subsection("POST /create-organization - Bootstrap completo")

    payload := map[string]interface{}{
        "name":     "Novo Restaurante",
        "password": "senha123456",
    }

    resp, err := ts.client.Request("POST", "/create-organization", payload, false)
    if err != nil {
        ts.addResult("POST /create-organization", false, err.Error())
        return
    }

    ts.addResult("POST /create-organization", true, "Organização criada com setup")
}

func (ts *TestSuite) TestOrganizationSeeding() {
    ts.logger.Subsection("POST /organization - Seeding")

    payload := map[string]interface{}{
        "name": "Test Org",
    }

    resp, err := ts.client.Request("POST", "/organization", payload, false)
    if err != nil {
        ts.addResult("POST /organization (seed)", false, err.Error())
        return
    }

    ts.addResult("POST /organization (seed)", true, "Org seeded")
}

// ... (adicionar mais 3 testes de seeding)
```

**Checklist**:
- [ ] Todos os 5 testes de seeding implementados
- [ ] Bootstrap flow completo testado
- [ ] User-org e user-project seeding funcionando
- [ ] Build sem erros

---

### Sprint 1.6: User-Organization Relations (4 testes - 3 horas)

```go
// Adicionar relações user-org

func (ts *TestSuite) TestGetUserOrganizations() {
    ts.logger.Subsection("GET /user-organization/user/:userId")

    userID := "test-user-id"
    path := fmt.Sprintf("/user-organization/user/%s", userID)

    resp, err := ts.client.Request("GET", path, nil, true)
    if err != nil {
        ts.addResult(path, false, err.Error())
        return
    }

    ts.addResult(path, true, "Orgs retrieved")
}

// ... (adicionar mais 3 testes)
```

**Checklist**:
- [ ] Todos os 4 testes de user-org implementados
- [ ] Relações funcionando
- [ ] Acesso validado
- [ ] Build sem erros

---

### Sprint 1.7: User-Project Relations (4 testes - 3 horas)

```go
// Adicionar relações user-project
// Seguir mesmo padrão de user-organization
```

**Checklist**:
- [ ] Todos os 4 testes de user-project implementados
- [ ] Relações funcionando
- [ ] Acesso validado
- [ ] Build sem erros

---

### Sprint 1.8: Org Management & Admin (4 testes - 3 horas)

```go
// POST /admin/reset-passwords
// DELETE /organization/:id/permanent
// Etc
```

**Checklist**:
- [ ] Admin reset-passwords testado
- [ ] Hard delete organização testado
- [ ] Reports básicos testados
- [ ] Build sem erros

---

## ✅ Sprint 1 - Conclusão

**Testes Implementados**: 34
**Cobertura Esperada**: 66% (102/153)
**Tempo Total**: 2-3 dias
**Validação**: Rodar `go run . -verbose` e confirmar que os 34 testes novos passam

```bash
# No fim do Sprint 1
cd LEP-teste-back
go run . -verbose > sprint1_results.txt
# Verificar: ~102 testes passando
```

---

## 🟠 SPRINT 2: ALTO (3-4 Dias)

**Objetivo**: Implementar 13 testes altos
**Cobertura Esperada**: 72% (115/153)

### Sprint 2.1: Settings & Configuration (3 testes)
- GET /project/settings/display
- PUT /project/settings/display
- POST /project/settings/display/reset

**Tempo**: 2 horas

### Sprint 2.2: Theme Customization (5 testes)
- GET /project/settings/theme
- POST /project/settings/theme
- PUT /project/settings/theme
- POST /project/settings/theme/reset
- DELETE /project/settings/theme

**Tempo**: 3 horas

### Sprint 2.3: Menu Advanced (5 testes)
- GET /menu/active-now
- GET /menu/active
- GET /menu/options
- PUT /menu/:id/manual-override
- DELETE /menu/manual-override

**Tempo**: 3 horas

**Total Sprint 2**: 8 horas (1 dia)

---

## 🟡 SPRINT 3: MÉDIO (3-4 Dias)

**Objetivo**: Implementar 38 testes médios
**Cobertura Esperada**: 75%+ (115+/153)

### Sprint 3.1: Product Advanced (10 testes)
**Tempo**: 4 horas

### Sprint 3.2: Category & Subcategory Hierarchies (10 testes)
**Tempo**: 4 horas

### Sprint 3.3: Remaining Filters & Relationships (18 testes)
**Tempo**: 5 horas

**Total Sprint 3**: 13 horas (1.5 dias)

---

## 📊 SUMÁRIO DO ROADMAP

| Sprint | Dias | Testes | Cobertura | Status |
|--------|------|--------|-----------|--------|
| 1 (Crítico) | 2-3 | 34 | 66% | 🔴 |
| 2 (Alto) | 1 | 13 | 72% | 🟠 |
| 3 (Médio) | 1.5 | 38 | 75% | 🟡 |
| **TOTAL** | **4.5-5.5** | **85** | **75%+** | ✅ |

---

## 🎯 CHECKLIST GLOBAL

### Antes de começar Sprint 1:
- [ ] Documentos compartilhados com time
- [ ] Responsabilidades atribuídas
- [ ] Ambiente de teste setup
- [ ] CI/CD pipeline pronto

### Durante Sprint 1:
- [ ] Implementar 34 testes diariamente
- [ ] Validar cobertura frequentemente
- [ ] Resolver blockers imediatamente
- [ ] Documentar problemas encontrados

### Fim de Sprint 1:
- [ ] Todos os 34 testes passando
- [ ] Cobertura em 66%+
- [ ] Build sem erros
- [ ] Code review completo

### Sprint 2 & 3:
- [ ] Mesma cadência de qualidade
- [ ] Atingir 75%+ de cobertura
- [ ] Validar novo features

### Objetivo Final:
- [ ] 115+ testes implementados
- [ ] 75%+ de cobertura
- [ ] Features críticas 100% testadas
- [ ] Pronto para produção

---

## 💡 DICAS DE IMPLEMENTAÇÃO

1. **Código DRY**: Criar helpers para testes repetitivos
2. **Fixtures**: Usar dados pré-criados quando possível
3. **Assertions**: Sempre validar status HTTP + resposta
4. **Cleanup**: Deletar dados de teste ao final
5. **Logging**: Adicionar `ts.addResult()` para cada validação

---

## 📞 SUPPORT

Se encontrar blockers:
1. Consultar ROUTE_COVERAGE_ANALYSIS.md para detalhes
2. Consultar MISSING_TESTS.md para exemplos
3. Verificar se há problemas no backend (usar BACKEND_FIXES_PROMPT.md)
4. Comunicar delay imediatamente ao lead

---

**Meta Final**: >75% cobertura em ~1-2 semanas
**Sucesso Esperado**: Sim - Todos os testes críticos estarão cobertos

