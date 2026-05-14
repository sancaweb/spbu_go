# GitHub Copilot Instructions — spbu_go

## Big Picture

- Modular monolith in Go: Gin + GORM + PostgreSQL, SSR templates, Alpine.js, Tailwind, jQuery DataTables.
- Wiring is manual in [cmd/main.go](../cmd/main.go): repository -> service -> handler -> route registration.
- Architecture by layer:
  - [internal/entity](../internal/entity) (GORM models + `TableName()`)
  - [internal/repository](../internal/repository) (DB access; interface + impl in same file)
  - [internal/service](../internal/service) (business logic)
  - [internal/handler](../internal/handler) (Gin HTTP + template/API responses)

## Run / Build / Validate

- Run app: `go run cmd/main.go`
- Build check: `go build ./...`
- Test sweep (if/when tests exist): `go test ./...`
- Config is loaded from `config/.env` (not root). DB driver is Postgres (`pkg/database/database.go`).

## Database & Startup Rules

- Startup executes manual SQL migration blocks in [cmd/main.go](../cmd/main.go), then `AutoMigrate`.
- SQL files under [migrations](../migrations) are not auto-executed.
- `seeders.Seed()` runs every startup and is idempotent (`FirstOrCreate` patterns in [seeders/seeder.go](../seeders/seeder.go)).
- Keep [database/schema.dbml](../database/schema.dbml) updated when schema changes.

## Conventions That Matter

- Most entities include `UpdatedBy`, `Updater`, `gorm.DeletedAt` (see [internal/entity/partner_entity.go](../internal/entity/partner_entity.go)).
- Always omit relation fields on save/create to avoid accidental nested inserts (example in [internal/repository/partner_repo.go](../internal/repository/partner_repo.go): `Omit("Updater")`).
- Archive/restore patterns use soft delete + `.Unscoped()` queries (see `FindInactive()`/`Restore()` in partner repo).
- Mutating endpoints use POST even for update/delete (see route definitions in [cmd/main.go](../cmd/main.go)).
- DataTables endpoints use [internal/dto/datatable.go](../internal/dto/datatable.go) and PostgreSQL `ILIKE` for search.

## HTTP / Session / Middleware Flow

- Session-based auth via `gin-contrib/sessions` cookie store; no JWT.
- `AuthRequired` loads full user from session and puts it in context as `user`.
- `SettingsMiddleware` injects `favicon`; handlers rendering HTML should pass both `User` and `Favicon`.

## Frontend Patterns

- Template names are paths relative to `templates/` (for example `"partner/index.html"`, `"transaction/penebusan/index.html"`).
- Templates are loaded via explicit Windows-safe glob patterns in [internal/server/router.go](../internal/server/router.go).
- Alpine components are registered in `document.addEventListener("alpine:init", ...)`.
- Server JSON is embedded using `<script type="application/json">` + `template.JS(...)` (see penebusan handler/template).
- Indonesian numeric formatting/parsing helpers live in [templates/includes/footer.html](../templates/includes/footer.html): `formatIDR`, `formatStock`, `parseIDR`, `formatInputIDR`, `formatInputStock`.

## Service/Repository Interface Placement

- Shared service interfaces are in [internal/service/interfaces.go](../internal/service/interfaces.go).
- Module-specific interfaces are declared in their module files (example: [internal/service/partner_service.go](../internal/service/partner_service.go), [internal/repository/penebusan_repo.go](../internal/repository/penebusan_repo.go)).

## Domain Notes

- Core modules currently wired: Users/Roles/Permissions, Settings, BBM/Tiang/Nozzle, Partner, Employee/Jabatan/Pendapatan/Potongan, Wallet, COA/COA Mapping, Penebusan.
- Useful business docs live in [.gemini/docs](../.gemini/docs) (menu, schema, per-module behavior).

## Implementation Playbook

- For consistent future feature work, follow [.github/feature-implementation-playbook.md](feature-implementation-playbook.md).
- Use quick templates by feature type in [.github/feature-templates](feature-templates):
  - [master-crud-template.md](feature-templates/master-crud-template.md)
  - [transaction-template.md](feature-templates/transaction-template.md)
  - [settings-kv-template.md](feature-templates/settings-kv-template.md)
