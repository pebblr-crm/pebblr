package domain

import "time"

// Customer represents a business or individual that a sales rep visits.
type Customer struct {
	ID      string       `json:"id"`
	Name    string       `json:"name"`
	Type    CustomerType `json:"type"`
	Address Address      `json:"address"`
	// Phone is the primary contact phone number.
	Phone string `json:"phone"`
	// Email is the primary contact email address.
	Email string `json:"email"`
	// Notes holds free-form notes about the customer.
	Notes     string    `json:"notes"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

// Address holds the physical location of a customer.
type Address struct {
	Street  string `json:"street"`
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Zip     string `json:"zip"`
}
