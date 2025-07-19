package models

import (
	"errors"
	"time"

	"github.com/bytedance/sonic"
	"github.com/google/uuid"
)

type Subscription struct {
	ID      int        `json:"id,omitempty" example:"1"`
	Name    string     `json:"name" example:"yandex"`
	Price   int        `json:"price" example:"400"`
	UID     uuid.UUID  `json:"uid" example:"60601fee-2bf1-4721-ae6f-7636e79a0cba"`
	Start   time.Time  `json:"start_date" example:"01-2025"`
	Expires *time.Time `json:"expires,omitempty" example:"02-2026"`
}

const layout = "01-2006"

func (s Subscription) MarshalJSON() ([]byte, error) {
	type Alias Subscription
	dst := &struct {
		Start   string  `json:"start_date"`
		Expires *string `json:"expires,omitempty"`
		*Alias
	}{
		Start: s.Start.Format(layout),
		Alias: (*Alias)(&s),
	}
	if s.Expires != nil {
		expiresStr := s.Expires.Format(layout)
		dst.Expires = &expiresStr
	}
	return sonic.Marshal(dst)
}

func (s *Subscription) UnmarshalJSON(data []byte) error {
	type Alias Subscription
	dst := &struct {
		Start   string  `json:"start_date"`
		Expires *string `json:"expires,omitempty"`
		*Alias
	}{
		Alias: (*Alias)(s),
	}
	if err := sonic.Unmarshal(data, &dst); err != nil {
		return err
	}
	var err error
	if s.Start, err = time.Parse(layout, dst.Start); err != nil {
		return errors.New("invalid start_date format: " + err.Error())
	}
	if dst.Expires != nil {
		parsed, err := time.Parse(layout, *dst.Expires)
		if err != nil {
			return errors.New("invalid expires format: " + err.Error())
		}
		s.Expires = &parsed
	}
	return nil
}

type ListOpts struct {
	Limit  int
	Offset int
	Filter map[string]interface{}
	Order  string
}
