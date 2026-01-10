# 🧪 LEP Backend Test Suite

Complete automated test suite for LEP system with **183+ tests** across **8 test files**.

## 📊 What This Does

Tests the complete backend API for:
- ✅ Authentication & Authorization (JWT tokens, permissions)
- ✅ Multi-tenant isolation (org/project validation)
- ✅ CRUD operations (users, products, orders, reservations, etc.)
- ✅ Settings & Theme customization
- ✅ Notifications & Webhooks (Twilio integration)
- ✅ Reports & Data export
- ✅ Advanced features (menu selection, filtering, relationships)

## 🚀 Quick Start

### Prerequisites
- Backend running on `http://localhost:8080`
- PostgreSQL configured
- Master Admin permissions enabled for test user

### Run Tests
```bash
go run . -verbose
```

**Expected Result**: ~200+/205 tests passing (97%+) in 2-3 minutes

## 📁 Test Files

| File | Tests | Purpose |
|------|-------|---------|
| `tests.go` | ~70 | Main orchestrator + Phases 1-10 |
| `tests_sprint1_critical.go` | 34 | Critical features (auth, webhooks, notifications) |
| `tests_sprint2_high.go` | 13 | Settings, theme, menu management |
| `tests_sprint3_medium.go` | 38 | Advanced filters, categories, tags |
| `tests_upload_fix.go` | 5 | Image upload tests |
| `tests_product_tags_optimization.go` | 3 | Product tags optimization |
| `tests_menu_intelligent_selection.go` | 5 | Intelligent menu selection |
| `tests_theme_customization.go` | 15 | Theme customization |

## 🎯 Test Coverage

- **Routes Tested**: 115+/153 (75%+)
- **Total Tests**: 183+
- **Success Rate**: 97%+

## 🔧 Architecture

```
main.go
  ├── config.go (Backend URL, test credentials)
  ├── client.go (HTTP client, request handling)
  ├── logger.go (Logging utilities)
  └── tests.go (Test orchestrator)
       ├── tests_sprint1_critical.go
       ├── tests_sprint2_high.go
       └── tests_sprint3_medium.go
```

## 📋 Configuration

Edit `config.go` to change:
- **Backend URL**: `BackendURL` (default: `http://localhost:8080`)
- **Test User**: Email and password for login
- **Multi-tenant Headers**: Organization and Project IDs

## 🛠️ Troubleshooting

### Tests failing with 401 errors
- Verify backend is running
- Check JWT token in config matches backend JWT_SECRET

### Tests failing with 403 errors
- Verify Master Admin permission is enabled: `go run cmd/create-master-admins/main.go` (in LEP-Back)

### Tests failing with connection errors
- Verify backend URL in `config.go`
- Ensure PostgreSQL is running

## 📈 Performance

- **Build Time**: <1 second
- **Execution Time**: 2-3 minutes for all tests
- **Memory**: ~100MB
- **Network**: Minimal (local loopback)

## ✅ Recent Updates

See `LATEST_UPDATES.txt` for the most recent changes and improvements.

---

**Status**: ✅ Ready for execution
**Version**: 1.0
**Last Updated**: 2025-11-08
