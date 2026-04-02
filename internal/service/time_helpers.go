package service

import "time"

// calendarMonths returns the number of calendar months spanned by the date range (minimum 1).
func calendarMonths(from, to time.Time) int {
	if to.Before(from) {
		return 1
	}
	months := (to.Year()-from.Year())*12 + int(to.Month()) - int(from.Month()) + 1
	if months < 1 {
		return 1
	}
	return months
}
