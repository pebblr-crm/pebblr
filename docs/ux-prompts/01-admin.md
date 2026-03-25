# Admin Persona

The admin manages the system configuration, users, and has full visibility into audit and telemetry.

## What the Admin Sees

- **Configuration:** Tenant settings, activity types, status workflows, field definitions, business rules (visit cadence, max activities/day, frequency targets)
- **Audit logs:** Full history of all changes across the system (who changed what, when, old/new values)
- **Telemetry:** System-wide business metrics and usage data
- **User management:** Create/edit users, assign roles (rep, manager, admin)
- **Teams:** Create teams, assign managers, manage team membership
- **Roles:** Control what each role can see and do

## Key Flows

### User & Team Setup
1. Create teams (e.g., by geographic sector)
2. Assign a manager to each team
3. Add reps to teams
4. Roles determine data visibility (reps see own targets, managers see team, admin sees all)

### Configuration
- Define target types (e.g., doctor, pharmacy) and their fields (specialty, potential, address, etc.)
- Define activity types (visit, vacation, training, etc.) and which ones block field activities
- Set status workflows and transitions (planned → completed → submitted)
- Set business rules: frequency targets per classification (A/B/C), visit cadence, max activities per day
- Configure recovery day rules (weekend work → earned recovery, claim window)

### Audit & Compliance
- View audit trail filtered by entity, user, date range
- Track all status changes, submissions, and edits across the system

## UX Notes

- Admin UX is secondary priority — functional and clear is sufficient
- Configuration changes affect all users, so confirmation and preview of impact is important
- Audit views need strong filtering and search — the volume of entries will be high
