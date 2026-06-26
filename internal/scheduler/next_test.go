package scheduler

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"
)

// Utility to build JSON from Go map with proper RFC3339 times
func j(v any) []byte {
	b, _ := json.Marshal(v)
	return b
}

func mustTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

func summary(schedule []byte) string {
	var s map[string]any
	_ = json.Unmarshal(schedule, &s)
	return fmt.Sprintf("schedule=%v", s)
}

type tc struct {
	name  string
	sched any
	from  time.Time
	exp   time.Time
}

func TestNextRunTime_AllPermutations(t *testing.T) {
	ref := time.Date(2025, 11, 10, 4, 21, 0, 0, time.UTC)

	cases := []tc{
		{
			name: "single future",
			sched: map[string]any{
				"kind":  "single",
				"runAt": mustTime("2025-11-10T05:00:00Z").Format(time.RFC3339),
			},
			from: ref,
			exp:  mustTime("2025-11-10T05:00:00Z").Truncate(time.Minute),
		},
		{
			name: "recurring minute every 15",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-10T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "minute",
					"interval": 15,
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2025-11-10T04:30:00Z"),
		},
		{
			name: "recurring hour every 2 at :10",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-10T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":       "hour",
					"interval":   2,
					"minuteMark": 10,
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2025-11-10T06:10:00Z"),
		},
		{
			name: "recurring day every 1 at 08:30",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-09T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "day",
					"interval": 1,
					"time":     "08:30",
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2025-11-10T08:30:00Z"),
		},
		{
			name: "recurring week every 2 on Mon,Wed,Fri at 07:00",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-03T00:00:00Z").Format(time.RFC3339), // Monday week anchor
				"rule": map[string]any{
					"freq":     "week",
					"interval": 2,
					"weekdays": []int{1, 3, 5},
					"time":     "07:00",
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"), // Monday 10th
			exp:  mustTime("2025-11-17T07:00:00Z"),
		},
		{
			name: "monthly by date on 15th and 30th at 09:15",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-10-01T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "month",
					"interval": 1,
					"mode":     "date",
					"days":     []int{15, 30},
					"time":     "09:15",
				},
			},
			from: mustTime("2025-11-10T10:00:00Z"),
			exp:  mustTime("2025-11-15T09:15:00Z"),
		},
		{
			name: "monthly ordinal second weekday at 10:45",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-01T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "month",
					"interval": 1,
					"mode":     "ordinal",
					"ordinal":  "second",
					"weekday":  "weekday",
					"time":     "10:45",
				},
			},
			from: mustTime("2025-11-10T11:00:00Z"),
			exp:  mustTime("2025-12-02T10:45:00Z"), // 2nd weekday in Dec 2025 is Tue Dec 2
		},
		{
			name: "yearly date March and July 5th at 12:00",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-01-01T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "year",
					"interval": 1,
					"mode":     "date",
					"months":   []int{3, 7},
					"day":      5,
					"time":     "12:00",
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2026-03-05T12:00:00Z"),
		},
		{
			name: "yearly ordinal last weekend day in Feb at 06:00 every 2 years",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2024-01-01T00:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "year",
					"interval": 2,
					"mode":     "ordinal",
					"months":   []int{2},
					"ordinal":  "last",
					"weekday":  "weekend",
					"time":     "06:00",
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2026-02-28T06:00:00Z"), // 2026 Feb 28 is Saturday
		},
		{
			name: "cron every 5 minutes",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "cron",
				"startAt": mustTime("2025-11-10T00:00:00Z").Format(time.RFC3339),
				"cron":    "*/5 * * * *",
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  mustTime("2025-11-10T04:25:00Z"),
		},
		{
			name: "endAt boundary yields no next run",
			sched: map[string]any{
				"kind":    "recurring",
				"mode":    "structured",
				"startAt": mustTime("2025-11-10T00:00:00Z").Format(time.RFC3339),
				"endAt":   mustTime("2025-11-10T04:00:00Z").Format(time.RFC3339),
				"rule": map[string]any{
					"freq":     "minute",
					"interval": 10,
				},
			},
			from: mustTime("2025-11-10T04:21:00Z"),
			exp:  time.Time{},
		},
	}

	for _, c := range cases {
		b := j(c.sched)
		act, err := NextRunTime(b, c.from)
		if c.exp.IsZero() {
			if err == nil {
				t.Errorf("%s\n  schedule: %s\n  from: %s\n  got: %s (expected error no next)", c.name, summary(b), c.from.Format(time.RFC3339), act.Format(time.RFC3339))
			}
			continue
		}
		if err != nil {
			t.Errorf("%s\n  schedule: %s\n  from: %s\n  error: %v\n  expected: %s", c.name, summary(b), c.from.Format(time.RFC3339), err, c.exp.Format(time.RFC3339))
			continue
		}
		if !act.Equal(c.exp) {
			t.Errorf("%s\n  schedule: %s\n  from: %s\n  got: %s\n  expected: %s", c.name, summary(b), c.from.Format(time.RFC3339), act.Format(time.RFC3339), c.exp.Format(time.RFC3339))
		}
	}
}
