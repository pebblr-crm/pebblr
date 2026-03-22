package store

// Store aggregates all repository interfaces.
// Pass a Store to services that need access to persistent data.
type Store interface {
	Users() UserRepository
	Teams() TeamRepository
	Targets() TargetRepository
	CalendarEvents() CalendarEventRepository
}
