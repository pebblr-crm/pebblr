package domain

// Customer represents a business or individual that a sales rep visits.
type Customer struct {
	ID      string
	Name    string
	Type    CustomerType
	Address Address
	// Phone is the primary contact phone number.
	Phone string
	// Email is the primary contact email address.
	Email string
	// Notes holds free-form notes about the customer.
	Notes string
}

// Address holds the physical location of a customer.
type Address struct {
	Street  string
	City    string
	State   string
	Country string
	Zip     string
}
