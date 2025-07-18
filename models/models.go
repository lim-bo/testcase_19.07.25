package models

import (
	"time"

	"github.com/google/uuid"
)

type Subscription struct {
	ID      int       `json:"id,omitempty"`
	Name    string    `json:"name"`
	Price   int       `json:"price"`
	UID     uuid.UUID `json:"uid"`
	Start   time.Time `json:"start_date"`
	Expires time.Time `json:"expires,omitempty"`
}
