# 🚀 Quick Start Guide - LEP Seeder v2.0

## 5-Minute Setup

### Step 1: Verify Backend is Running
```bash
# Terminal 1 - Verificar se LEP-Back está rodando
curl http://localhost:8080/ping
# Esperado: "pong" ou status 200
```

### Step 2: Run the Seeder
```bash
# Terminal 2 - Executar seeder
cd LEP-Script/LEP-execute-seed
go run .
```

### Step 3: Monitor Output
```
✅ Organization criada
✅ Menus criados (3x)
✅ Categories criadas (6x)
✅ Subcategories criadas (12x)
✅ Environments criados (3x)
✅ Tables criadas (9x)
✅ Products criados (33x)
✅ Users criados (4x) ← NEW
✅ Customers criados (5x) ← NEW
✅ Tags criadas (4x) ← NEW
✅ Reservations criadas (4x) ← NEW
```

---

## Command Reference

### Execute Both Seeds (Default)
```bash
go run .
# Executa: seed-fattoria.json + seed-data.json
```

### Execute Only Fattoria
```bash
go run . -file seed-fattoria.json
# Cria: 73 items (org, menus, products, users, customers, reservations, tags)
```

### Execute Only Data
```bash
go run . -file seed-data.json
# Cria: 73 items (diferentes restaurante)
```

### Verbose Mode (Debug)
```bash
go run . -verbose
# Mostra: Todos os request/response payloads em JSON
```

### Build Binary
```bash
go build -o lep-seeder .
./lep-seeder           # Run without compiling each time
```

---

## What Gets Created

### Fattoria Restaurant (seed-fattoria.json)
```
Organization: Fattoria (Italian, Toscana)
├── Staff (4 users)
│   ├── João Silva (admin)
│   ├── Maria Santos (manager)
│   ├── Carlos Oliveira (waiter)
│   └── Ana Costa (kitchen)
├── Customers (5)
│   ├── Pedro Rossi (VIP - window seat preference)
│   ├── Lucia Ferreira (Vegetarian)
│   ├── Roberto Martins (Executive)
│   ├── Fernanda Alves (Seafood allergy)
│   └── Michel Dubois (Wine enthusiast)
├── Menu (3 menus, 6 categories, 12 subcategories)
├── Products (33 dishes + wines)
│   └── Tags (4)
│       ├── Vegetariano
│       ├── Sem Glúten
│       ├── Especial da Casa
│       └── Picanço
├── Environments (3)
│   └── Tables (9 mesas)
└── Reservations (4 confirmed)
    ├── Birthday (4 people)
    ├── Romance (2 people)
    ├── Business lunch (3 people)
    └── Family dinner (5 people)
```

### Total Entities Created
- **Organization:** 1
- **Menus:** 3
- **Categories:** 6
- **Subcategories:** 12
- **Environments:** 3
- **Tables:** 9
- **Products:** 33
- **Tags:** 4
- **Users:** 4
- **Customers:** 5
- **Reservations:** 4

**Per File:** ~73 items
**Both Files:** ~146 items total

---

## Idempotency (Safe to Run Multiple Times)

```bash
# First run: Creates 73 items
go run .
# Output: [✓] Criados: 73, [⏭] Já existiam: 0, [✗] Erros: 0

# Second run: Detects existing items, skips them
go run .
# Output: [✓] Criados: 0, [⏭] Já existiam: 73, [✗] Erros: 0

# Third run: Same as second
go run .
# Output: [✓] Criados: 0, [⏭] Já existiam: 73, [✗] Erros: 0
```

✅ No duplicates are created on subsequent runs!

---

## Error Troubleshooting

### "Connection refused" / "no such host"
```
❌ Error: dial tcp localhost:8080: connect: connection refused
```
**Solution:** Start LEP-Back first
```bash
# Terminal 1
cd LEP-Back
go run main.go
# Wait for "Server running on :8080"

# Terminal 2
cd LEP-Script/LEP-execute-seed
go run .
```

### "invalid email" / Status 400
```
❌ [user] joao@fattoria.com.br: status 400: invalid email
```
**Solution:** Check backend email validation rules. May need to update seed-fattoria.json emails.

### "already exists" / Status 409
```
⏭ [product] Spaghetti já existe
```
**Solution:** Normal! Seeder detected existing item and skipped it (idempotency working).

### "not found" / Status 404
```
❌ [reservation] FAT-20251120-001: status 404: customer not found
```
**Solution:** Make sure Passo 10 (Create Customers) completed successfully before Passo 12 (Create Reservations).

---

## Configuration File (config.yaml)

Located in project directory. Customize if needed:

```yaml
server:
  url: http://localhost:8080      # Change if backend on different port/host
  timeout: 30                      # Increase if running slow

auth:
  organization_name: "LEP Fattoria"  # Your restaurant name
  fallback_email: "admin@lep-fattoria.com"
  fallback_password: "password"
  auto_email: true

seed:
  file: seed-fattoria.json         # Default file to seed
  stop_on_error: false             # Continue on errors
  parallel: false                  # Sequential execution

logging:
  level: debug                     # info or debug
  show_payloads: true              # Show request/response bodies
```

---

## Performance Notes

### Typical Execution Times
- **Single seed file:** 2-5 seconds
- **Both seed files:** 4-10 seconds
- **With -verbose flag:** +50% time (extra logging overhead)

### What Affects Speed
- Backend response time
- Network latency
- Database performance
- Number of items in seed file

### Optimization Tips
```yaml
# For faster execution in development:
logging:
  level: info          # Less logging = faster
  show_payloads: false # No JSON parsing = faster

# For production/stable environments:
seed:
  parallel: true       # Parallel execution (if backend supports)
```

---

## Success Checklist

After running `go run .`, verify:

- [x] No "Error" messages in output
- [x] Final summary shows "[✓] Total Criados: 146" (or similar)
- [x] No "[✗] Total Erros: X" > 0
- [x] All 12 steps completed ("Passo 1" through "Passo 12")
- [x] Time taken is reasonable (< 30 seconds)

---

## Next Steps

### 1. Verify Data in Database
```bash
# Connect to your database
psql -U your_user -d lep_database

# Check created tables
SELECT COUNT(*) FROM "users";        -- Should be >= 4
SELECT COUNT(*) FROM "customers";    -- Should be >= 5
SELECT COUNT(*) FROM "reservations"; -- Should be >= 4
SELECT COUNT(*) FROM "tags";         -- Should be >= 4
```

### 2. Test via API
```bash
# Get users
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "X-Lpe-Organization-Id: YOUR_ORG_ID" \
     http://localhost:8080/user

# Get customers
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "X-Lpe-Organization-Id: YOUR_ORG_ID" \
     http://localhost:8080/customer

# Get reservations
curl -H "Authorization: Bearer YOUR_TOKEN" \
     -H "X-Lpe-Organization-Id: YOUR_ORG_ID" \
     http://localhost:8080/reservation
```

### 3. Test via Frontend
1. Start frontend: `npm run dev`
2. Login with created user: `joao@fattoria.com.br` / `senha123`
3. Navigate to Users, Customers, Reservations sections
4. Verify data is displayed correctly

---

## Files Reference

| File | Purpose | Size |
|------|---------|------|
| `main.go` | Seed execution orchestration | 587 lines |
| `client_v2.go` | API communication & CRUD | 815 lines |
| `config.go` | Configuration loading | 140 lines |
| `seed_data.go` | Go struct definitions | 270 lines |
| `seed-fattoria.json` | Test data (Fattoria restaurant) | 890 lines |
| `seed-data.json` | Test data (La Bella Italia) | 623 lines |
| `config.yaml` | Configuration file (user-editable) | 19 lines |
| `logger.go` | Structured logging | 40 lines |

---

## Tips & Tricks

### Rerun Seeder After Database Reset
```bash
# If you reset your database and rerun seeder:
go run .
# It will recreate all entities (idempotency still works within same run)
```

### Create Multiple Organizations
```bash
# Modify config.yaml to test different organizations:
# First: organization_name: "Fattoria"
# go run . -file seed-fattoria.json

# Then: organization_name: "La Bella Italia"
# go run . -file seed-data.json
# Creates separate multi-tenant data!
```

### Debug Specific Steps
```bash
# Set logging to debug and run
go run . -verbose
# Look for "[DEBUG]" messages showing request/response flow
```

### Check What Would Be Created
```bash
# Run once:
go run .
# Summary shows what was created

# To see what WOULD be created (without actually seeding):
# Edit seed-fattoria.json, remove users/customers/reservations sections
# Run again to see only org/menus/products
```

---

## Support

For issues:
1. Check error messages in output
2. Verify backend is running: `curl http://localhost:8080/ping`
3. Check database connection
4. Run with `-verbose` flag for detailed logs
5. Check `PHASE2_COMPLETE.md` for detailed architecture

---

**Happy Seeding! 🌱**
