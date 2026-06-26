package scheduler

import (
	"encoding/json"
	"errors"
	"sort"
	"time"

	cron "github.com/robfig/cron/v3"
)

// Public API
// NextRunTime computes the next time the given schedule will run at or after the provided reference time.
// If no reference time is provided, time.Now() is used.
// If the schedule has no future run time (e.g., ended), ErrNoNextRun is returned.
func NextRunTime(scheduleJSON []byte, from ...time.Time) (time.Time, error) {
	var zero time.Time
	ref := time.Now()
	if len(from) > 0 {
		ref = from[0]
	}
	ref = ref.Truncate(time.Minute)

	var s Schedule
	if err := json.Unmarshal(scheduleJSON, &s); err != nil {
		return zero, err
	}

	s.normalize()

	sRef := ref
	// Respect startAt/endAt window if recurring
	if s.Kind == "recurring" {
		if s.StartAt.IsZero() {
			return zero, errors.New("recurring schedule missing startAt")
		}
		if ref.Before(s.StartAt) {
			sRef = s.StartAt
		}
		if !s.EndAt.IsZero() && !sRef.Before(s.EndAt) {
			return zero, ErrNoNextRun
		}
	}

	switch s.Kind {
	case "single":
		if s.Single.RunAt.IsZero() {
			return zero, errors.New("single schedule missing runAt")
		}
		if s.Single.RunAt.Before(ref) {
			return zero, ErrNoNextRun
		}
		return s.Single.RunAt, nil
	case "recurring":
		if s.Mode == "cron" {
			return nextCron(s.Cron, sRef, s.EndAt)
		}
		return nextStructured(&s.Structured, sRef, s.StartAt, s.EndAt)
	default:
		return zero, errors.New("unknown schedule kind")
	}
}

var ErrNoNextRun = errors.New("no next run time")

// JSON model

type Schedule struct {
	Kind string `json:"kind"`

	// single
	Single struct {
		RunAt time.Time `json:"runAt"`
	} `json:"-"`

	// recurring
	Mode       string     `json:"mode,omitempty"`
	StartAt    time.Time  `json:"startAt,omitempty"`
	EndAt      time.Time  `json:"endAt,omitempty"`
	Cron       string     `json:"cron,omitempty"`
	Structured Structured `json:"rule,omitempty"`

	// raw fields for custom unmarshal
	RunAtRaw string `json:"runAt"`
}

type Structured struct {
	Freq string `json:"freq"`

	Interval   int    `json:"interval,omitempty"`
	MinuteMark int    `json:"minuteMark,omitempty"`
	TimeHHMM   string `json:"time,omitempty"`
	Weekdays   []int  `json:"weekdays,omitempty"`
	Mode       string `json:"mode,omitempty"`
	Days       []int  `json:"days,omitempty"`
	Day        int    `json:"day,omitempty"`
	Ordinal    string `json:"ordinal,omitempty"`
	Weekday    any    `json:"weekday,omitempty"`
	Months     []int  `json:"months,omitempty"`
	Month      int    `json:"month,omitempty"`
}

func (s *Schedule) normalize() {
	// Parse single runAt if present
	if s.RunAtRaw != "" {
		if t, err := time.Parse(time.RFC3339, s.RunAtRaw); err == nil {
			s.Single.RunAt = t.Truncate(time.Minute)
		}
	}
	// Ensure StartAt/EndAt minute granularity
	s.StartAt = s.StartAt.Truncate(time.Minute)
	s.EndAt = s.EndAt.Truncate(time.Minute)

	// Parse rule into Structured already bound via json tags
	// Nothing else needed here.
}

// cron mode
func nextCron(cronstr string, ref time.Time, end time.Time) (time.Time, error) {
	var zero time.Time
	if cronstr == "" {
		return zero, errors.New("cron schedule missing cron string")
	}
	parser := cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow)
	sched, err := parser.Parse(cronstr)
	if err != nil {
		return zero, err
	}
	// Inclusive: if ref is an exact run moment, return ref
	cand := sched.Next(ref.Add(-1 * time.Second))
	if !cand.IsZero() && cand.Equal(ref) {
		if !end.IsZero() && cand.After(end) {
			return zero, ErrNoNextRun
		}
		return cand, nil
	}
	cand = sched.Next(ref)
	if !end.IsZero() && (cand.IsZero() || cand.After(end)) {
		return zero, ErrNoNextRun
	}
	return cand, nil
}

// structured mode
func nextStructured(r *Structured, ref time.Time, start time.Time, end time.Time) (time.Time, error) {
	switch r.Freq {
	case "minute":
		iv := max(1, r.Interval)
		anchor := start
		return nextEveryNMinutes(anchor, ref, iv, end)
	case "hour":
		iv := max(1, r.Interval)
		mm := clamp(r.MinuteMark, 0, 59)
		anchor := time.Date(start.Year(), start.Month(), start.Day(), start.Hour(), mm, 0, 0, start.Location())
		return nextEveryNHours(anchor, ref, iv, mm, end)
	case "day":
		iv := max(1, r.Interval)
		hh, mn := parseHHMM(r.TimeHHMM)
		anchor := time.Date(start.Year(), start.Month(), start.Day(), hh, mn, 0, 0, start.Location())
		return nextEveryNDays(anchor, ref, iv, hh, mn, end)
	case "week":
		iv := max(1, r.Interval)
		hh, mn := parseHHMM(r.TimeHHMM)
		wds := append([]int(nil), r.Weekdays...)
		sort.Ints(wds)
		return nextEveryNWeeks(start, ref, iv, wds, hh, mn, end)
	case "month":
		iv := max(1, r.Interval)
		hh, mn := parseHHMM(r.TimeHHMM)
		if r.Mode == "ordinal" {
			return nextMonthlyOrdinal(start, ref, iv, r.Ordinal, r.Weekday, hh, mn, end)
		}
		// date mode
		days := append([]int(nil), r.Days...)
		if len(days) == 0 && r.Day != 0 {
			days = []int{r.Day}
		}
		sort.Ints(days)
		return nextMonthlyByDate(start, ref, iv, days, hh, mn, end)
	case "year":
		hh, mn := parseHHMM(r.TimeHHMM)
		months := append([]int(nil), r.Months...)
		if len(months) == 0 && r.Month != 0 {
			months = []int{r.Month}
		}
		sort.Ints(months)
		if r.Mode == "ordinal" {
			return nextYearlyOrdinal(start, ref, max(1, r.Interval), months, r.Ordinal, r.Weekday, hh, mn, end)
		}
		return nextYearlyByDate(start, ref, max(1, r.Interval), months, r.Day, hh, mn, end)
	default:
		return time.Time{}, errors.New("unknown structured freq")
	}
}

// Helpers
func parseHHMM(s string) (int, int) {
	if len(s) >= 4 {
		if len(s) == 5 && s[2] == ':' {
			// 00:00
			h := atoiSafe(s[0:2])
			m := atoiSafe(s[3:5])
			return clamp(h, 0, 23), clamp(m, 0, 59)
		}
	}
	return 0, 0
}

func atoiSafe(s string) int {
	v := 0
	for i := 0; i < len(s); i++ {
		c := s[i]
		if c < '0' || c > '9' {
			return v
		}
		v = v*10 + int(c-'0')
	}
	return v
}

func clamp(v, lo, hi int) int {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func nextEveryNMinutes(anchor, ref time.Time, interval int, end time.Time) (time.Time, error) {
	var zero time.Time
	anc := anchor.Truncate(time.Minute)
	cur := ref.Truncate(time.Minute)
	if cur.Before(anc) {
		cur = anc
	}
	deltaMin := int(cur.Sub(anc).Minutes())
	if deltaMin < 0 {
		deltaMin = 0
	}
	rem := deltaMin % interval
	if rem != 0 {
		cur = cur.Add(time.Duration(interval-rem) * time.Minute)
	}
	if !end.IsZero() && cur.After(end) {
		return zero, ErrNoNextRun
	}
	return cur, nil
}

func nextEveryNHours(anchor, ref time.Time, interval int, minuteMark int, end time.Time) (time.Time, error) {
	var zero time.Time
	cur := time.Date(ref.Year(), ref.Month(), ref.Day(), ref.Hour(), minuteMark, 0, 0, ref.Location())
	if ref.Minute() > minuteMark || (ref.Minute() == minuteMark && ref.Second() > 0) {
		cur = cur.Add(time.Hour)
	}
	anc := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), anchor.Hour(), minuteMark, 0, 0, anchor.Location())
	if cur.Before(anc) {
		cur = anc
	}
	deltaH := int(cur.Sub(anc).Hours())
	if deltaH < 0 {
		deltaH = 0
	}
	rem := deltaH % interval
	if rem != 0 {
		cur = cur.Add(time.Duration(interval-rem) * time.Hour)
	}
	if !end.IsZero() && cur.After(end) {
		return zero, ErrNoNextRun
	}
	return cur, nil
}

func nextEveryNDays(anchor, ref time.Time, interval, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	cur := time.Date(ref.Year(), ref.Month(), ref.Day(), hh, mn, 0, 0, ref.Location())
	if ref.After(cur) { // strictly after, not inclusive
		cur = cur.Add(24 * time.Hour)
	}
	anc := time.Date(anchor.Year(), anchor.Month(), anchor.Day(), hh, mn, 0, 0, anchor.Location())
	if cur.Before(anc) {
		cur = anc
	}
	deltaD := int(cur.Sub(anc).Hours() / 24)
	if deltaD < 0 {
		deltaD = 0
	}
	rem := deltaD % interval
	if rem != 0 {
		cur = cur.Add(time.Duration(interval-rem) * 24 * time.Hour)
	}
	if !end.IsZero() && cur.After(end) {
		return zero, ErrNoNextRun
	}
	return cur, nil
}

func nextEveryNWeeks(start, ref time.Time, interval int, weekdays []int, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	if len(weekdays) == 0 { // any day of week
		weekdays = []int{0, 1, 2, 3, 4, 5, 6}
	}
	// Calculate week index relative to start (weeks since a Sunday-based week containing start)
	anchorWeekStart := weekStart(start)
	cur := ref
	if cur.Before(start) {
		cur = start
	}

	for i := 0; i < 4000; i++ { // safety cap ~76 years
		curWeekStart := weekStart(cur)
		weeksSince := int(curWeekStart.Sub(anchorWeekStart).Hours() / (24 * 7))
		if weeksSince < 0 {
			weeksSince = 0
		}
		if weeksSince%interval != 0 {
			// move to next aligned week
			offset := interval - (weeksSince % interval)
			cur = curWeekStart.Add(time.Duration(offset*7) * 24 * time.Hour)
		}
		// within week, find earliest weekday >= cur
		weekBase := weekStart(cur)
		for _, wd := range weekdays {
			d := time.Date(weekBase.Year(), weekBase.Month(), weekBase.Day(), hh, mn, 0, 0, weekBase.Location())
			d = d.Add(time.Duration(wd) * 24 * time.Hour)
			if !d.Before(cur.Truncate(time.Minute)) {
				if !end.IsZero() && d.After(end) {
					return zero, ErrNoNextRun
				}
				return d, nil
			}
		}
		// go to next aligned week
		cur = weekBase.Add(time.Duration(interval*7) * 24 * time.Hour)
		if !end.IsZero() && cur.After(end) {
			return zero, ErrNoNextRun
		}
	}
	return zero, errors.New("no candidate found (week)")
}

func weekStart(t time.Time) time.Time {
	// Week starts Sunday (0)
	wd := int(t.Weekday()) // 0..6
	d := time.Date(t.Year(), t.Month(), t.Day(), 0, 0, 0, 0, t.Location())
	return d.Add(time.Duration(-wd) * 24 * time.Hour)
}

func daysIn(m time.Month, year int) int {
	return time.Date(year, m+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func nextMonthlyByDate(start, ref time.Time, interval int, days []int, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	if len(days) == 0 {
		return zero, errors.New("monthly date mode requires days")
	}
	cur := ref
	if cur.Before(start) {
		cur = start
	}

	// iterate months aligned to interval
	year, month := cur.Year(), cur.Month()
	// compute month delta from start
	deltaMonths := (year-start.Year())*12 + int(month-start.Month())
	rem := mod(deltaMonths, interval)
	if rem != 0 {
		add := interval - rem
		cur = time.Date(start.Year(), start.Month(), 1, hh, mn, 0, 0, start.Location()).AddDate(0, deltaMonths+add, 0)
	} else {
		cur = time.Date(year, month, 1, hh, mn, 0, 0, cur.Location())
	}

	for i := 0; i < 1200; i++ { // 100 years
		yy, mm := cur.Year(), cur.Month()
		dim := daysIn(mm, yy)
		for _, d := range days {
			if d <= dim {
				cand := time.Date(yy, mm, d, hh, mn, 0, 0, cur.Location())
				if !cand.Before(ref) {
					if !end.IsZero() && cand.After(end) {
						return zero, ErrNoNextRun
					}
					return cand, nil
				}
			}
		}
		// next interval month
		cur = time.Date(cur.Year(), cur.Month(), 1, hh, mn, 0, 0, cur.Location()).AddDate(0, interval, 0)
		if !end.IsZero() && cur.After(end) {
			return zero, ErrNoNextRun
		}
	}
	return zero, errors.New("no candidate found (month date)")
}

func nextMonthlyOrdinal(start, ref time.Time, interval int, ordinal string, weekday any, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	cur := ref
	if cur.Before(start) {
		cur = start
	}
	year, month := cur.Year(), cur.Month()
	deltaMonths := (year-start.Year())*12 + int(month-start.Month())
	rem := mod(deltaMonths, interval)
	if rem != 0 {
		add := interval - rem
		cur = time.Date(start.Year(), start.Month(), 1, hh, mn, 0, 0, start.Location()).AddDate(0, deltaMonths+add, 0)
	} else {
		cur = time.Date(year, month, 1, hh, mn, 0, 0, cur.Location())
	}

	for i := 0; i < 1200; i++ {
		cand := nthInMonth(cur.Year(), cur.Month(), ordinal, weekday, hh, mn, cur.Location())
		if !cand.IsZero() && !cand.Before(ref) {
			if !end.IsZero() && cand.After(end) {
				return zero, ErrNoNextRun
			}
			return cand, nil
		}
		cur = time.Date(cur.Year(), cur.Month(), 1, hh, mn, 0, 0, cur.Location()).AddDate(0, interval, 0)
		if !end.IsZero() && cur.After(end) {
			return zero, ErrNoNextRun
		}
	}
	return zero, errors.New("no candidate found (month ordinal)")
}

func nextYearlyByDate(start, ref time.Time, interval int, months []int, day, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	if len(months) == 0 {
		return zero, errors.New("yearly date mode requires months")
	}
	cur := ref
	if cur.Before(start) {
		cur = start
	}
	// align to interval years from start
	deltaYears := cur.Year() - start.Year()
	remY := mod(deltaYears, interval)
	if remY != 0 {
		cur = time.Date(start.Year()+deltaYears+(interval-remY), time.January, 1, hh, mn, 0, 0, start.Location())
	}
	for i := 0; i < 400; i++ { // 400-year cycle safe
		for _, m := range months {
			mm := time.Month(clamp(m, 1, 12))
			dim := daysIn(mm, cur.Year())
			dd := clamp(day, 1, dim)
			cand := time.Date(cur.Year(), mm, dd, hh, mn, 0, 0, cur.Location())
			if !cand.Before(ref) {
				if !end.IsZero() && cand.After(end) {
					return zero, ErrNoNextRun
				}
				return cand, nil
			}
		}
		cur = time.Date(cur.Year()+interval, time.January, 1, hh, mn, 0, 0, cur.Location())
		if !end.IsZero() && cur.After(end) {
			return zero, ErrNoNextRun
		}
	}
	return zero, errors.New("no candidate found (year date)")
}

func nextYearlyOrdinal(start, ref time.Time, interval int, months []int, ordinal string, weekday any, hh, mn int, end time.Time) (time.Time, error) {
	var zero time.Time
	if len(months) == 0 {
		return zero, errors.New("yearly ordinal mode requires months")
	}
	cur := ref
	if cur.Before(start) {
		cur = start
	}
	deltaYears := cur.Year() - start.Year()
	remY := mod(deltaYears, interval)
	if remY != 0 {
		cur = time.Date(start.Year()+deltaYears+(interval-remY), time.January, 1, hh, mn, 0, 0, start.Location())
	}
	for i := 0; i < 400; i++ {
		for _, m := range months {
			cand := nthInMonth(cur.Year(), time.Month(clamp(m, 1, 12)), ordinal, weekday, hh, mn, cur.Location())
			if !cand.IsZero() && !cand.Before(ref) {
				if !end.IsZero() && cand.After(end) {
					return zero, ErrNoNextRun
				}
				return cand, nil
			}
		}
		cur = time.Date(cur.Year()+interval, time.January, 1, hh, mn, 0, 0, cur.Location())
		if !end.IsZero() && cur.After(end) {
			return zero, ErrNoNextRun
		}
	}
	return zero, errors.New("no candidate found (year ordinal)")
}

func mod(a, b int) int {
	r := a % b
	if r < 0 {
		r += b
	}
	return r
}

func nthInMonth(year int, month time.Month, ordinal string, weekday any, hh, mn int, loc *time.Location) time.Time {
	var zero time.Time
	// Resolve the weekday set
	matcher := func(_ time.Time) bool { return true }
	switch v := weekday.(type) {
	case string:
		switch v {
		case "day":
			matcher = func(time.Time) bool { return true }
		case "weekday":
			matcher = func(d time.Time) bool { wd := d.Weekday(); return wd >= time.Monday && wd <= time.Friday }
		case "weekend":
			matcher = func(d time.Time) bool { wd := d.Weekday(); return wd == time.Saturday || wd == time.Sunday }
		default:
			return zero
		}
	case float64:
		wd := time.Weekday(int(v))
		matcher = func(d time.Time) bool { return d.Weekday() == wd }
	case int:
		wd := time.Weekday(v)
		matcher = func(d time.Time) bool { return d.Weekday() == wd }
	default:
		return zero
	}

	firstOfMonth := time.Date(year, month, 1, hh, mn, 0, 0, loc)
	dim := daysIn(month, year)

	pickIndex := func(idx int) time.Time {
		if idx < 1 {
			return zero
		}
		day := 0
		count := 0
		for d := 1; d <= dim; d++ {
			dt := time.Date(year, month, d, hh, mn, 0, 0, loc)
			if matcher(dt) {
				count++
				if count == idx {
					day = d
					break
				}
			}
		}
		if day == 0 {
			return zero
		}
		return time.Date(year, month, day, hh, mn, 0, 0, loc)
	}

	switch ordinal {
	case "first":
		return pickIndex(1)
	case "second":
		return pickIndex(2)
	case "third":
		return pickIndex(3)
	case "fourth":
		return pickIndex(4)
	case "fifth":
		return pickIndex(5)
	case "last":
		// iterate backwards
		for d := dim; d >= 1; d-- {
			dt := time.Date(year, month, d, hh, mn, 0, 0, loc)
			if matcher(dt) {
				return dt
			}
		}
	case "next_to_last":
		found := 0
		for d := dim; d >= 1; d-- {
			dt := time.Date(year, month, d, hh, mn, 0, 0, loc)
			if matcher(dt) {
				found++
				if found == 2 {
					return dt
				}
			}
		}
	}
	_ = firstOfMonth
	return zero
}
