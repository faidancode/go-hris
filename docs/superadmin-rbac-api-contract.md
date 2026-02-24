# SUPERADMIN RBAC API Contract

Dokumen ini fokus untuk kebutuhan frontend Next.js pada menu:
- Role Management
- Permission Management
- Assign Role ke User/Employee

Base URL: `/api/v1`  
Auth: `Authorization: Bearer <access_token>`

## 1) List Permission Catalog

Endpoint:
- `GET /rbac/permissions`

RBAC middleware:
- wajib punya `role:manage`

Success response (`200`):
```json
{
  "ok": true,
  "data": [
    {
      "id": "0f18f2f1-24b8-4f4e-a5ce-b7b1df2dd876",
      "resource": "user",
      "action": "read",
      "label": "Melihat User",
      "category": "User Management"
    }
  ],
  "meta": null,
  "error": null
}
```

## 2) List Roles (by company context from JWT)

Endpoint:
- `GET /rbac/roles`

RBAC middleware:
- wajib punya `role:read`

Success response (`200`):
```json
{
  "ok": true,
  "data": [
    {
      "id": "f11eb2f9-95ab-4380-8675-ad9a5f4d69b4",
      "name": "HR",
      "description": "Manage employees and attendance",
      "permissions": [
        "employee:read",
        "employee:create",
        "user:read"
      ]
    }
  ],
  "meta": null,
  "error": null
}
```

## 3) Get Role Detail

Endpoint:
- `GET /rbac/roles/:id`

RBAC middleware:
- wajib punya `role:read`

Success response (`200`):
```json
{
  "ok": true,
  "data": {
    "id": "f11eb2f9-95ab-4380-8675-ad9a5f4d69b4",
    "name": "HR",
    "description": "Manage employees and attendance",
    "permissions": [
      "employee:read",
      "employee:create",
      "user:read"
    ]
  },
  "meta": null,
  "error": null
}
```

## 4) Create Role

Endpoint:
- `POST /rbac/roles`

RBAC middleware:
- wajib punya `role:manage`

Request body:
```json
{
  "name": "UserAdmin",
  "description": "Manage users only",
  "permissions": [
    "user:read",
    "user:create",
    "user:update"
  ]
}
```

Success response (`201`):
```json
{
  "ok": true,
  "data": null,
  "meta": null,
  "error": null
}
```

## 5) Update Role (including permission mapping)

Endpoint:
- `PUT /rbac/roles/:id`

RBAC middleware:
- wajib punya `role:manage`

Request body:
```json
{
  "name": "UserAdmin",
  "description": "Manage users and role read",
  "permissions": [
    "user:read",
    "user:create",
    "user:update",
    "role:read"
  ]
}
```

Success response (`200`):
```json
{
  "ok": true,
  "data": null,
  "meta": null,
  "error": null
}
```

## 6) Delete Role

Endpoint:
- `DELETE /rbac/roles/:id`

RBAC middleware:
- wajib punya `role:manage`

Success response (`200`):
```json
{
  "ok": true,
  "data": null,
  "meta": null,
  "error": null
}
```

## 7) List User + Roles (for User Management table)

Endpoint:
- `GET /users/with-roles`

RBAC middleware:
- wajib punya `user:read`

Success response (`200`):
```json
[
  {
    "id": "8ab4897f-a159-4a3e-81c2-2b42fd65f6f1",
    "employee_id": "f690292e-c490-4369-b8b5-c5b8473cdf76",
    "email": "hr@demo.local",
    "full_name": "Demo HR",
    "is_active": true,
    "roles": [
      "HR"
    ],
    "created_at": "2026-02-20 10:30:00"
  }
]
```

Catatan:
- Endpoint ini saat ini mengembalikan array langsung (tanpa envelope `ok/data`).

## 8) Assign Role ke User (mapping ke employee_roles)

Endpoint:
- `PATCH /users/:id/role`

RBAC middleware:
- wajib punya `role:manage`

Request body:
```json
{
  "role_name": "HR"
}
```

Success response (`200`):
- empty body

Catatan:
- `:id` adalah `users.id`
- backend akan lookup `employee_id` dari user tersebut, lalu insert ke `employee_roles`.
- format saat ini bersifat add-only (`ON CONFLICT DO NOTHING`), jadi bisa multi-role.

## Error Envelope (RBAC endpoints)

Untuk endpoint di modul `/rbac`, error format:
```json
{
  "ok": false,
  "data": null,
  "meta": null,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Input tidak valid",
    "details": "..."
  }
}
```

## Minimal Frontend Flow (SUPERADMIN)

1. `GET /rbac/permissions` untuk render grouped checkbox.
2. `GET /rbac/roles` untuk list role.
3. `POST /rbac/roles` atau `PUT /rbac/roles/:id` untuk simpan role+permission.
4. `GET /users/with-roles` untuk tabel user.
5. `PATCH /users/:id/role` untuk assign role ke user/employee.
