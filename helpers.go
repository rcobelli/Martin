package main

import (
	"math"
	"sort"
	"time"
)

func isTooLongAgo(date string, months int) bool {
	myDate, err := time.Parse("2006-01-02", date)
	if err != nil {
		panic(err)
	}

	delta := time.Now().AddDate(0, -months, 0)
	return myDate.Before(delta)
}

func sortData(key int, data *[]contact) {
	sort.Slice(*data, func(i, j int) bool {
		// If we get a negative key, we need to swap the variables to sort ascending
		if key < 0 {
			k := i
			i = j
			j = k
		}

		// Ensure we don't attempt a negative index
		col := int(math.Abs(float64(key)))

		switch COLUMN_HEADERS[col] {
		case "Name":
			return (*data)[i].name < (*data)[j].name
		case "Birthday":
			return (*data)[i].birthday < (*data)[j].birthday
		case "How I Know Them":
			return (*data)[i].org < (*data)[j].org
		case "Tier":
			return (*data)[i].tier < (*data)[j].tier
		case "Last Contact Date":
			return (*data)[i].lastContactDate < (*data)[j].lastContactDate
		case "Last Contact Note":
			return (*data)[i].lastContactNote < (*data)[j].lastContactNote
		case "LinkedIn URL":
			return (*data)[i].linkedinURL < (*data)[j].linkedinURL
		default:
			panic("Unknown sort field")
		}
	})
}
