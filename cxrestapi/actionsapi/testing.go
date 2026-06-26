package actionsapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/execution"
	"chronix/internal/sqlsyntax"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func testAction(c *gin.Context) {
	var p actionTestPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	conn, err := db.DbConnection.Where(db.DbConnection.ID.Eq(p.ConnectionID)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Connection not found")
		return
	}
	if conn.Suspended != nil && *conn.Suspended {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Connection is suspended")
		return
	}

	steps := make([]models.ActionStep, 0, len(p.Steps))
	for _, s := range p.Steps {
		steps = append(steps, models.ActionStep{
			StepOrder:      s.Order,
			Name:           s.Name,
			SqlText:        s.SQLText,
			TimeoutSeconds: s.TimeoutSeconds,
			OnFailure:      s.OnFailure,
			Expectation:    toMap(s.Expectation),
			OutputCapture:  toMap(s.OutputCapture),
		})
	}

	results, err := execution.TestDatabaseAction(c.Request.Context(), steps, conn, p.Variables)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Test failed", err.Error())
		return
	}

	restresponse.RestSuccess(c, results)
}

func validateActionStep(c *gin.Context) {
	var body struct {
		Dialect   string         `json:"dialect"`
		SQLText   string         `json:"sqlText"`
		Variables map[string]any `json:"variables"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	passed, failMessage := sqlsyntax.ValidateSQL(sqlsyntax.Dialect(body.Dialect), body.SQLText, body.Variables)
	issues := make([]gin.H, 0)
	if !passed {
		issues = append(issues, gin.H{"code": "sql.syntax", "message": failMessage})
	}
	restresponse.RestSuccess(c, gin.H{"ok": passed, "issues": issues})
}
