package handlers

import (
	"context"
	"time"
)

var availableDates map[int]map[string][]time.Time

// InitTestDates creates and returns a map with all test availableDates
// All availableDates are set statically via now.AddDate()
func InitTestDates() {
	availableDates = make(map[int]map[string][]time.Time)
	now := time.Now().UTC()
	availableDates[1] = make(map[string][]time.Time)

	// Private tours - 8
	availableDates[1]["private"] = []time.Time{
		now.AddDate(0, 0, 2),  // The day after tomorrow
		now.AddDate(0, 0, 3),  // +3 days
		now.AddDate(0, 0, 5),  // +5 days
		now.AddDate(0, 0, 7),  // +7 days
		now.AddDate(0, 0, 9),  // +9 days
		now.AddDate(0, 0, 10), // +10 days
		now.AddDate(0, 0, 12), // +12 days
		now.AddDate(0, 0, 14), // +14 days
	}

	// Group tours - 14
	availableDates[1]["public"] = []time.Time{
		now.AddDate(0, 0, 1),  // Tomorrow
		now.AddDate(0, 0, 2),  // The day after tomorrow
		now.AddDate(0, 0, 3),  // +3 days
		now.AddDate(0, 0, 4),  // +4 days
		now.AddDate(0, 0, 5),  // +5 days
		now.AddDate(0, 0, 6),  // +6 days
		now.AddDate(0, 0, 7),  // +7 days
		now.AddDate(0, 0, 8),  // +8 days
		now.AddDate(0, 0, 9),  // +9 days
		now.AddDate(0, 0, 10), // +10 days
		now.AddDate(0, 0, 11), // +11 days
		now.AddDate(0, 0, 12), // +12 days
		now.AddDate(0, 0, 13), // +13 days
		now.AddDate(0, 0, 14), // +14 days
	}

	availableDates[2] = make(map[string][]time.Time)

	availableDates[2]["no"] = []time.Time{
		now.AddDate(0, 0, 1), // Tomorrow
		now.AddDate(0, 0, 2), // The day after tomorrow
		now.AddDate(0, 0, 3), // +3 days
		now.AddDate(0, 0, 4), // +4 days
		now.AddDate(0, 0, 5), // +5 days
		now.AddDate(0, 0, 6), // +6 days
		now.AddDate(0, 0, 7), // +7 days
	}

	availableDates[3] = make(map[string][]time.Time)

	availableDates[3]["private"] = []time.Time{
		now.AddDate(0, 0, 2),
		now.AddDate(0, 0, 3),
		now.AddDate(0, 0, 9),
		now.AddDate(0, 0, 10),
		now.AddDate(0, 0, 16),
		now.AddDate(0, 0, 17),
	}

	availableDates[3]["public"] = []time.Time{
		now.AddDate(0, 0, 1),  // Tomorrow
		now.AddDate(0, 0, 3),  // +3 days
		now.AddDate(0, 0, 4),  // +4 days
		now.AddDate(0, 0, 5),  // +5 days
		now.AddDate(0, 0, 6),  // +6 days
		now.AddDate(0, 0, 7),  // +7 days
		now.AddDate(0, 0, 8),  // +8 days
		now.AddDate(0, 0, 9),  // +9 days
		now.AddDate(0, 0, 10), // +10 days
		now.AddDate(0, 0, 11), // +11 days
	}

	availableDates[4] = make(map[string][]time.Time)
	availableDates[4]["no"] = []time.Time{}

	availableDates[5] = make(map[string][]time.Time)

	availableDates[5]["no"] = []time.Time{
		now.AddDate(0, 0, 3),  // 3 days
		now.AddDate(0, 0, 5),  // 5 days
		now.AddDate(0, 0, 10), // 10 days
		now.AddDate(0, 0, 15), // 15 days
		now.AddDate(0, 0, 20), // 20 days
	}

	availableDates[6] = make(map[string][]time.Time)

	availableDates[6]["private"] = []time.Time{
		now.AddDate(0, 0, 1),  // Tomorrow
		now.AddDate(0, 0, 2),  // The day after tomorrow
		now.AddDate(0, 0, 3),  // +3 days
		now.AddDate(0, 0, 4),  // +4 days
		now.AddDate(0, 0, 5),  // +5 days
		now.AddDate(0, 0, 8),  // +8 days
		now.AddDate(0, 0, 9),  // +9 days
		now.AddDate(0, 0, 10), // +10 days
	}

	availableDates[6]["public"] = []time.Time{
		now.AddDate(0, 0, 2),
		now.AddDate(0, 0, 3),
		now.AddDate(0, 0, 9),
		now.AddDate(0, 0, 10),
	}
}

// GetAvailableDates retrieves the list of available dates for the service
func GetAvailableDates(
	ctx context.Context,
	serviceID int,
	visitType string,
	limit int,
) ([]time.Time, error) {

	// We get the dates for the specified service
	if serviceavailableDates, ok := availableDates[serviceID]; ok {
		if availableDates, ok := serviceavailableDates[visitType]; ok {
			// We limit the quantity
			if len(availableDates) > limit {
				return availableDates[:limit], nil
			}
			return availableDates, nil
		}
	}

	// If there are no dates for the specified type, we return an empty list.
	return []time.Time{}, nil
}

// IsDateAvailable checks if the date is in the list of available dates.
func IsDateAvailable(
	ctx context.Context,
	serviceID int,
	date time.Time,
	visitType string,
) (bool, error) {

	// Normalize the date (discard the time)
	date = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())

	// We get the dates for the service
	if serviceavailableDates, ok := availableDates[serviceID]; ok {
		if availableDates, ok := serviceavailableDates[visitType]; ok {
			for _, d := range availableDates {
				// We only compare dates, without time
				dNorm := time.Date(d.Year(), d.Month(), d.Day(), 0, 0, 0, 0, d.Location())
				if dNorm.Equal(date) {
					return true, nil
				}
			}
		}
	}

	return false, nil
}
