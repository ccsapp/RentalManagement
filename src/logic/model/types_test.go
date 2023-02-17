package model

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestTimePeriod_RestrictTo_equal(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := *timePeriod.RestrictTo(&timePeriod)

	assert.Equal(t, timePeriod, restricted)
}

func TestTimePeriod_RestrictTo_inside(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := *timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Equal(t, TimePeriod{
		StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 12, 1, 0, 0, 0, 0, time.UTC),
	}, restricted)
}

func TestTimePeriod_RestrictTo_outside(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := *timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Equal(t, timePeriod, restricted)
}

func TestTimePeriod_RestrictTo_leftOverlap(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := *timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Equal(t, TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
	}, restricted)
}

func TestTimePeriod_RestrictTo_rightOverlap(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := *timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2022, 6, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Equal(t, TimePeriod{
		StartDate: time.Date(2020, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}, restricted)
}

func TestTimePeriod_RestrictTo_leftOutside(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2019, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2019, 12, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Nil(t, restricted)
}

func TestTimePeriod_RestrictTo_rightOutside(t *testing.T) {
	timePeriod := TimePeriod{
		StartDate: time.Date(2020, 1, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	restricted := timePeriod.RestrictTo(&TimePeriod{
		StartDate: time.Date(2021, 6, 1, 0, 0, 0, 0, time.UTC),
		EndDate:   time.Date(2022, 12, 1, 0, 0, 0, 0, time.UTC),
	})

	assert.Nil(t, restricted)
}
