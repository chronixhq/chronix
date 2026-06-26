package cxrestapi

import (
	"chronix/internal/agentmux"
	"chronix/internal/db"
	"chronix/internal/scheduler"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

// dashboardRouter registers dashboard-related endpoints.
func dashboardRouter(app *gin.Engine) {
	app.GET("/dashboard/summary", getDashboardSummary)
}

type dashboardStats struct {
	Jobs               int64 `json:"jobs"`
	Running            int64 `json:"running"`
	Actions            int64 `json:"actions"`
	ConnectionsOk      int64 `json:"connectionsOk"`
	ConnectionsError   int64 `json:"connectionsError"`
	ConnectionsUnknown int64 `json:"connectionsUnknown"`
	ConnectionsTotal   int64 `json:"connectionsTotal"`
	AgentsKnown        int64 `json:"agentsKnown"`
	AgentsOnline       int64 `json:"agentsOnline"`
	AgentsOffline      int64 `json:"agentsOffline"`
	AgentsPending      int64 `json:"agentsPending"`
}

type dashboardUpcoming struct {
	ID     string    `json:"id"`
	Name   string    `json:"name"`
	When   time.Time `json:"when"`
	Status string    `json:"status"`
}

type dashboardActivity struct {
	ID   string    `json:"id"`
	When time.Time `json:"when"`
	Text string    `json:"text"`
}

func getDashboardSummary(c *gin.Context) {
	var stats dashboardStats

	stats.Jobs, _ = db.Job.Count()
	stats.Actions, _ = db.Action.Count()
	stats.Running, _ = db.JobRun.Where(db.JobRun.Status.Eq("running")).Count()

	if cnt, err := db.DbConnection.Count(); err == nil {
		stats.ConnectionsTotal += cnt
		ok, _ := db.DbConnection.Where(db.DbConnection.LastStatus.Eq("ok")).Count()
		stats.ConnectionsOk += ok
		errs, _ := db.DbConnection.Where(db.DbConnection.LastStatus.Eq("error")).Count()
		stats.ConnectionsError += errs
	}
	if cnt, err := db.ShellConnection.Count(); err == nil {
		stats.ConnectionsTotal += cnt
		ok, _ := db.ShellConnection.Where(db.ShellConnection.LastStatus.Eq("ok")).Count()
		stats.ConnectionsOk += ok
		errs, _ := db.ShellConnection.Where(db.ShellConnection.LastStatus.Eq("error")).Count()
		stats.ConnectionsError += errs
	}
	if cnt, err := db.WebtaskConnection.Count(); err == nil {
		stats.ConnectionsTotal += cnt
		ok, _ := db.WebtaskConnection.Where(db.WebtaskConnection.LastStatus.Eq("ok")).Count()
		stats.ConnectionsOk += ok
		errs, _ := db.WebtaskConnection.Where(db.WebtaskConnection.LastStatus.Eq("error")).Count()
		stats.ConnectionsError += errs
	}
	stats.ConnectionsUnknown = stats.ConnectionsTotal - stats.ConnectionsOk - stats.ConnectionsError
	stats.AgentsKnown, _ = db.Agent.Count()

	// Online agents: union of currently connected via agentmux + recently seen in DB (within threshold)
	onlineThreshold := 2 * time.Minute
	// Build set from live manager
	liveSet := map[string]struct{}{}
	for _, id := range agentmux.DefaultManager.List() {
		if id == "" {
			continue
		}
		liveSet[id] = struct{}{}
	}
	cutoff := time.Now().Add(-onlineThreshold)
	if recent, err := db.Agent.Select(db.Agent.UUID).Where(db.Agent.LastSeenAt.Gte(cutoff)).Find(); err == nil {
		for _, r := range recent {
			if r.UUID == "" {
				continue
			}
			liveSet[r.UUID] = struct{}{}
		}
	}
	stats.AgentsOnline = int64(len(liveSet))
	if stats.AgentsKnown > stats.AgentsOnline {
		stats.AgentsOffline = stats.AgentsKnown - stats.AgentsOnline
	} else {
		stats.AgentsOffline = 0
	}
	// Pending registrations (not expired)
	now := time.Now().UTC()
	stats.AgentsPending, _ = db.AgentRegistrationRequest.Where(db.AgentRegistrationRequest.Status.Eq("pending"), db.AgentRegistrationRequest.ExpiresAt.Gt(now)).Count()

	// Upcoming jobs: compute next run time for enabled jobs with schedule_json
	// Handle minute-collision immediately after a run finishes (same as jobs.list):
	// if next equals last at minute precision, advance the reference by 1 minute to find the subsequent occurrence.
	upcoming := make([]dashboardUpcoming, 0, 5)
	{
		if rows, err := db.Job.Select(db.Job.ID, db.Job.Name, db.Job.Enabled, db.Job.ScheduleJSON).Where(db.Job.Enabled.Is(true)).Find(); err == nil {
			// Build a map of last run timestamps per job to resolve immediate post-run collisions
			lastByJob := map[int64]*time.Time{}
			{
				ids := make([]int64, 0, len(rows))
				for _, r := range rows {
					if r.ID != nil {
						ids = append(ids, *r.ID)
					}
				}
				if len(ids) > 0 {
					type lastRow struct {
						JobID      int64      `json:"job_id"`
						QueuedAt   *time.Time `json:"queued_at"`
						StartedAt  *time.Time `json:"started_at"`
						FinishedAt *time.Time `json:"finished_at"`
					}
					var lr []lastRow
					if err := db.JobRun.
						Select(db.JobRun.JobID, db.JobRun.QueuedAt, db.JobRun.StartedAt, db.JobRun.FinishedAt).
						Where(db.JobRun.JobID.In(ids...)).
						Order(db.JobRun.StartedAt.Desc(), db.JobRun.QueuedAt.Desc(), db.JobRun.FinishedAt.Desc()).
						Scan(&lr); err == nil {
						for _, x := range lr {
							if _, ok := lastByJob[x.JobID]; ok {
								continue
							}
							// Prefer finished, then started, then queued
							if x.FinishedAt != nil {
								lastByJob[x.JobID] = x.FinishedAt
							} else if x.StartedAt != nil {
								lastByJob[x.JobID] = x.StartedAt
							} else if x.QueuedAt != nil {
								lastByJob[x.JobID] = x.QueuedAt
							}
						}
					}
				}
			}

			for _, r := range rows {
				if r.ID == nil || len(r.ScheduleJSON) == 0 || (r.Enabled != nil && !*r.Enabled) {
					continue
				}
				sjson, err := json.Marshal(r.ScheduleJSON)
				if err != nil {
					continue
				}
				var nextAt time.Time
				if t, err := scheduler.NextRunTime(sjson); err == nil && !t.IsZero() {
					nextAt = t
				} else {
					continue
				}
				// If nextAt matches the last run timestamp at minute precision, compute the subsequent occurrence
				if last := lastByJob[*r.ID]; last != nil {
					if nextAt.Truncate(time.Minute).Equal(last.Truncate(time.Minute)) {
						if t2, err := scheduler.NextRunTime(sjson, last.Add(time.Minute)); err == nil && !t2.IsZero() {
							nextAt = t2
						} else {
							// No further occurrences; skip adding to upcoming
							continue
						}
					}
				}
				upcoming = append(upcoming, dashboardUpcoming{
					ID:     itoa64(*r.ID),
					Name:   r.Name,
					When:   nextAt,
					Status: "scheduled",
				})
			}
		}
	}
	sort.Slice(upcoming, func(i, j int) bool { return upcoming[i].When.Before(upcoming[j].When) })
	if len(upcoming) > 5 {
		upcoming = upcoming[:5]
	}

	// Recent activity: use the same unified aggregation as the Activity page for consistency
	// Additionally, filter out user login/logout events from the Dashboard view.
	activity := make([]dashboardActivity, 0, 5)
	{
		uc := userFromGinContext(c)
		// Fetch a slightly larger pool to account for filtering
		merged := aggregateUnifiedActivity(*uc, 50)
		// Filter out login/logout actions (case-insensitive exact match)
		filtered := make([]unifiedActivityItem, 0, len(merged))
		for _, it := range merged {
			a := strings.ToLower(strings.TrimSpace(it.Action))
			if a == "login" || a == "logout" {
				continue
			}
			filtered = append(filtered, it)
		}
		// Cap to 5 items for the dashboard card
		maxItems := 5
		if len(filtered) < maxItems {
			maxItems = len(filtered)
		}
		for i := 0; i < maxItems; i++ {
			it := filtered[i]
			// Parse time
			t, err := time.Parse(time.RFC3339, it.When)
			if err != nil {
				t = time.Now().UTC()
			}
			// Prefer ID as-is; it already carries a prefix (run:/ua:)
			activity = append(activity, dashboardActivity{ID: it.ID, When: t, Text: it.Action})
		}
	}

	resp := gin.H{
		"stats":    stats,
		"upcoming": upcoming,
		"activity": activity,
	}
	restresponse.RestSuccess(c, resp)
}

func itoa64(v int64) string { return fmt.Sprintf("%d", v) }
