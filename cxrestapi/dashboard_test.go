package cxrestapi

import (
	cxuserpkg "chronix/internal/cxuser"
	"chronix/internal/db"
	"chronix/internal/db/models"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

func TestGetDashboardSummary(t *testing.T) {
	// Setup temp DB
	dbFile := "test_dashboard_summary.db"
	defer func() { _ = os.Remove(dbFile) }()

	gormDB, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.SetDefault(gormDB)
	db.DB = gormDB

	// Create tables
	err = gormDB.AutoMigrate(
		&models.Job{},
		&models.Action{},
		&models.JobRun{},
		&models.DbConnection{},
		&models.ShellConnection{},
		&models.WebtaskConnection{},
		&models.Agent{},
		&models.AgentRegistrationRequest{},
		&models.UserActivity{},
		&models.CxUser{},
	)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()

	// Seed data
	// DB connections: 1 OK, 1 Error, 1 Unknown (NULL status)
	if err := db.DbConnection.Create(&models.DbConnection{Name: "db-ok", LastStatus: utilities.Ptr("ok"), CreatedAt: &now, UpdatedAt: &now, Driver: "sqlite", Dsn: utilities.Ptr("test"), AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: utilities.Ptr(int64(0)), Enabled: utilities.Ptr(true), Suspended: utilities.Ptr(false)}); err != nil {
		t.Fatalf("failed to create db-ok: %v", err)
	}
	if err := db.DbConnection.Create(&models.DbConnection{Name: "db-error", LastStatus: utilities.Ptr("error"), CreatedAt: &now, UpdatedAt: &now, Driver: "sqlite", Dsn: utilities.Ptr("test"), AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: utilities.Ptr(int64(0)), Enabled: utilities.Ptr(true), Suspended: utilities.Ptr(false)}); err != nil {
		t.Fatalf("failed to create db-error: %v", err)
	}
	if err := db.DbConnection.Create(&models.DbConnection{Name: "db-unknown", LastStatus: nil, CreatedAt: &now, UpdatedAt: &now, Driver: "sqlite", Dsn: utilities.Ptr("test"), AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: utilities.Ptr(int64(0)), Enabled: utilities.Ptr(true), Suspended: utilities.Ptr(false)}); err != nil {
		t.Fatalf("failed to create db-unknown: %v", err)
	}

	// Shell connections: 1 OK, 2 Unknown (empty status)
	if err := db.ShellConnection.Create(&models.ShellConnection{Name: "shell-ok", LastStatus: utilities.Ptr("ok"), CreatedAt: now, UpdatedAt: now, Mode: "localhost", AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: 0, Enabled: true, Suspended: false}); err != nil {
		t.Fatalf("failed to create shell-ok: %v", err)
	}
	if err := db.ShellConnection.Create(&models.ShellConnection{Name: "shell-un1", LastStatus: utilities.Ptr(""), CreatedAt: now, UpdatedAt: now, Mode: "localhost", AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: 0, Enabled: true, Suspended: false}); err != nil {
		t.Fatalf("failed to create shell-un1: %v", err)
	}
	if err := db.ShellConnection.Create(&models.ShellConnection{Name: "shell-un2", LastStatus: utilities.Ptr("something-else"), CreatedAt: now, UpdatedAt: now, Mode: "localhost", AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: 0, Enabled: true, Suspended: false}); err != nil {
		t.Fatalf("failed to create shell-un2: %v", err)
	}

	// Webtask connections: 1 Error, 1 Unknown
	if err := db.WebtaskConnection.Create(&models.WebtaskConnection{Name: "web-err", LastStatus: utilities.Ptr("error"), CreatedAt: &now, UpdatedAt: &now, AuthType: "none", AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: utilities.Ptr(int64(0)), Enabled: utilities.Ptr(true), Suspended: utilities.Ptr(false)}); err != nil {
		t.Fatalf("failed to create web-err: %v", err)
	}
	if err := db.WebtaskConnection.Create(&models.WebtaskConnection{Name: "web-un", LastStatus: nil, CreatedAt: &now, UpdatedAt: &now, AuthType: "none", AutoCheckEnabled: utilities.Ptr(int64(0)), AutoCheckIntervalSeconds: utilities.Ptr(int64(0)), Enabled: utilities.Ptr(true), Suspended: utilities.Ptr(false)}); err != nil {
		t.Fatalf("failed to create web-un: %v", err)
	}

	// Total: 3 (DB) + 3 (Shell) + 2 (Webtask) = 8
	// OK: 1 (DB) + 1 (Shell) + 0 (Webtask) = 2
	// Error: 1 (DB) + 0 (Shell) + 1 (Webtask) = 2
	// Unknown: 1 (DB) + 2 (Shell) + 1 (Webtask) = 4

	// Seed a job with a simple recurring schedule
	// Schedule: every minute
	schedule := map[string]any{
		"kind":    "recurring",
		"mode":    "cron",
		"cron":    "* * * * *",
		"startAt": now.Add(-10 * time.Minute).Format(time.RFC3339),
	}
	enabled := true
	job := &models.Job{
		Name:         "Test Job",
		Enabled:      &enabled,
		ScheduleJSON: datatypes.JSONMap(schedule),
		CreatedAt:    now,
		UpdatedAt:    now,
		ActionID:     1, // dummy
		TargetKind:   "dummy",
	}
	if err := db.Job.Create(job); err != nil {
		t.Fatalf("failed to create job: %v", err)
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user", &cxuserpkg.CxUser{
			CxUser: models.CxUser{
				ID:   1,
				Name: "Test User",
			},
		})
		c.Next()
	})
	dashboardRouter(r)

	req, _ := http.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Stats    dashboardStats      `json:"stats"`
		Upcoming []dashboardUpcoming `json:"upcoming"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Upcoming) == 0 {
		t.Error("expected at least one upcoming job, got none")
	}
	if len(resp.Upcoming) > 1 {
		for i := 1; i < len(resp.Upcoming); i++ {
			if resp.Upcoming[i].When.Before(resp.Upcoming[i-1].When) {
				t.Fatalf("expected upcoming jobs sorted ascending, got %v before %v", resp.Upcoming[i], resp.Upcoming[i-1])
			}
		}
	}

	s := resp.Stats
	if s.ConnectionsTotal != 8 {
		t.Errorf("expected 8 total connections, got %d", s.ConnectionsTotal)
	}
	if s.ConnectionsOk != 2 {
		t.Errorf("expected 2 ok connections, got %d", s.ConnectionsOk)
	}
	if s.ConnectionsError != 2 {
		t.Errorf("expected 2 error connections, got %d", s.ConnectionsError)
	}
	if s.ConnectionsUnknown != 4 {
		t.Errorf("expected 4 unknown connections, got %d", s.ConnectionsUnknown)
	}
}

func TestGetDashboardSummary_UpcomingCappedAndSorted(t *testing.T) {
	dbFile := "test_dashboard_upcoming.db"
	defer func() { _ = os.Remove(dbFile) }()

	gormDB, err := gorm.Open(sqlite.Open(dbFile), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	db.SetDefault(gormDB)
	db.DB = gormDB

	err = gormDB.AutoMigrate(
		&models.Job{},
		&models.Action{},
		&models.JobRun{},
		&models.DbConnection{},
		&models.ShellConnection{},
		&models.WebtaskConnection{},
		&models.Agent{},
		&models.AgentRegistrationRequest{},
		&models.UserActivity{},
		&models.CxUser{},
	)
	if err != nil {
		t.Fatal(err)
	}

	now := time.Now().UTC()
	enabled := true
	cronSpecs := []string{
		"*/6 * * * *",
		"*/5 * * * *",
		"*/4 * * * *",
		"*/3 * * * *",
		"*/2 * * * *",
		"*/1 * * * *",
	}

	for i, cronSpec := range cronSpecs {
		schedule := datatypes.JSONMap(map[string]any{
			"kind":    "recurring",
			"mode":    "cron",
			"cron":    cronSpec,
			"startAt": now.Add(-10 * time.Minute).Format(time.RFC3339),
		})
		job := &models.Job{
			Name:         "Job " + string(rune('A'+i)),
			Enabled:      &enabled,
			ScheduleJSON: schedule,
			CreatedAt:    now,
			UpdatedAt:    now,
			ActionID:     int64(i + 1),
			TargetKind:   "dummy",
		}
		if err := db.Job.Create(job); err != nil {
			t.Fatalf("failed to create job %d: %v", i, err)
		}
	}

	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("user", &cxuserpkg.CxUser{CxUser: models.CxUser{ID: 1, Name: "Test User"}})
		c.Next()
	})
	dashboardRouter(r)

	req, _ := http.NewRequest(http.MethodGet, "/dashboard/summary", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status OK, got %d, body: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Upcoming []dashboardUpcoming `json:"upcoming"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatal(err)
	}

	if len(resp.Upcoming) != 5 {
		t.Fatalf("expected 5 upcoming jobs, got %d", len(resp.Upcoming))
	}
	for i := 1; i < len(resp.Upcoming); i++ {
		if resp.Upcoming[i].When.Before(resp.Upcoming[i-1].When) {
			t.Fatalf("expected upcoming jobs sorted ascending, got %v before %v", resp.Upcoming[i], resp.Upcoming[i-1])
		}
	}
}
