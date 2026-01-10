# ❌ Lista Completa de Testes Faltantes

**Total de testes faltantes: 85 rotas (55.5%)**
**Testes recomendados a implementar: ~53**

---

## 🔴 CRÍTICO (Máxima Prioridade - 31 rotas)

Estas rotas são core features que DEVEM ter testes antes de qualquer deploy.

### 1. Webhooks Twilio (2 rotas)
```
❌ POST /webhook/twilio/status
   - Callback de status de envio do Twilio
   - Impacto: Sistema de notificações quebrado
   - Teste: TestTwilioStatusWebhook()

❌ POST /webhook/twilio/inbound/:orgId/:projectId
   - Callback de mensagem inbound do Twilio
   - Impacto: Recebimento de mensagens não testado
   - Teste: TestTwilioInboundWebhook()
```

### 2. Notificações (7 rotas)
```
❌ POST /notification/send
   - Enviar notificação manualmente
   - Impacto: Sistema de notificações não validado
   - Teste: TestNotificationSend()

❌ POST /notification/event
   - Registrar evento de notificação
   - Impacto: Disparadores não testados
   - Teste: TestNotificationEvent()

❌ GET /notification/logs/:orgId/:projectId
   - Buscar logs de notificações
   - Impacto: Auditoria não testada
   - Teste: TestGetNotificationLogs()

❌ GET /notification/templates/:orgId/:projectId
   - Listar templates de notificação
   - Impacto: Templates não testados
   - Teste: TestGetNotificationTemplates()

❌ POST /notification/template
   - Criar template de notificação
   - Impacto: Criação de templates não validada
   - Teste: TestCreateNotificationTemplate()

❌ PUT /notification/template
   - Atualizar template de notificação
   - Impacto: Edição de templates não testada
   - Teste: TestUpdateNotificationTemplate()

❌ POST /notification/config
   - Configurar notificações
   - Impacto: Configuração não validada
   - Teste: TestCreateNotificationConfig()
```

### 3. Auth/Security (2 rotas)
```
❌ POST /logout
   - Fazer logout e invalidar token
   - Impacto: Segurança de sessão não testada
   - Teste: TestLogout()
   - Nota: Testado em tests_upload_fix.go mas não em main

❌ POST /checkToken
   - Validar token JWT
   - Impacto: Validação de token não validada
   - Teste: TestCheckToken()
```

### 4. Order Status & Progress (2 rotas)
```
❌ GET /order/:id/progress
   - Rastrear progresso do pedido na cozinha
   - Impacto: Rastreamento não testado
   - Teste: TestGetOrderProgress()

❌ PUT /order/:id/status
   - Atualizar status do pedido
   - Impacto: Transição de estados não validada
   - Teste: TestUpdateOrderStatus()
```

### 5. Organization Hard Delete (1 rota)
```
❌ DELETE /organization/:id/permanent
   - Deletar organização permanentemente
   - Impacto: Limpeza de dados crítica não testada
   - Teste: TestPermanentDeleteOrganization()
```

### 6. Seeding/Bootstrap (5 rotas)
```
❌ POST /create-organization
   - Criar organização inicial com setup completo
   - Impacto: Onboarding não testado
   - Teste: TestCreateOrganizationWithSetup()

❌ POST /organization (seeding)
   - Seeding de organização
   - Impacto: Setup automático não testado
   - Teste: TestOrganizationSeeding()

❌ POST /project (seeding)
   - Seeding de projeto
   - Impacto: Setup de projeto não testado
   - Teste: TestProjectSeeding()

❌ POST /user-organization/user/:userId
   - Associar usuário a organização (seeding)
   - Impacto: Relacionamento não testado
   - Teste: TestUserOrganizationSeeding()

❌ POST /user-project/user/:userId
   - Associar usuário a projeto (seeding)
   - Impacto: Relacionamento não testado
   - Teste: TestUserProjectSeeding()
```

### 7. Admin Features (1 rota)
```
❌ POST /admin/reset-passwords
   - Reset de senhas de admins
   - Impacto: Gerenciamento de admin não testado
   - Teste: TestAdminResetPasswords()
```

### 8. Reports/Analytics (5 rotas)
```
❌ GET /reports/occupancy
   - Relatório de ocupação do restaurante
   - Impacto: Analytics não testado
   - Teste: TestOccupancyReport()

❌ GET /reports/reservations
   - Relatório de reservas
   - Impacto: Analytics de reservas não testado
   - Teste: TestReservationsReport()

❌ GET /reports/waitlist
   - Relatório de fila de espera
   - Impacto: Analytics de fila não testado
   - Teste: TestWaitlistReport()

❌ GET /reports/leads
   - Relatório de leads
   - Impacto: CRM não testado
   - Teste: TestLeadsReport()

❌ GET /reports/export/:type
   - Exportar dados em diferentes formatos
   - Impacto: Export não testado
   - Teste: TestExportReport()
```

### 9. User-Organization Relations (4 rotas)
```
❌ DELETE /user-organization/user/:userId/org/:orgId
   - Remover acesso do usuário à organização
   - Impacto: Gerenciamento de acesso não testado
   - Teste: TestRemoveUserFromOrganization()

❌ PUT /user-organization/:id
   - Atualizar relacionamento user-org
   - Impacto: Mudança de roles não testada
   - Teste: TestUpdateUserOrganization()

❌ GET /user-organization/user/:userId
   - Listar organizações do usuário
   - Impacto: Relacionamento não testado
   - Teste: TestGetUserOrganizations()

❌ GET /user-organization/org/:orgId
   - Listar usuários da organização
   - Impacto: Gerenciamento de acesso não testado
   - Teste: TestGetOrganizationUsers()
```

### 10. User-Project Relations (5 rotas)
```
❌ DELETE /user-project/user/:userId/proj/:projectId
   - Remover acesso do usuário ao projeto
   - Impacto: Acesso a projeto não testado
   - Teste: TestRemoveUserFromProject()

❌ PUT /user-project/:id
   - Atualizar relacionamento user-project
   - Impacto: Mudança de roles não testada
   - Teste: TestUpdateUserProject()

❌ GET /user-project/user/:userId
   - Listar projetos do usuário
   - Impacto: Acesso não testado
   - Teste: TestGetUserProjects()

❌ GET /user-project/user/:userId/org/:orgId
   - Listar projetos do usuário em uma org
   - Impacto: Filtro não testado
   - Teste: TestGetUserProjectsByOrg()

❌ GET /user-project/proj/:projectId
   - Listar usuários do projeto
   - Impacto: Acesso do projeto não testado
   - Teste: TestGetProjectUsers()
```

---

## 🟠 ALTO (Alta Prioridade - 15 rotas)

Rotas importantes que devem ter testes em breve.

### 1. Settings/Configuration (5 rotas)
```
❌ GET /project/settings/display
   - Obter configurações de display
   - Teste: TestGetDisplaySettings()

❌ PUT /project/settings/display
   - Atualizar configurações de display
   - Teste: TestUpdateDisplaySettings()

❌ POST /project/settings/display/reset
   - Reset configurações de display
   - Teste: TestResetDisplaySettings()
```

### 2. Theme Customization (5 rotas)
```
❌ GET /project/settings/theme
   - Obter tema customizado
   - Teste: TestGetTheme()
   - Nota: Parcialmente implementado em tests_theme_customization.go

❌ POST /project/settings/theme
   - Criar tema customizado
   - Teste: TestCreateTheme()

❌ PUT /project/settings/theme
   - Atualizar tema customizado
   - Teste: TestUpdateTheme()

❌ POST /project/settings/theme/reset
   - Reset para tema padrão
   - Teste: TestResetTheme()

❌ DELETE /project/settings/theme
   - Deletar tema customizado
   - Teste: TestDeleteTheme()
```

### 3. Menu Advanced (5 rotas)
```
❌ GET /menu/active-now
   - Obter menu ativo no momento
   - Teste: TestGetActiveMenuNow()

❌ GET /menu/active
   - Listar menus ativos
   - Teste: TestGetActiveMenus()

❌ GET /menu/options
   - Obter opções de menu disponíveis
   - Teste: TestGetMenuOptions()

❌ PUT /menu/:id/manual-override
   - Override manual de seleção de menu
   - Teste: TestMenuManualOverride()

❌ DELETE /menu/manual-override
   - Remover override manual
   - Teste: TestRemoveMenuManualOverride()
```

---

## 🟡 MÉDIO (Prioridade Normal - 39 rotas)

Rotas importantes mas menos críticas para o negócio.

### 1. Product Advanced (10 rotas)
```
❌ GET /product/purchase/:id
   - Obter detalhes de produto para compra
   - Teste: TestGetProductPurchaseDetails()

❌ GET /product/by-tag
   - Listar produtos por tag
   - Teste: TestGetProductsByTag()

❌ GET /product/:id/tags
   - Listar tags do produto
   - Teste: TestGetProductTags()

❌ POST /product/:id/tags
   - Adicionar tag ao produto
   - Teste: TestAddProductTag()

❌ DELETE /product/:id/tags/:tagId
   - Remover tag do produto
   - Teste: TestRemoveProductTag()

❌ PUT /product/:id/order
   - Reordenar produto
   - Teste: TestReorderProduct()

❌ PUT /product/:id/status
   - Atualizar status do produto
   - Teste: TestUpdateProductStatus()

❌ GET /product/type/:type
   - Listar produtos por tipo
   - Teste: TestGetProductsByType()

❌ GET /product/category/:categoryId
   - Listar produtos da categoria
   - Teste: TestGetProductsByCategory()

❌ GET /product/subcategory/:subcategoryId
   - Listar produtos da subcategoria
   - Teste: TestGetProductsBySubcategory()
```

### 2. Upload Generic (1 rota)
```
❌ POST /upload/:category/image
   - Upload genérico de imagem
   - Teste: TestGenericImageUpload()
```

### 3. Category Hierarchy (4 rotas)
```
❌ GET /category/active
   - Listar categorias ativas
   - Teste: TestGetActiveCategories()

❌ GET /category/menu/:menuId
   - Listar categorias do menu
   - Teste: TestGetCategoriesByMenu()

❌ PUT /category/:id/order
   - Reordenar categoria
   - Teste: TestReorderCategory()

❌ PUT /category/:id/status
   - Atualizar status da categoria
   - Teste: TestUpdateCategoryStatus()
```

### 4. Subcategory Hierarchy (6 rotas)
```
❌ GET /subcategory/active
   - Listar subcategorias ativas
   - Teste: TestGetActiveSubcategories()

❌ GET /subcategory/category/:categoryId
   - Listar subcategorias da categoria
   - Teste: TestGetSubcategoriesByCategory()

❌ PUT /subcategory/:id/order
   - Reordenar subcategoria
   - Teste: TestReorderSubcategory()

❌ PUT /subcategory/:id/status
   - Atualizar status da subcategoria
   - Teste: TestUpdateSubcategoryStatus()

❌ POST /subcategory/:id/category/:categoryId
   - Adicionar categoria à subcategoria
   - Teste: TestAddCategoryToSubcategory()

❌ DELETE /subcategory/:id/category/:categoryId
   - Remover categoria da subcategoria
   - Teste: TestRemoveCategoryFromSubcategory()

❌ GET /subcategory/:id/categories
   - Listar categorias da subcategoria
   - Teste: TestGetSubcategoryCategories()
```

### 5. Tag Filtering (2 rotas)
```
❌ GET /tag/active
   - Listar tags ativas
   - Teste: TestGetActiveTags()

❌ GET /tag/entity/:entityType
   - Listar tags por tipo de entidade
   - Teste: TestGetTagsByEntity()
```

### 6. User Advanced (2 rotas)
```
❌ GET /user/group/:id
   - Obter grupo de usuários
   - Teste: TestGetUserGroup()

❌ POST /user/:id/organizations-projects
   - Criar acesso a org/project
   - Teste: TestCreateUserAccess()
```

### 7. Environment Filtering (1 rota)
```
❌ GET /environment/active
   - Listar ambientes ativos
   - Teste: TestGetActiveEnvironments()
```

### 8. Organization Management (3 rotas)
```
❌ GET /organization/email
   - Obter organização por email
   - Teste: TestGetOrganizationByEmail()

❌ PUT /organization/:id
   - Atualizar organização
   - Teste: TestUpdateOrganization()

❌ DELETE /organization/:id
   - Deletar organização (soft delete)
   - Teste: TestDeleteOrganization()
```

### 9. Project Management (3 rotas)
```
❌ GET /project/:id
   - Obter projeto por ID
   - Teste: TestGetProjectById()

❌ PUT /project/:id
   - Atualizar projeto
   - Teste: TestUpdateProject()

❌ DELETE /project/:id
   - Deletar projeto
   - Teste: TestDeleteProject()
```

### 10. Public Routes (1 rota)
```
❌ GET /public/project/:orgId/:projId
   - Obter info pública do projeto
   - Teste: TestGetPublicProjectInfo()
```

### 11. Upload File Serving (2 rotas)
```
❌ GET /uploads/:orgId/:projId/:category/:filename
   - Servir arquivo de upload
   - Teste: TestServeUploadedFile()

❌ GET /static/:category/:filename
   - Servir arquivo estático (compat)
   - Teste: TestServeStaticFile()
```

### 12. Public Routes Partial (1 rota)
```
⚠️ GET /public/times/:orgId/:projId
   - Obter horários disponíveis
   - Status: Parcialmente testado
   - Teste: TestGetPublicTimes() (reforçar cobertura)
```

---

## 📊 RESUMO DE TESTES FALTANTES

| Prioridade | Categoria | Rotas | Testes | Status |
|------------|-----------|-------|--------|--------|
| 🔴 Crítico | Webhooks | 2 | 2 | ❌ |
| 🔴 Crítico | Notificações | 7 | 7 | ❌ |
| 🔴 Crítico | Auth | 2 | 2 | ❌ |
| 🔴 Crítico | Order Status | 2 | 2 | ❌ |
| 🔴 Crítico | Org Hard Delete | 1 | 1 | ❌ |
| 🔴 Crítico | Seeding | 5 | 5 | ❌ |
| 🔴 Crítico | Admin | 1 | 1 | ❌ |
| 🔴 Crítico | Reports | 5 | 5 | ❌ |
| 🔴 Crítico | User-Org | 4 | 4 | ❌ |
| 🔴 Crítico | User-Proj | 5 | 5 | ❌ |
| **Subtotal Crítico** | | **34** | **34** | |
| 🟠 Alto | Settings | 3 | 3 | ❌ |
| 🟠 Alto | Theme | 5 | 5 | ❌ |
| 🟠 Alto | Menu Adv | 5 | 5 | ❌ |
| **Subtotal Alto** | | **13** | **13** | |
| 🟡 Médio | Product Adv | 10 | 10 | ❌ |
| 🟡 Médio | Category | 4 | 4 | ❌ |
| 🟡 Médio | Subcategory | 7 | 7 | ❌ |
| 🟡 Médio | Tags | 2 | 2 | ❌ |
| 🟡 Médio | User Adv | 2 | 2 | ❌ |
| 🟡 Médio | Environment | 1 | 1 | ❌ |
| 🟡 Médio | Org Mgmt | 3 | 3 | ❌ |
| 🟡 Médio | Project Mgmt | 3 | 3 | ❌ |
| 🟡 Médio | Public | 2 | 2 | ❌ |
| 🟡 Médio | Upload File | 2 | 2 | ❌ |
| 🟡 Médio | Misc | 1 | 1 | ❌ |
| **Subtotal Médio** | | **38** | **38** | |
| | **TOTAL** | **85** | **85** | ❌ |

---

## 🎯 ESTIMATIVA DE ESFORÇO

```
Crítico (34 testes):
  - 2-3 horas de implementação
  - Deve ser feito ANTES de deploy

Alto (13 testes):
  - 1.5-2 horas de implementação
  - Deve ser feito no próximo sprint

Médio (38 testes):
  - 3-4 horas de implementação
  - Pode ser feito em médio prazo

TOTAL ESTIMADO: 6-9 horas de trabalho
```

---

## ✅ RECOMENDAÇÃO FINAL

1. **Implementar 34 testes críticos IMEDIATAMENTE** (antes de deploy)
2. **Implementar 13 testes altos em MÉDIO PRAZO** (próximas 2 semanas)
3. **Implementar 38 testes médios em LONGO PRAZO** (próximo mês)

Com isso, atingirá ~80% de cobertura com 68 + 34 + 13 = **115 testes de 153 rotas (75%)**

