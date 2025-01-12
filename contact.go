package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"google.golang.org/api/people/v1"
)

type contact struct {
	_person         *people.Person
	name            string
	birthday        string
	org             string
	linkedinURL     string
	tier            int
	lastContactDate string
	lastContactNote string
}

func parseGooglePersonToContact(c *people.Person) contact {
	if len(c.Names) == 0 {
		return contact{}
	}

	var birthday string
	if len(c.Birthdays) > 0 {
		birthday = fmt.Sprintf("%02d/%02d/%04d", c.Birthdays[0].Date.Month, c.Birthdays[0].Date.Day, c.Birthdays[0].Date.Year)
	}
	org := ""
	if len(c.Organizations) > 0 {
		org = c.Organizations[0].Name
	}
	linkedin := ""
	if len(c.Urls) > 0 {
		linkedin = c.Urls[0].Value
	}

	lastContactDate := ""
	lastContactNote := ""
	tier := 0
	for i := 0; i < len(c.UserDefined); i++ {
		if c.UserDefined[i].Key == "Tier" {
			tier, _ = strconv.Atoi(c.UserDefined[i].Value)

		} else if c.UserDefined[i].Key == "LastContactDate" {
			lastContactDate = c.UserDefined[i].Value

		} else if c.UserDefined[i].Key == "LastContactNote" {
			lastContactNote = c.UserDefined[i].Value
		}
	}

	return contact{
		name:            c.Names[0].DisplayName,
		birthday:        birthday,
		org:             org,
		linkedinURL:     linkedin,
		tier:            tier,
		lastContactDate: lastContactDate,
		lastContactNote: lastContactNote,
		_person:         c,
	}
}

func updateContactStruct(contact *contact, colName string, val string) error {
	var appErr error

	switch colName {
	case "Name":
		contact.name = val
		contact._person.Names[0].DisplayName = val
	case "Birthday":
		myDate, err := time.Parse("01/02/2006", val)
		if err != nil {
			return err
		}
		contact.birthday = val

		contact._person.Birthdays = nil

		contact._person.Birthdays = append(contact._person.Birthdays, &people.Birthday{
			Date: &people.Date{
				Day:   int64(myDate.Day()),
				Month: int64(myDate.Month()),
				Year:  int64(myDate.Year()),
			},
		})

	case "How I Know Them":
		contact.org = val

		contact._person.Organizations = nil
		contact._person.Organizations = append(contact._person.Organizations, &people.Organization{
			Name: val,
		})
	case "Tier":
		num, err := strconv.Atoi(val)
		if err != nil {
			return err
		}
		contact.tier = num

		foundMatch := false
		for i := 0; i < len(contact._person.UserDefined); i++ {
			if contact._person.UserDefined[i].Key == "Tier" {
				foundMatch = true
				contact._person.UserDefined[i].Value = val
			}
		}
		if !foundMatch {
			contact._person.UserDefined = append(contact._person.UserDefined, &people.UserDefined{
				Key:   "Tier",
				Value: val,
			})
		}
	case "Last Contact Date":
		// Google stores this as a string so we just need to validate that it is a
		// valid date string but don't need to store the Time object
		_, err := time.Parse("2006-01-02", val)
		if err != nil {
			return err
		}
		contact.lastContactDate = val

		foundMatch := false
		for i := 0; i < len(contact._person.UserDefined); i++ {
			if contact._person.UserDefined[i].Key == "LastContactDate" {
				foundMatch = true
				contact._person.UserDefined[i].Value = val
				break
			}
		}
		if !foundMatch {
			contact._person.UserDefined = append(contact._person.UserDefined, &people.UserDefined{
				Key:   "LastContactDate",
				Value: val,
			})
		}
	case "Last Contact Note":
		contact.lastContactNote = val

		foundMatch := false
		for i := 0; i < len(contact._person.UserDefined); i++ {
			if contact._person.UserDefined[i].Key == "LastContactNote" {
				foundMatch = true
				contact._person.UserDefined[i].Value = val
				break
			}
		}
		if !foundMatch {
			contact._person.UserDefined = append(contact._person.UserDefined, &people.UserDefined{
				Key:   "LastContactNote",
				Value: val,
			})
		}
	case "LinkedIn URL":
		contact.linkedinURL = val

		if val == "" {
			contact._person.Urls = nil
			return appErr
		}

		contact._person.Urls = nil
		contact._person.Urls = append(contact._person.Urls, &people.Url{
			Type:  "LinkedIn",
			Value: val,
		})
	default:
		return fmt.Errorf("unknown column header on edit")
	}

	return appErr
}

func generateDetailsString(c contact) string {
	prettyJSON, _ := json.MarshalIndent(c._person, "", "  ")
	return string(prettyJSON)
}
