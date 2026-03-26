# Rep (Salesperson) Persona — Primary Design Focus

The rep is the core user. They visit customer locations on a recurring schedule, plan their routes, and submit feedback after each visit.

## What the Rep Sees

- **Their assigned targets:** ~150 per quarter, organized by type (e.g., doctor, pharmacy)
- **Target details:** Name, classification (A/B/C), location, specialty, and other fields depending on type
- **Their planner:** A calendar/schedule view for planning and tracking visits
- **Their compliance stats:** Am I meeting visit frequency targets? What's my coverage?
- **Their recovery balance:** Days earned (weekend work) vs. days taken

## Core Flow

### 1. Quarterly Assignment
Each quarter, the rep receives ~150 targets. Targets are classified by priority:
- **A-class:** High priority — must be visited frequently (e.g., 4×/quarter)
- **B-class:** Medium — visited less often (e.g., 2×/quarter)
- **C-class:** Low — visited least (e.g., 1×/quarter)

The rep needs to see at a glance: how many targets, what classifications, and what the visit requirements are for the period.

### 2. Planning Visits
Visits repeat on a ~3-week cycle. A typical day has 10–15 activities.

**Key planning behaviors:**
- **Geo-based batching (P0):** Reps plan visits by geographic proximity for route efficiency. A map view where they can select nearby targets and schedule them as a batch is essential.
- **Week-at-a-time scheduling:** Reps typically plan a full week of visits, then repeat the pattern. They need a single-pane view of the week's schedule.
- **Week cloning:** A planned week can be duplicated to a future week, preserving the day-of-week pattern (e.g., Monday targets stay on Monday). This is the primary reuse mechanism.
- **Saved target groups (collections):** Reps can save sets of targets (e.g., "Bucharest North route") for quick reference when planning. These are target lists, not schedules.
- **Visit types:** Some visits are face-to-face (route-relevant), others are remote (not route-relevant). The planner should distinguish these.
- **Joint visits:** Occasionally two reps visit a target together. Both need to see and manage the shared activity.

**Constraints the planner must respect:**
- Minimum days between visits to the same target (e.g., 21 days)
- Maximum activities per day (e.g., 10)
- Blocking days: vacation, holidays, and recovery days block field activities
- A-class targets need more visits than C-class — the planner should surface who's under-visited

### 3. Executing & Submitting (P0)
During or after a visit, the rep updates the activity:
- Transitions status (e.g., planned → completed or cancelled)
- Submits required feedback (fields vary by target type — e.g., doctor visits may require different info than pharmacy visits)
- Submission locks the activity — no further edits

This needs to be fast and low-friction. Reps are doing this 10–15 times a day, often on mobile or between appointments.

### 4. Non-Visit Activities
Other activity types include:
- Vacation, public holidays
- Training, team meetings, cycle meetings
- Business travel, business meals
- Administrative time, lunch breaks
- Recovery days (earned by working weekends, claimable within a window)

Non-visit activities may block field visits for that day (e.g., vacation = no visits).

### 5. Tracking Progress
Reps need a lightweight dashboard showing:
- **Coverage:** What % of their targets have been visited this period?
- **Frequency compliance:** Per classification (A/B/C), am I meeting the required visit count?
- **Activity breakdown:** How many planned vs. completed vs. cancelled?
- **Recovery balance:** Days earned vs. taken (if applicable)

This should be glanceable — not a separate analytics page, but integrated into the daily/weekly view.

## UX Notes

- **Geography first, time second.** Planning starts from "where am I going," not "what time is it."
- **Batch-friendly.** Every interaction should assume the rep is doing it 10+ times. Minimize clicks, support multi-select, allow bulk actions.
- **Field-friendly.** Feedback submission happens between appointments — assume constrained attention, small screen, one hand.
- **Gentle compliance nudges.** Show "3 A-class targets need visits this week" inline, not buried in a dashboard.
- **The 3-week cycle matters.** The calendar isn't standard weekly — help users think in "visiting windows," not just Mon–Fri.
