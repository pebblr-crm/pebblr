# Phase 4 — Post Go-Live Optimizations ❌

> Back to [overview](PLAN.md)

## Checklist

20. ❌ Weekend activity + recovery days
21. ❌ Drag & drop calendar
22. ❌ Copy-paste activities
23. ❌ Advanced filtering with saved filters
24. ❌ Target group management (quarterly)
25. ❌ Plan generation (rule-based monthly plan proposal)

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

## 2. Drag & Drop Calendar ❌

Enhance the planner with drag-and-drop:

- Drag activities between days to reschedule
- Drag to resize (change duration)
- Visual feedback for blocked days (vacation/holiday)
- Calls `PUT /activities/:id` with updated `due_date` and `duration`
- Blocked if activity is submitted

---

## 3. Copy-Paste Activities ❌

Allow reps to duplicate activities:

- Select one or more activities → "Copy"
- Click a target day → "Paste" (creates new activities with same type/fields, new date)
- Useful for recurring weekly patterns
- Copied activities get `planificat` status regardless of source status

---

## 4. Advanced Filtering with Saved Filters ❌

- Filter activities by: type, status, date range, target, creator, team, routing week
- Save filter combinations as named presets (stored in localStorage or user preferences)
- Quick-switch between saved filters in planner and activity list

---

## 5. Target Group Management (Quarterly) ❌

Each quarter, managers define which targets each rep should focus on:

- Admin/manager UI to assign targets to reps for the quarter
- Bulk assignment (CSV upload or multi-select)
- Track changes: which targets were added/removed from target group
- Frequency rules apply to target group specifically
- Dashboard shows target group coverage vs all-target coverage

---

## 6. Plan Generation ❌

Auto-propose a monthly activity plan based on rules:

- Input: rep's target group, frequency rules, blocked days (holidays, vacations), routing preferences
- Algorithm: distribute required visits across available days, respecting max activities/day
- Output: draft activities in `planificat` status for the rep to review and adjust
- Rep can accept, modify, or reject proposed activities
- Not replacing manual planning — augmenting it

---

## 7. Data Migration from Twenty CRM ❌

For go-live, existing Twenty CRM data needs to be imported:

- **Doctors** — export from per-user `*Doctori` objects → `POST /targets/import`
- **Pharmacies** — export from per-user `*Farmacii` objects → `POST /targets/import`
- **Historical activities** — export from per-user `*Tasks` + master `tasks` → `POST /activities/import` (new endpoint, admin only)

Script in `scripts/import-twenty.sh` or a Go CLI tool under `cmd/import/`.
