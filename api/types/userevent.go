package types

import (
	"strconv"
	"time"
)

type Payload struct {
	UserEvents []UserEvent
}

type UserEvent struct {
	Time        time.Time `json:"time"`
	Uuid        string    `json:"uuid"`
	Id          int64     `json:"id"`
	Active      bool      `json:"active"`
	Email       string    `json:"email"`
	Phone       string    `json:"phone"`
	Description string    `json:"description"`
	Credits     float64   `json:"credits"`
}

func (u UserEvent) CSVData() (data []string) {
	return []string{
		u.Time.Format(time.RFC3339Nano),
		u.Uuid,
		strconv.FormatInt(u.Id, 10),  // Format ID as string
		strconv.FormatBool(u.Active), // Format Active as string
		u.Email,
		u.Phone,
		u.Description,
		strconv.FormatFloat(u.Credits, 'f', -1, 64), // Format Credits as string
	}
}

func (u UserEvent) Equal(o UserEvent) bool {
	return u.Time.Equal(o.Time) &&
		u.Uuid == o.Uuid &&
		u.Id == o.Id &&
		u.Active == o.Active &&
		u.Email == o.Email &&
		u.Phone == o.Phone &&
		u.Description == o.Description &&
		u.Credits == o.Credits
}
