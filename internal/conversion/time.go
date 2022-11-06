package conversion

import "time"

var Durations = map[string]map[string]TimeDiff{
	"": {
		"s": TimeDiff{Duration: time.Second},
		"m": TimeDiff{Duration: time.Minute},
		"h": TimeDiff{Duration: time.Hour},
		"d": TimeDiff{Days: 1},
		"w": TimeDiff{Days: 7},
		"M": TimeDiff{Months: 1},
		"q": TimeDiff{Months: 3},
		"y": TimeDiff{Years: 1},
	},
	"en": {
		"second":  TimeDiff{Duration: time.Second},
		"seconds": TimeDiff{Duration: time.Second},
		"minute":  TimeDiff{Duration: time.Minute},
		"minutes": TimeDiff{Duration: time.Minute},
		"hour":    TimeDiff{Duration: time.Hour},
		"hours":   TimeDiff{Duration: time.Hour},
		"day":     TimeDiff{Days: 1},
		"days":    TimeDiff{Days: 1},
		"week":    TimeDiff{Days: 7},
		"weeks":   TimeDiff{Days: 7},
		"month":   TimeDiff{Months: 1},
		"months":  TimeDiff{Months: 1},
		"year":    TimeDiff{Years: 1},
		"years":   TimeDiff{Years: 1},
	},
	"sv": {
		"sekund":   TimeDiff{Duration: time.Second},
		"sekunder": TimeDiff{Duration: time.Second},
		"minute":   TimeDiff{Duration: time.Minute},
		"minuter":  TimeDiff{Duration: time.Minute},
		"timme":    TimeDiff{Duration: time.Hour},
		"timmar":   TimeDiff{Duration: time.Hour},
		"dag":      TimeDiff{Days: 1},
		"dagar":    TimeDiff{Days: 1},
		"vecka":    TimeDiff{Days: 7},
		"veckor":   TimeDiff{Days: 7},
		"månad":    TimeDiff{Months: 1},
		"månader":  TimeDiff{Months: 1},
		"år":       TimeDiff{Years: 1},
	},
}

// TimeDiff is an overeager extension to time.Duration allowing diffs
// by days, months, and years as well (all of which have unfortunate
// and annoyingly moving properties).
type TimeDiff struct {
	Duration time.Duration
	Days     int16
	Months   int8
	Years    int8
}

// AddTime modifies the given [time.Time] with the duration specified
// by TimeDiff.
func (d TimeDiff) AddTime(t time.Time) time.Time {
	t = t.Add(d.Duration)
	t = t.AddDate(int(d.Years), int(d.Months), int(d.Days))
	return t
}

// ApproximateDuration returns the average Julian duration of a
// TimeDiff. This should be good enough for any real cooking
// application, since a recipe imprecise enough to say
// "wait three months" is unlikely to care about the exact length of
// those months.
func (d TimeDiff) ApproximateDuration() time.Duration {
	return d.Duration +
		time.Duration(d.Days)*24*time.Hour +
		time.Duration(d.Months)*304/10*24*time.Hour +
		time.Duration(d.Years)*36525/100*24*time.Hour
}
