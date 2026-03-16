package domain

// CustomerType classifies the type of customer.
type CustomerType string

const (
	// CustomerTypeRetail is a retail storefront.
	CustomerTypeRetail CustomerType = "retail"
	// CustomerTypeWholesale is a wholesale distributor.
	CustomerTypeWholesale CustomerType = "wholesale"
	// CustomerTypeHospitality is a hospitality venue (hotel, restaurant, etc.).
	CustomerTypeHospitality CustomerType = "hospitality"
	// CustomerTypeInstitutional is an institutional buyer (school, hospital, etc.).
	CustomerTypeInstitutional CustomerType = "institutional"
	// CustomerTypeOther covers types not listed above.
	CustomerTypeOther CustomerType = "other"
)

// Valid returns true if the customer type is a recognized value.
func (t CustomerType) Valid() bool {
	switch t {
	case CustomerTypeRetail, CustomerTypeWholesale, CustomerTypeHospitality,
		CustomerTypeInstitutional, CustomerTypeOther:
		return true
	}
	return false
}
