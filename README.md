# Admin Platform

Independent permission-management platform based on `github.com/ngq/gorp` and Vben Admin.

## Scope

- JWT authentication
- Users and departments
- Roles, menus, buttons, and API permissions
- Casbin policy synchronization
- System parameters and dashboard
- Operation, login, and error logs

## Backend

The backend uses gorp Contract HTTP mode and MySQL.

```powershell
cd backend
Copy-Item .env.example .env
go run ./cmd/app
```

The local default administrator created by the development seed is `admin / admin123`.
Change this password after the first login.

## Frontend

```powershell
cd web
pnpm install --frozen-lockfile
pnpm --filter @vben/web-antd dev
```

The frontend development server proxies `/api` to `http://localhost:8080`.
