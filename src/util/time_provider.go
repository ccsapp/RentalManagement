package util

import "time"

//go:generate mockgen -source=./time_provider.go -package=mocks -destination=../mocks/mock_time_provider.go

type ITimeProvider interface {
	Now() time.Time
}

type TimeProvider struct{}

func (tp TimeProvider) Now() time.Time {
	return time.Now()
}
