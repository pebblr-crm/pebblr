# UX Specification: Inline-Editable Activity Detail View

**Task:** pebblr-d315
**Status:** Draft
**Date:** 2026-03-23
**Author:** ux-spec-writer (research agent)

---

## 1. Overview

Replace the current three-component flow (read-only detail page + separate edit page + ActivityForm) with a single unified, inline-editable activity detail view. The new view is designed for field sales reps using tablets and phones while on the go.

### Problem Statement

Current friction points for mobile reps:
- **Two pages** to view and edit the same activity (navigation overhead)
- **Explicit Save button** requires conscious action; data can be lost if the rep navigates away
- **Separate "Change Status" section** at the bottom of the page is easy to miss and requires scrolling
- **Edit link in header** navigates away, breaking flow for quick note edits

### Desired Outcome

A single-page view where all fields are always editable, changes persist automatically, and Submit is the only deliberate action.

---

## 2. Interaction Model

### 2.1 Single Unified View

Delete `/activities/$activityId/edit` as a navigable route. All interactions happen on `/activities/$activityId`.

The detail view renders the same fields as the current edit form. When the activity is not submitted, every editable field is immediately interactive — no mode switch required.

### 2.2 Always-Editable Fields

Fields render as interactive controls (inputs, selects, toggle chips) at all times. There is no "click to edit" pattern; inputs are open and accessible. This suits glove-friendly and touch scenarios where an extra tap is a barrier.

**Exception:** Once `submittedAt` is set, all fields become read-only and the Submit button is hidden. A "Submitted" lock badge is shown in the header (same as current).

### 2.3 Status Inline in Header

The status badge in the header becomes a tappable control. On tap, a compact dropdown/action sheet appears showing valid transitions (sourced from `config.activities.status_transitions[currentStatus]`). Selecting a transition immediately fires `PATCH /activities/:id/status` and dismisses the dropdown.

On desktop, this renders as a styled `<select>` or popover menu anchored to the badge. On mobile, it uses a native `<select>` or bottom sheet to maximize touch target size.

The separate "Change Status" card section is removed.

### 2.4 Auto-Save on Blur/Change

Changes to any field trigger a **debounced PATCH** to `PUT /activities/:id` (full update). The debounce window is **1500ms** after the last change event.

- `onBlur` also immediately flushes the debounce timer (saves as soon as the rep moves focus)
- No explicit Save button anywhere in the view
- The "Update Activity" / "Save" buttons and the Cancel button are removed

#### Save State Indicator

A subtle, non-intrusive indicator in the header shows save state:

| State | Indicator |
|---|---|
| Idle (saved) | Nothing shown |
| Dirty (unsaved) | Small pulsing dot + "Saving…" text in muted grey |
| Saving in flight | Spinner icon, "Saving…" text |
| Save succeeded | Brief checkmark flash (500ms), then cleared |
| Save failed | Red dot + "Not saved — tap to retry" tappable text |

The indicator must be small and positioned so it does not compete with the Submit button.

### 2.5 Submit as the Only Explicit Action

The Submit button remains as the single primary action. It:
1. Flushes any pending debounced save first (awaits the PATCH response)
2. Then calls `POST /activities/:id/submit`
3. On success, switches the entire view to locked/read-only mode and shows the Submitted badge

Submit validation (required fields) is performed server-side on `POST /submit`. The UI does **not** validate required fields during inline editing. If the submit call returns a `422` with `fields` errors, display inline field-level error messages (same mechanism as current form's `serverErrors` prop).

### 2.6 Discard / Navigation Away

Since there is no Cancel button and changes are auto-saved, no discard confirmation is required. If a PATCH is in-flight when the user navigates away, the browser's natural fetch completion handles it (fire-and-forget is acceptable for field updates; submissions are the only critical operation).

---

## 3. Field Behaviors

### 3.1 Core Fields

| Field | Control | Editable | Notes |
|---|---|---|---|
| Activity type | Static label (read-only) | Never | Cannot change type after creation |
| Status | Tappable badge → dropdown | Yes (not submitted) | Drives `PATCH /status` endpoint, not the full update |
| Due date | `<input type="date">` | Yes (not submitted) | |
| Duration | `<select>` | Yes (not submitted) | Options from `config.activities.durations` |
| Target | Search-as-you-type combobox | Yes (not submitted) | Same logic as current form |
| Joint visit user | Text input | Yes (not submitted) | |

### 3.2 Dynamic Fields (per activity type)

Rendered from `typeConfig.fields` exactly as ActivityForm does today, but directly in the detail view without requiring navigation. Field types:

| Type | Control |
|---|---|
| `text` | `<input type="text">` |
| `select` | `<select>` with options from `options_ref` |
| `multi_select` | Toggle chip row (current multi-select UI in ActivityForm) |
| `date` | `<input type="date">` |

Fields with `editable: false` remain read-only regardless of submission status.

### 3.3 Required Field Indication

Required fields (from `fieldDef.required`) show the `*` indicator as they do today. However, **no client-side required validation blocks editing or auto-save**. Required enforcement only triggers on Submit (server validates and returns 422 field errors). This allows reps to save partial progress mid-visit and complete required fields later.

---

## 4. Auto-Save Strategy

### 4.1 Debounce Implementation

```
useDebounce(fieldValues, 1500ms) → triggers PATCH
```

State shape for the inline editor:
- `localData: Partial<UpdateActivityInput>` — mirrors current form state
- `saveState: 'idle' | 'dirty' | 'saving' | 'error'`
- `pendingPatch: ReturnType<typeof useUpdateActivity>` — existing mutation hook

On each field change:
1. Update `localData` (local state)
2. Set `saveState = 'dirty'`
3. Reset debounce timer

On debounce fire / field blur:
1. Set `saveState = 'saving'`
2. Call `updateMutation.mutate(localData)`
3. On success: set `saveState = 'idle'`, update TanStack Query cache (already handled by `useUpdateActivity`)
4. On error: set `saveState = 'error'`, show retry indicator

### 4.2 Optimistic Updates

The existing `useUpdateActivity` hook already does `queryClient.setQueryData` on success. No change needed there. The local `localData` state is the source of truth for the input values; the remote cache is updated on save.

### 4.3 Submit Pre-Flight Save

Before calling `submitMutation.mutate(activityId)`:
1. If `saveState === 'dirty'`, flush the debounce immediately and await the PATCH
2. If `saveState === 'error'`, block Submit and show "Cannot submit — unsaved changes. Tap to retry."
3. Once `saveState === 'idle'`, proceed with `POST /submit`

### 4.4 Status Changes

Status changes (`PATCH /status`) are **separate** from the debounced full update. They fire immediately on selection (no debounce). This keeps status semantics clean and avoids a race condition where a deferred full update overwrites a just-applied status change.

The `status` field in `localData` must be kept in sync after a status PATCH succeeds (update local state from the server response).

---

## 5. Mobile Considerations

### 5.1 Layout

- **Single-column** layout below `sm` breakpoint (≤ 640px). Two-column grid only on wider viewports.
- **Generous tap targets**: all interactive controls min `44px` tall. Current `py-2` buttons meet this for most sizes; date inputs may need explicit `min-h-[44px]`.
- **Sticky Submit bar**: On mobile, the Submit button floats in a sticky bottom bar (fixed to viewport bottom) with a safe-area inset (`pb-safe`). This keeps the primary action reachable without scrolling to the bottom of long forms.
- **No hover-dependent interactions**: The status dropdown must work with tap/touch only — no hover-to-reveal patterns.

### 5.2 Status Dropdown

On viewport width < 640px:
- Render status transitions as a **bottom sheet / action sheet** (slide-up modal) rather than an inline dropdown. Each transition is a large tappable row (min 56px tall).
- The action sheet has a "Cancel" row at the bottom.

On viewport width ≥ 640px:
- Render as a Popover anchored to the status badge, or a styled `<select>`.

### 5.3 Target Search Combobox

The target search combobox presents a challenge on mobile: the dropdown list beneath the input can be cut off by the virtual keyboard. Mitigations:
- Use `position: fixed` for the dropdown list on mobile so it renders above the keyboard.
- Alternatively, tap on the target field opens a full-screen search modal (overlay) on mobile, which avoids keyboard/overflow issues entirely. **Preferred approach.**

### 5.4 Multi-Select Chip Rows

Chip rows should wrap naturally. Each chip must be at least `36px` tall with `8px` horizontal padding — tap-friendly.

### 5.5 Date Inputs

Use native `<input type="date">` which triggers the platform date picker on iOS/Android. Do not use a custom JS date picker — native pickers are more reliable on touch devices.

### 5.6 Font Sizes

Minimum body font size `14px` (`text-sm`). Labels at `11px` (`text-[11px]`) are acceptable for secondary metadata. Never go below `11px`.

---

## 6. Component Architecture

### 6.1 Route: `$activityId.tsx`

This route becomes the sole activity view. The `/edit` route is removed.

High-level structure:
```
ActivityDetailPage
├── ActivityHeader          — type label, status badge (tappable), date, duration, lock badge
├── SaveStateIndicator      — shows saving/error/idle state
├── ActivityCoreFields      — due date, duration, target, joint visit user
├── ActivityDynamicFields   — per-type fields from typeConfig.fields
└── ActivitySubmitBar       — sticky bottom bar with Submit button (hidden when submitted)
```

### 6.2 Custom Hook: `useInlineActivityEditor`

Encapsulates all inline editing state and auto-save logic. Returns:
```ts
interface InlineActivityEditor {
  localData: Partial<UpdateActivityInput>
  saveState: 'idle' | 'dirty' | 'saving' | 'error'
  fieldErrors: ValidationFieldError[]
  handleFieldChange: (key: string, value: unknown) => void
  handleFieldBlur: () => void
  handleStatusChange: (newStatus: string) => void
  handleSubmit: () => Promise<void>
  retrySave: () => void
}
```

This hook owns:
- Debounce timer
- `localData` state
- `saveState` state
- Integration with `useUpdateActivity`, `usePatchActivityStatus`, `useSubmitActivity`

### 6.3 `ActivityForm.tsx`

`ActivityForm.tsx` is still used by the New Activity flow (`/activities/new`). It does not need changes from this spec. The inline edit view uses `useInlineActivityEditor` and renders field controls directly — it does not reuse `ActivityForm`.

If field control rendering is duplicated between `ActivityForm` and the new inline view, extract a `renderActivityField(fieldDef, value, onChange)` utility function in a future refactor. Out of scope for the initial implementation.

### 6.4 Route Cleanup

- Remove `$activityId.edit.tsx` (or keep as a redirect to `$activityId` to avoid broken links in bookmarks/notifications).
- Update the "Edit" link in the current `$activityId.tsx` header — it's replaced by the inline editing model.

---

## 7. Edge Cases

| Scenario | Handling |
|---|---|
| Concurrent edits (two tabs/devices) | Last PATCH wins. No conflict detection required at MVP. |
| PATCH returns 409 conflict | Show generic "Save failed — tap to retry" error. Log to console. |
| Network offline mid-edit | `saveState = 'error'`. Do not block further typing. Retry on reconnect or manual tap. |
| Activity deleted server-side while open | PATCH returns 404. Show "Activity no longer exists" toast and navigate back to dashboard. |
| `submittedAt` set server-side by another session | On PATCH response, check if `submittedAt` is now set. If yes, reload and switch to locked view. |
| Submit while status PATCH in-flight | Disable Submit button while `statusMutation.isPending`. |
| Required field missing on Submit | Server returns 422 with `fields` array. Display field-level errors inline. Re-enable Submit after rep corrects fields. |
| Activity type with no dynamic fields | `ActivityDynamicFields` renders nothing. Layout remains clean. |

---

## 8. Out of Scope (Future)

- Offline-first / local queue for saves when fully offline
- Conflict resolution / optimistic locking
- Real-time collaborative editing (SSE push for external updates)
- Rich text / markdown fields
- Photo/file attachment fields
- Undo / revision history

---

## 9. Open Questions for Implementation

1. **Bottom sheet component**: Does the project have an existing bottom sheet / modal sheet primitive, or does one need to be built?
2. **Target full-screen search**: Confirm the full-screen search modal approach is preferred over the scrollable combobox for mobile target selection.
3. **Redirect from `/edit`**: Keep the edit route as a silent redirect or remove the route entirely? Removing is cleaner; a redirect is safer if external systems link to the edit URL.
4. **`usePatchActivityStatus` race condition**: Currently the full PUT overrides all fields including status. Confirm the backend correctly ignores a stale status value in the body if a status PATCH is in-flight, or implement request ordering.
