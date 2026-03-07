You are a senior Go backend architect.

I want you to generate a production-ready starter project blueprint.

PROJECT GOAL:
Build a modular Monolith User Management System for internal SPBU system.

MAIN FEATURES:
- Authentication (Login / Logout)
- User CRUD
- Role CRUD
- Permission CRUD
- Role has many permissions
- User has one role
- Middleware-based authorization
- Session-based login (not JWT)

TECH STACK:
- Go
- Gin Framework
- GORM
- golang-migrate for migrations
- Seeder support
- MySQL
- Use DECIMAL type for all money and liter related fields
- Use server-side rendering (html/template)
- Bootstrap 5 admin layout (no SPA)

ARCHITECTURE REQUIREMENTS:
- Clean Architecture style (handler, service, repository separation)
- Modular structure so new modules can be added easily
- No global variables
- Dependency injection pattern
- Config file support (.env)
- Use bcrypt for password hashing
- Use middleware for authentication & authorization

PROJECT STRUCTURE SHOULD INCLUDE:
- cmd/
- internal/
    - user/
    - role/
    - permission/
    - auth/
- migrations/
- seeders/
- templates/
- static/
- config/

DATABASE REQUIREMENTS:
- users table
- roles table
- permissions table
- role_permissions pivot table
- users.role_id foreign key
- Proper indexing
- Soft delete support

I want you to generate:
1. Folder structure
2. Database schema (migration files)
3. Seeder examples
4. Model struct definitions
5. Repository layer example
6. Service layer example
7. Handler layer example
8. Middleware example
9. Login implementation
10. Example HTML template structure with layout system

Make the code clean, well-commented, scalable, and easy to extend.
Explain each architectural decision briefly.

Do not simplify.
Think like building a real scalable system.

Follow best practices used in enterprise Go applications.
Avoid anti-patterns.
Use interface abstraction for repository layer.
