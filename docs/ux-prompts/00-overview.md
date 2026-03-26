# Pebblr — UX Design Prompt

A planning and tracking tool for field salespeople who visit customers on recurring schedules.

This prompt is split by persona. The **Rep** persona is the primary design focus.

## Personas

| Persona | Focus | File |
|---------|-------|------|
| Admin | Configuration, audits, user management | [01-admin.md](01-admin.md) |
| Manager | Team oversight, dashboards, alerts | [02-manager.md](02-manager.md) |
| Rep (salesperson) | Planning, executing, tracking visits | [03-rep.md](03-rep.md) |

## Design Priorities

| Priority | Capability |
|----------|------------|
| **P0** | Geo-based batch scheduling of visits (map + list) |
| **P0** | Fast activity completion and feedback submission (10–15×/day) |
| **P1** | Week-at-a-glance planner with clone/reuse |
| **P1** | At-a-glance compliance tracking (am I on target?) |
| **P2** | Saved target groups (collections) for planning shortcuts |
| **P2** | Joint visit coordination |
| **P2** | Recovery day balance and claiming |

## Key UX Considerations

- The heaviest interaction is scheduling + completing visits. Minimize clicks for these.
- Reps think in geography first, then time. Planning starts from "where," not "when."
- Feedback submission happens in the field — assume constrained attention, possibly mobile.
- Compliance info should create gentle nudges, not dashboards you have to seek out.
- The 3-week cycle means the calendar isn't standard weekly — the system should help users think in "visiting windows," not just calendar weeks.
