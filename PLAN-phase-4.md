# Phase 4 — Post Go-Live Optimizations ❌

> Back to [overview](PLAN.md)

## Checklist

24. ❌ Weekend activity + recovery days
25. ❌ Target group management (quarterly)
26. ❌ Plan generation (rule-based monthly plan proposal)
27. ❌ i18n / Romanian UI
28. ❌ Data migration from Twenty CRM

### Removed / Superseded

The following items from the original Phase 4 have been absorbed or superseded by the map planner:

- ~~Drag & drop calendar~~ — the map planner provides drag-and-drop target-to-day assignment. The week/month views remain read-only calendars; rescheduling individual activities can be done via edit.
- ~~Copy-paste activities~~ — replaced by **Clone week** (Phase 3, item 21), which operates at the right abstraction level for the 3-week rotation cycle.
- ~~Advanced filtering with saved filters~~ — replaced by **Target collections** (Phase 3, item 20). Reps don't need to filter activities — they need to quickly re-select groups of targets for planning. Collections solve this at the source.

---

## 1. Weekend Activity + Recovery Days ❌

Config rules:

```json
"recovery": {
  "weekend_activity_flag": true,
  "recovery_window_days": 5,
  "recovery_type": "full_day"
}
```

- If a rep works on a weekend (creates a field activity on Saturday/Sunday), they earn a recovery day
- Recovery day must be taken within `recovery_window_days` business days
- Recovery type (full_day or half_day) is configurable
- Backend enforces: track weekend activities, validate recovery day claims
- Frontend: show recovery day balance in planner sidebar

---

## 2. Target Group Management (Quarterly) ❌

Each quarter, managers define which targets each rep should focus on:

- Admin/manager UI to assign targets to reps for the quarter
- Bulk assignment (CSV upload or multi-select)
- Track changes: which targets were added/removed from target group
- Frequency rules apply to target group specifically
- Dashboard shows target group coverage vs all-target coverage

Note: This is distinct from **Target collections** (Phase 3, item 20). Collections are user-created groupings for planning convenience. Target group management is manager-driven assignment of which targets a rep is responsible for in a given quarter.

---

## 3. Plan Generation ❌

Auto-propose a monthly activity plan based on rules:

- Input: rep's assigned targets, frequency rules, blocked days (holidays, vacations), routing preferences, saved collections
- Algorithm: distribute required visits across available days, respecting max activities/day and target collections as scheduling hints
- Output: draft activities in `planned` status for the rep to review and adjust
- Rep can accept, modify, or reject proposed activities
- Not replacing manual planning — augmenting it
- **Map planner integration:** generated plan shown on map with routes visualized

---

## 4. i18n / Romanian UI ❌

DrMax Romania needs a Romanian-language interface. The tenant config already drives labels for domain concepts (activity types, statuses, field labels), but the UI chrome (buttons, navigation, empty states, error messages) is in English.

### Approach

- Lightweight i18n: tenant config `locale` field already exists (`"locale": "en"`)
- Add a `ui_strings` section to tenant config with overridable UI labels
- Fallback to English for any missing keys
- No heavy i18n framework needed — this is single-tenant, single-locale per deployment
- Key areas: navigation labels, button text ("Create", "Save", "Submit"), planner headers, empty states, error/success toasts

---

## 5. Data Migration from Twenty CRM ❌

For go-live, existing Twenty CRM data needs to be imported:

- **Doctors** — export from per-user `*Doctori` objects → `POST /targets/import`
- **Pharmacies** — export from per-user `*Farmacii` objects → `POST /targets/import`
- **Historical activities** — export from per-user `*Tasks` + master `tasks` → `POST /activities/import` (new endpoint, admin only)
- **Target collections** — if reps have informal groupings in Twenty, import as collections

Script in `scripts/import-twenty.sh` or a Go CLI tool under `cmd/import/`.
