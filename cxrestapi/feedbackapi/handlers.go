package feedbackapi

import (
	activitypkg "chronix/internal/activity"
	"chronix/internal/db"
	"chronix/internal/db/models"
	serverruntime "chronix/internal/serverruntime"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"chronix/cxrestapi/apiutil"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func postBugReport(c *gin.Context) {
	handleFeedbackSubmission(c, "bug")
}

func postFeatureRequest(c *gin.Context) {
	handleFeedbackSubmission(c, "feature")
}

func handleFeedbackSubmission(c *gin.Context, kind string) {
	user := apiutil.UserFromGinContext(c)

	summary := c.PostForm("summary")
	description := c.PostForm("description")

	if summary == "" || description == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Summary and description are required")
		return
	}

	var reportID int64
	now := time.Now()

	if kind == "bug" {
		report := models.BugReport{
			Summary:     summary,
			Description: description,
			UserID:      user.ID,
			CreatedAt:   now,
			Status:      "open",
		}
		if err := db.BugReport.Create(&report); err != nil {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving bug report", err.Error())
			return
		}
		reportID = utilities.PtrVal(report.ID)
	} else {
		request := models.FeatureRequest{
			Summary:     summary,
			Description: description,
			UserID:      user.ID,
			CreatedAt:   now,
			Status:      "open",
		}
		if err := db.FeatureRequest.Create(&request); err != nil {
			restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving feature request", err.Error())
			return
		}
		reportID = utilities.PtrVal(request.ID)
	}

	if err := saveAttachments(c, kind, reportID); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving attachments", err.Error())
		return
	}

	action := "Bug Report Submitted"
	if kind == "feature" {
		action = "Feature Request Submitted"
	}
	_ = activitypkg.RecordUserActivity(user.ID, action, summary, c.ClientIP(), c.Request.UserAgent())

	restresponse.RestSuccess(c, gin.H{"id": reportID})
}

func saveAttachments(c *gin.Context, kind string, reportID int64) error {
	form, err := c.MultipartForm()
	if err != nil {
		return nil
	}
	files := form.File["attachments"]
	if len(files) == 0 {
		return nil
	}

	feedbackDir := filepath.Join(serverruntime.DataDir, "feedback")
	if err := os.MkdirAll(feedbackDir, 0755); err != nil {
		return err
	}

	now := time.Now()
	for _, file := range files {
		ext := filepath.Ext(file.Filename)
		uniqueName := fmt.Sprintf("%s_%d_%d%s", kind, reportID, time.Now().UnixNano(), ext)
		dst := filepath.Join(feedbackDir, uniqueName)

		if err := c.SaveUploadedFile(file, dst); err != nil {
			slog.Error("Error saving uploaded file", "error", err, "filename", file.Filename)
			continue
		}

		attachment := models.FeedbackAttachment{
			FileName:    file.Filename,
			FilePath:    uniqueName,
			ContentType: file.Header.Get("Content-Type"),
			FileSize:    file.Size,
			CreatedAt:   now,
		}
		if kind == "bug" {
			attachment.BugReportID = &reportID
		} else {
			attachment.FeatureRequestID = &reportID
		}

		if err := db.FeedbackAttachment.Create(&attachment); err != nil {
			slog.Error("Error saving attachment to DB", "error", err, "filename", file.Filename)
		}
	}
	return nil
}

func getBugReports(c *gin.Context) {
	reports, err := db.BugReport.Preload(db.BugReport.FeedbackAttachments).Order(db.BugReport.CreatedAt.Desc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching bug reports", err.Error())
		return
	}
	restresponse.RestSuccess(c, reports)
}

func getFeatureRequests(c *gin.Context) {
	requests, err := db.FeatureRequest.Preload(db.FeatureRequest.FeedbackAttachments).Order(db.FeatureRequest.CreatedAt.Desc()).Find()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error fetching feature requests", err.Error())
		return
	}
	restresponse.RestSuccess(c, requests)
}

func getFeedbackAttachment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid attachment ID")
		return
	}

	attachment, err := db.FeedbackAttachment.Where(db.FeedbackAttachment.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "Attachment not found")
		return
	}

	feedbackDir := filepath.Join(serverruntime.DataDir, "feedback")
	filePath := filepath.Join(feedbackDir, attachment.FilePath)

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "File not found on disk")
		return
	}

	c.Header("Content-Disposition", fmt.Sprintf("inline; filename=%q", attachment.FileName))
	c.File(filePath)
}

func patchBugReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var input struct {
		Summary     *string `json:"summary"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid input", err.Error())
		return
	}

	report, err := db.BugReport.Where(db.BugReport.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "Bug report not found")
		return
	}

	if input.Summary != nil {
		report.Summary = *input.Summary
	}
	if input.Description != nil {
		report.Description = *input.Description
	}
	if input.Status != nil {
		report.Status = *input.Status
	}

	if _, err := db.BugReport.Where(db.BugReport.ID.Eq(id)).Updates(report); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error updating bug report", err.Error())
		return
	}

	restresponse.RestSuccess(c, report)
}

func deleteBugReport(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	attachments, _ := db.FeedbackAttachment.Where(db.FeedbackAttachment.BugReportID.Eq(id)).Find()
	feedbackDir := filepath.Join(serverruntime.DataDir, "feedback")
	for _, att := range attachments {
		_ = os.Remove(filepath.Join(feedbackDir, att.FilePath))
	}

	_, _ = db.FeedbackAttachment.Where(db.FeedbackAttachment.BugReportID.Eq(id)).Delete()

	if _, err := db.BugReport.Where(db.BugReport.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error deleting bug report", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}

func patchFeatureRequest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	var input struct {
		Summary     *string `json:"summary"`
		Description *string `json:"description"`
		Status      *string `json:"status"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid input", err.Error())
		return
	}

	request, err := db.FeatureRequest.Where(db.FeatureRequest.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "Feature request not found")
		return
	}

	if input.Summary != nil {
		request.Summary = *input.Summary
	}
	if input.Description != nil {
		request.Description = *input.Description
	}
	if input.Status != nil {
		request.Status = *input.Status
	}

	if _, err := db.FeatureRequest.Where(db.FeatureRequest.ID.Eq(id)).Updates(request); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error updating feature request", err.Error())
		return
	}

	restresponse.RestSuccess(c, request)
}

func deleteFeatureRequest(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	attachments, _ := db.FeedbackAttachment.Where(db.FeedbackAttachment.FeatureRequestID.Eq(id)).Find()
	feedbackDir := filepath.Join(serverruntime.DataDir, "feedback")
	for _, att := range attachments {
		_ = os.Remove(filepath.Join(feedbackDir, att.FilePath))
	}

	_, _ = db.FeedbackAttachment.Where(db.FeedbackAttachment.FeatureRequestID.Eq(id)).Delete()

	if _, err := db.FeatureRequest.Where(db.FeatureRequest.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error deleting feature request", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}

func postBugReportAttachments(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := saveAttachments(c, "bug", id); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving attachments", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}

func postFeatureRequestAttachments(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)

	if err := saveAttachments(c, "feature", id); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving attachments", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}

func deleteFeedbackAttachment(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid attachment ID")
		return
	}

	attachment, err := db.FeedbackAttachment.Where(db.FeedbackAttachment.ID.Eq(id)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.NotFound, "Attachment not found")
		return
	}

	feedbackDir := filepath.Join(serverruntime.DataDir, "feedback")
	filePath := filepath.Join(feedbackDir, attachment.FilePath)
	_ = os.Remove(filePath)

	if _, err := db.FeedbackAttachment.Where(db.FeedbackAttachment.ID.Eq(id)).Delete(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error deleting attachment from DB", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}
