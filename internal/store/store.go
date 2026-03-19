package store

// Store aggregates all repository interfaces.
// Pass a Store to services that need access to persistent data.
type Store interface {
	Leads() LeadRepository
	Users() UserRepository
	Teams() TeamRepository
	Events() EventRepository
	Customers() CustomerRepository
}
