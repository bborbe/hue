// Copyright (c) 2025 Benjamin Borbe All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package time

import (
	"context"
	stdtime "time"

	"github.com/bborbe/collection"
	"github.com/bborbe/errors"
	"github.com/bborbe/parse"
	"github.com/bborbe/validation"
)

const (
	Sunday Weekday = iota
	Monday
	Tuesday
	Wednesday
	Thursday
	Friday
	Saturday
)

var AvailableWeekdays = Weekdays{
	Sunday,
	Monday,
	Tuesday,
	Wednesday,
	Thursday,
	Friday,
	Saturday,
}

func ParseWeekdays(ctx context.Context, values any) (Weekdays, error) {
	ints, err := parse.ParseIntArray(ctx, values)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse values failed")
	}
	return AsWeekdays(ints), nil
}

func AsWeekdays[T ~int](values []T) Weekdays {
	result := make(Weekdays, len(values))
	for i, w := range values {
		result[i] = AsWeekday(w)
	}
	return result
}

type Weekdays []Weekday

func (w Weekdays) Weekdays() []stdtime.Weekday {
	result := make([]stdtime.Weekday, len(w))
	for i, weekday := range w {
		result[i] = weekday.Weekday()
	}
	return result
}

func (w Weekdays) Validate(ctx context.Context) error {
	for _, w := range w {
		if err := w.Validate(ctx); err != nil {
			return errors.Wrap(ctx, err, "validation failed")
		}
	}
	return nil
}

func (w Weekdays) Contains(value Weekday) bool {
	return collection.Contains(w, value)
}

func ParseWeekday(ctx context.Context, value any) (*Weekday, error) {
	i, err := parse.ParseInt(ctx, value)
	if err != nil {
		return nil, errors.Wrapf(ctx, err, "parse value failed")
	}
	return AsWeekday(i).Ptr(), nil
}

func AsWeekday[T ~int](value T) Weekday {
	return Weekday(value)
}

type Weekday stdtime.Weekday

func (w Weekday) Validate(ctx context.Context) error {
	if AvailableWeekdays.Contains(w) == false {
		return errors.Wrapf(ctx, validation.Error, "Weekdays contains invalid value")
	}
	return nil
}

func (w Weekday) String() string {
	return w.Weekday().String()
}

func (w Weekday) Weekday() stdtime.Weekday {
	return stdtime.Weekday(w)
}

func (w Weekday) Ptr() *Weekday {
	return &w
}
