package api

const (
	headerContentType     = "Content-Type"
	contentTypeJSON       = "application/json"
	errMissingUser        = "missing authenticated user"
	errInvalidRequestBody = "invalid request body"
	errUnexpected         = "an unexpected error occurred"
	errInvalidPeriod      = "invalid period or date format"
	dateFormat            = "2006-01-02"

	// maxPaginationLimit caps the number of items a client can request per page.
	maxPaginationLimit = 200

	// maxImportItems caps the number of targets in a single import request.
	maxImportItems = 1000

	// maxBatchItems caps the number of items in a single batch create request.
	maxBatchItems = 100
)
