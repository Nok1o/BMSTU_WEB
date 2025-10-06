package forms

import (
	"errors"
	"net/url"
	"strconv"
)

//easyjson:json
type SearchForm struct {
	ToSearch string `json:"to_search" validate:"required"`
	Count    uint   `json:"count" validate:"required"`
}

func (s *SearchForm) Unpack(values url.Values) error {
	if !values.Has("to_search") {
		return errors.New("username parameter missing")
	}

	if !values.Has("count") {
		return errors.New("count parameter missing")
	}

	s.ToSearch = values.Get("to_search")

	usersCount, err := strconv.ParseInt(values.Get("count"), 10, 64)
	if err != nil {
		return errors.New("failed to parse count")
	}
	if usersCount < 0 {
		return errors.New("count must be positive")
	}

	s.Count = uint(usersCount)
	return nil
}
