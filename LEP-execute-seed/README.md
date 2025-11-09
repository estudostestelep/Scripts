# 🌱 LEP Database Seeder

Automated Go tool to populate LEP database with initial data (organizations, users, menus, categories, products, tables, environments, etc.) via HTTP API.

## 📊 What This Seeder Does

- ✅ **Organization Setup** - Creates organization, project, and admin user
- ✅ **User Management** - Creates admin user with master admin permissions
- ✅ **Menu Management** - Creates menu structures
- ✅ **Categories** - Creates product categories and subcategories
- ✅ **Environments** - Creates physical dining areas
- ✅ **Tables** - Creates table configurations with capacity and locations
- ✅ **Products** - Creates menu items with pricing (normal, glass, bottle)
- ✅ **Idempotent** - Safe to run multiple times (skips existing data)
- ✅ **API-Based** - Works with remote backends (no database access required)
- ✅ **Customizable** - Easy to modify seed data in JSON files

## 🛠️ Technology Stack

- **Go 1.18+** - Programming language
- **HTTP Client** - RESTful API communication
- **JSON** - Data format for seed configuration
- **UUID** - Unique identifier generation

## 🚀 Quick Start

### Prerequisites

- **Go 1.18+**
- **Backend running on http://localhost:8080**
- **PostgreSQL configured** (for backend)

### Installation & Running

```bash
# Navigate to seeder directory
cd LEP-Script/LEP-execute-seed

# Run seeder (default: localhost:8080, seed-fattoria.json)
go run .

# Run with custom URL
go run . -url http://your-api.com:8080

# Run with custom seed file
go run . -file seed-custom.json

# Run with verbose logging
go run . -verbose

# Combine multiple options
go run . -url http://api.example.com -file seed-data.json -verbose
```

### Build Binary

```bash
# Compile to executable
go build -o lep-seeder.exe

# Run executable
./lep-seeder.exe
# or with parameters
./lep-seeder.exe -url http://localhost:8080 -verbose
```

## 🎨 Logging Symbols

The seeder uses colorized symbols to indicate different types of messages:

| Symbol | Color | Meaning | Example |
|--------|-------|---------|---------|
| `[ℹ]` | Blue | Information message | `[ℹ] Carregando seed-fattoria.json...` |
| `[✓]` | Green | Success - item created or operation completed | `[✓] Organização OK (ID: xxx)` |
| `[⏭]` | Cyan | Skip - item already exists | `[⏭] Menu Menu Principal já existe` |
| `[✗]` | Red | Error - operation failed | `[✗] Erro ao criar menu: status 403` |
| `[⚠]` | Yellow | Warning - something to pay attention to | `[⚠] Token expirado, reconectando...` |
| `[D]` | Yellow | Debug - detailed technical information (verbose mode only) | `[D] [/menu] Status: 201, Body: {...}` |

### Log Levels

- **Info (default)**: Shows [ℹ], [✓], [⏭], [✗], [⚠] messages
- **Debug (verbose mode)**: Also shows [D] messages with full HTTP requests/responses

### Example Output Interpretation

```
[ℹ] Carregando seed-fattoria.json...           ← Info: loading file
[ℹ] Arquivo carregado com 66 items             ← Info: 66 items loaded
[✓] Organização OK (ID: c5d9...)               ← Success: organization created
[⏭] Menu Menu Principal já existe              ← Skip: menu already exists
[✗] Erro ao criar menu: status 403             ← Error: failed to create menu
[D] [/menu] Status: 403, Body: {...}           ← Debug: HTTP response details
```

## 📋 Parameters

| Parameter | Default | Description |
|-----------|---------|-------------|
| `-url` | `http://localhost:8080` | Backend API base URL |
| `-file` | `seed-fattoria.json` | JSON file with seed data |
| `-verbose` | `false` | Enable detailed logging (shows [D] debug messages) |

## 📁 Project Structure

```
LEP-execute-seed/
├── main.go              # Entry point - orchestrates seeding
├── client.go            # HTTP client for API communication
├── logger.go            # Colorized logging
├── seed_data.go         # Data structures and types
├── seed-fattoria.json   # Fattoria restaurant seed data
├── go.mod               # Go dependencies
└── README.md            # This file
```

## 📝 Seed Data Format

The `seed-fattoria.json` file defines all data to be created:

```json
{
  "organization": {
    "name": "Fattoria Pizzeria",
    "email": "admin@fattoria.com",
    "phone": "+55 11 99999-9999",
    "address": "Rua Principal, 123",
    "website": "https://fattoria.com",
    "description": "Italian restaurant"
  },
  "menus": [
    {
      "name": "Main Menu",
      "description": "Primary menu",
      "active": true,
      "order": 1
    }
  ],
  "categories": [
    {
      "name": "Beverages",
      "description": "Drinks section",
      "menu_id_ref": 0,
      "active": true,
      "order": 1
    }
  ],
  "subcategories": [
    {
      "name": "Red Wines",
      "description": "Red wine selection",
      "category_id_ref": 0,
      "active": true,
      "order": 1
    }
  ],
  "environments": [
    {
      "name": "Main Dining Area",
      "capacity": 50,
      "active": true
    }
  ],
  "tables": [
    {
      "number": 1,
      "capacity": 4,
      "location": "Window seat",
      "status": "available",
      "environment_id_ref": 0
    }
  ],
  "products": [
    {
      "name": "Red Wine",
      "description": "Premium Italian red wine",
      "type": "wine",
      "price_normal": 150.00,
      "price_glass": 30.00,
      "price_bottle": 150.00,
      "category_id_ref": 0,
      "vintage": "2020",
      "country": "Italy",
      "winery": "Antinori",
      "active": true
    }
  ]
}
```

### Reference Fields

When referencing previously created entities, use index-based references:

- `menu_id_ref` - Index (0-based) in `menus` array
- `category_id_ref` - Index in `categories` array
- `subcategory_id_ref` - Index in `subcategories` array
- `environment_id_ref` - Index in `environments` array

## 🔄 Execution Flow

The seeder executes in the following order:

1. **Step 1** - Create Organization + Project + Admin User
2. **Step 2** - Login admin user and obtain JWT token
3. **Step 3** - Create menus
4. **Step 4** - Create categories
5. **Step 5** - Create subcategories
6. **Step 6** - Create environments
7. **Step 7** - Create tables
8. **Step 8** - Create products
9. **Step 9** - Create users (waiter, manager, kitchen staff, etc.)
10. **Step 10** - Create customers (with contact info and preferences)
11. **Step 11** - Create tags (dietary restrictions, special attributes)
12. **Step 12** - Create reservations (bookings with customer and table references)

## 🛡️ Idempotency

The seeder is **idempotent** and safe to run multiple times:

- ✅ Skips existing entities without error
- ✅ Creates missing entities
- ✅ Reports all operations at the end
- ✅ Returns exit code 1 if errors occurred

### Example Output

```
========== 🌱 LEP Database Seeder ==========
[ℹ] URL: http://localhost:8080
[ℹ] File: seed-fattoria.json
[ℹ] Verbose: false

========== Step 1: Creating Organization ==========
>>> Creating organization: Fattoria Pizzeria
[✓] Organization created: Fattoria Pizzeria
[✓] Project created: Fattoria
[✓] Admin user created: admin@fattoria.com

========== Step 2: Authenticating Admin ==========
>>> Logging in: admin@fattoria.com
[✓] Successfully authenticated

========== Step 3: Creating Menus ==========
[✓] Menu created: Main Menu

========== Execution Summary ==========
[✓] Created: 15
[⏭] Already existed: 5
[✗] Errors: 0
========== Completed: 16:45:23 ==========
```

## 🧪 Usage Examples

### Basic Seeding

```bash
# Seed localhost with Fattoria data
go run .
```

### Custom Environment

```bash
# Seed staging environment
go run . -url https://api-staging.example.com -verbose

# Seed production environment
go run . -url https://api.example.com -file seed-production.json
```

### Development & Testing

```bash
# Test with verbose output
go run . -url http://localhost:8080 -verbose

# Create custom seed file for testing
go run . -file seed-test.json -verbose
```

## 🔧 Troubleshooting

### "undefined: APIClient"
**Solution**: Run `go run .` instead of `go run main.go` to compile all files

### "connection refused" on API
**Solution**: Verify backend is running:
```bash
curl http://localhost:8080/ping
# Should return: "pong"
```

### "Credenciais inválidas" (401 Invalid Credentials)
**Solution**: Check that the fallback credentials in `config.yaml` match the backend:
```yaml
auth:
  fallback_email: "pablo@lep.com"
  fallback_password: "senha123"
```

Verify the admin user exists in your backend database. If needed, manually create the user or update the credentials.

### "Access denied: token inválido: signature is invalid" (403)
**Solution**: JWT signature validation failure - the backend cannot verify the token signature. This occurs when:

1. **JWT keys don't match**: Backend's `JWT_SECRET_PRIVATE_KEY` and `JWT_SECRET_PUBLIC_KEY` are not properly configured
2. **Keys changed between login and API calls**: The keys were rotated without restarting the backend

**Fix**:
```bash
# Check backend JWT configuration
echo $JWT_SECRET_PRIVATE_KEY
echo $JWT_SECRET_PUBLIC_KEY

# Ensure keys are set and restart backend
# Backend: cd LEP-Back && go run main.go

# Then re-run seeder
cd LEP-Script/LEP-execute-seed && go run .
```

### "database connection error"
**Solution**: Check backend environment variables:
```bash
# Verify database is accessible
PGPASSWORD=your_password psql -h localhost -U lep_user -d lep_database -c "SELECT 1"
```

### Data already exists
**Solution**: Seeder is idempotent, it will skip existing data:
```bash
# Safe to run multiple times
go run .
go run .  # Second run skips existing data
```

To recreate data, clear the database:
```bash
# Warning: This deletes all data!
# Option 1: Drop and recreate database (PostgreSQL)
psql -U postgres -c "DROP DATABASE lep_database; CREATE DATABASE lep_database;"

# Option 2: Manually delete organization from backend
# Use backend admin panel to delete "LEP Fattoria" organization
```

### "Já existe uma organização com esse nome" (400 Organization already exists)
**Solution**: The organization already exists in the database. The seeder will:

1. Detect the existing organization (400 error)
2. Try to login with the creation email/password
3. Fall back to the `fallback_email`/`fallback_password` from config.yaml

Ensure the fallback credentials match the admin user in your database.

## ⚙️ Backend Configuration

The LEP backend needs proper environment variables in `.env`:

```bash
# Database
DB_USER=lep_user
DB_PASS=lep_password
DB_NAME=lep_database
INSTANCE_UNIX_SOCKET=/path/to/socket  # For GCP Cloud SQL

# Authentication
JWT_SECRET_PRIVATE_KEY=your_private_key
JWT_SECRET_PUBLIC_KEY=your_public_key

# Optional
ENABLE_CRON_JOBS=true
GIN_MODE=debug
```

## 📊 Default Seed Data (Fattoria)

The `seed-fattoria.json` creates a complete restaurant management system with:

### Infrastructure Setup
- **Organization**: LEP Fattoria (Italian cuisine restaurant)
- **Project**: Fattoria project
- **Admin User**: pablo@lep.com (for initial setup)

### Menu & Product Catalog
- **Menus**: 3 menus (Principal, Almoço, Noite)
- **Categories**: 6 product categories
- **Subcategories**: 12 subcategories for fine-grained organization
- **Products**: 33 menu items (appetizers, pastas, meats, seafood, desserts, wines, beverages)
- **Tags**: 4 product tags (Vegetariano, Sem Glúten, Especial da Casa, Picanço)

### Physical Layout
- **Environments**: 3 dining areas (Salão, Varanda, Sala Privativa)
- **Tables**: 9 tables with different capacities

### Staff & Customers
- **Users**: 4 staff members (Admin, Manager, Waiter, Kitchen)
- **Customers**: 5 customer profiles with contact info and preferences

### Operations
- **Reservations**: 4 sample reservations with confirmation keys
- **Settings**: Configuration for reservation window, email notifications, theme customization
- **Notification Templates**: 3 templates for confirmations, reminders, cancellations

## 🧑‍💻 Development

To customize seed data:

1. Edit `seed-fattoria.json` with your data
2. Test with `go run . -verbose` to see detailed logs
3. Don't modify Go files unless updating data structures
4. Run `go mod tidy` if adding dependencies

## 📚 Dependencies

- **`github.com/google/uuid`** - UUID generation and parsing

Install via:
```bash
go mod tidy
```

## 📊 Build Status

- ✅ Builds successfully (0 errors)
- ✅ Runs without errors
- ✅ Handles all edge cases
- ✅ Idempotent and safe to run multiple times
- ✅ Production-ready

---

**Version**: 1.0
**Status**: ✅ Production Ready
**Last Updated**: 2025-11-09
