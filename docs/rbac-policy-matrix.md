# RBAC Policy Matrix (Recommended)

## Roles
- `SUPERADMIN` (bootstrap/dev only)
- `Owner`
- `HR`
- `Finance`
- `Employee`

## Action Legend
- `R` = read
- `C` = create
- `U` = update
- `D` = delete
- `A` = approve
- `P` = pay
- `M` = manage
- `X` = cancel

## Matrix

| Module | SUPERADMIN | Owner | HR | Finance | Employee |
|---|---|---|---|---|---|
| `employee` | R,C,U,D | R,C,U,D | R,C,U,D | R | R (self only) |
| `department` | R,C,U,D | R,C,U,D | R,C,U,D | R | - |
| `position` | R,C,U,D | R,C,U,D | R,C,U,D | R | - |
| `salary` | R,U,D* | R,U,D* | R,U | R,U | R (self only) |
| `payroll` | R,C,U,D,A,P,X | R,C,U,D,A,P,X | R,C,U,D,X | R,A,P | R (self only) |
| `leave` | R,C,U,D,A,M,X | R,C,U,D,A,M,X | R,C,U,D,A,M,X | R | R,C,X (self only) |
| `role` | R,M | R,M | R | - | - |
| `company` | R,U | R,U | R | R | - |
| `attendance` | R,M | R,M | R,M | R | R (self only) |

Notes:
- `D*` pada salary direkomendasikan hanya untuk record draft/koreksi; idealnya gunakan versioning, bukan hard delete.
- Untuk modul operasional (`employee`, `leave`, `payroll`), lebih aman gunakan soft-delete/status transition daripada hard delete.
- `SUPERADMIN` sebaiknya hanya untuk bootstrap environment development.

## Delete Policy (Recommended)

| Module | Delete Allowed | Rule |
|---|---|---|
| `employee` | Yes (soft) | Gunakan soft delete/termination, bukan hard delete permanen |
| `department` | Conditional | Tidak boleh jika masih dipakai employee/position aktif |
| `position` | Conditional | Tidak boleh jika masih dipakai employee aktif |
| `salary` | Limited | Hindari delete; gunakan koreksi/versioning |
| `payroll` | Limited | Hanya draft/belum approved/paid |
| `leave` | Limited | Prefer `cancel` daripada delete |
| `role` | Conditional | Tidak boleh hapus system role / role yang masih dipakai |
| `permission` | No | Master data, tidak direkomendasikan delete |
| `company` | No (operational) | Pakai deactivate/soft delete |

## Mapping To Existing Permission Keys

Gunakan key permission yang sudah ada di sistem:
- `employee`: `read`, `create`, `update`, `delete`
- `department`: `read`, `create`, `update`, `delete`
- `position`: `read`, `create`, `update`, `delete`
- `salary`: `read`, `update`
- `payroll`: `read`, `create`, `approve`, `pay`, `delete`
- `leave`: `read`, `create`, `approve`, `manage`
- `role`: `read`, `manage`
- `company`: `read`, `update`
- `attendance`: `read`, `manage`

Jika ingin mendukung `cancel` secara eksplisit, tambahkan action permission baru:
- `leave:cancel`
- `payroll:cancel`
