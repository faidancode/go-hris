-- Rollback demo employee role assignment from 000017.
-- Keep role and permission master data intact to avoid removing shared references.

DELETE FROM employee_roles er
USING employees e, roles r, companies c
WHERE er.employee_id = e.id
  AND er.role_id = r.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email = 'owner@demo.com'
  AND r.name = 'Owner';

DELETE FROM employee_roles er
USING employees e, roles r, companies c
WHERE er.employee_id = e.id
  AND er.role_id = r.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email = 'hr@demo.com'
  AND r.name = 'HR';

DELETE FROM employee_roles er
USING employees e, roles r, companies c
WHERE er.employee_id = e.id
  AND er.role_id = r.id
  AND e.company_id = c.id
  AND c.registration_number = 'DEMO-001'
  AND e.email = 'employee@demo.com'
  AND r.name = 'Employee';
