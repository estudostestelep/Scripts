# Instruções para Executar o Seed

## Problema Atual
Sua instalação do Go está corrompida e precisa ser reinstalada.

## Solução em 3 Passos

### Passo 1: Desinstalar Go Atual

1. Abra o **Painel de Controle**
2. Vá em **Programas e Recursos**
3. Encontre **Go Programming Language** na lista
4. Clique com botão direito e selecione **Desinstalar**
5. Aguarde a desinstalação

OU use PowerShell como Administrador:

```powershell
# PowerShell como Admin
Remove-Item -Recurse -Force C:\Go
```

---

### Passo 2: Baixar e Instalar Go 1.24.11

1. Acesse: https://go.dev/dl/
2. Baixe: **go1.24.11.windows-amd64.msi**
3. Execute o instalador
4. Clique em **Next > Next > Install**
5. Aguarde a instalação
6. Clique em **Finish**

---

### Passo 3: Executar o Seed

Abra um **novo terminal PowerShell** (feche o antigo) e execute:

```powershell
cd "C:\Users\pablo\OneDrive\Área de Trabalho\Trabalho Be Growth\Projetos\LEP\LEP-Script\LEP-execute-seed"

# Verificar se Go está instalado
go version

# Atualizar dependências
go mod tidy

# Executar seed
go run main.go
```

---

## Resultado Esperado

Você verá mensagens assim:

```
========== Passo 1: Criando Organização ==========
Organização criada: Fattoria

========== Passo 2: Criando Menus ==========
Menu criado: Menu Principal
Menu criado: Menu de Almoço
Menu criado: Menu da Noite

========== Passo 3: Criando Categories ==========
Category criada: Entradas
Category criada: Massas
... (continua)

========== Resumo Final ==========
✓ Criados: ~80 registros
✓ Pulados: ~5 registros
✗ Falhados: 0 registros
```

---

## Se Tiver Problemas

### Erro: "go: command not found"
- Feche e reabra o terminal
- Ou reinicie o computador

### Erro: "connection refused"
- Verifique se o backend está rodando em localhost:8080
- Inicie o backend antes de executar o seed

### Erro: "organization_id required"
- Certifique-se de que o backend está configurado corretamente
- Verifique se o usuário pablo@lep.com existe

---

## Arquivos Criados

Após executar o seed com sucesso, você terá no banco:

- 3 Menus (Principal, Almoço, Noite)
- 6 Categories (Entradas, Massas, Carnes, Peixes, Sobremesas, Bebidas)
- 12 Subcategories (com relacionamentos corretos)
- 36 Products (incluindo 5 vinhos com metadados completos)
- 9 Tables
- 3 Environments
- 4 Users
- 5 Customers
- 4 Reservations
- 4 Tags
- 1 Settings
- 3 NotificationTemplates
- 1 ThemeCustomization

**Taxa de Sucesso Esperada: 100%**
