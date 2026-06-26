package cxrestapi

import (
	"chronix/internal/db"
	jobrunpkg "chronix/internal/jobrun"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func cancelRun(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if runID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid runId")
		return
	}
	jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runID)).First()
	if err != nil || jr == nil || jr.ID == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "run not found")
		return
	}
	if !jobrunpkg.CancelRun(runID) {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "run is not currently running or could not be cancelled")
		return
	}
	restresponse.RestSuccess(c, gin.H{"status": "cancellationRequested", "runId": runID})
}

func rerunRun(c *gin.Context) {
	runID := strings.TrimSpace(c.Param("runId"))
	if runID == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "invalid runId")
		return
	}
	jr, err := db.JobRun.Where(db.JobRun.RunUID.Eq(runID)).First()
	if err != nil || jr == nil || jr.JobID == nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "run not found")
		return
	}
	user := userFromGinContext(c)
	newRunID, err := jobrunpkg.EnqueueJobRun(*jr.JobID, user.ID)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "failed to queue rerun", err)
		return
	}
	restresponse.RestSuccess(c, gin.H{"status": "queued", "runId": newRunID})
}
