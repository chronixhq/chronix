package shellactionsapi

import (
	"chronix/internal/db"
	"chronix/internal/db/models"
	"chronix/internal/execution"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func testShellAction(c *gin.Context) {
	var p shellActionTestPayload
	if err := c.BindJSON(&p); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	conn, err := db.ShellConnection.Where(db.ShellConnection.ID.Eq(p.ShellConnectionID)).First()
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Connection not found")
		return
	}
	if conn.Suspended {
		restresponse.RestErrorRespond(c, restresponse.FailedPrecondition, "Connection is suspended")
		return
	}

	steps := make([]models.ShellActionStep, 0, len(p.Steps))
	for _, s := range p.Steps {
		capBytes := capOutputCaptureBytes(s.OutputCaptureMaxBytes)
		steps = append(steps, models.ShellActionStep{
			StepOrder:             &s.Order,
			Name:                  s.Name,
			RunMode:               &s.RunMode,
			Command:               s.Command,
			ScriptText:            s.ScriptText,
			ShellPath:             &s.ShellPath,
			WorkingDir:            s.WorkingDir,
			TimeoutSeconds:        s.TimeoutSeconds,
			OutputCaptureMaxBytes: &capBytes,
			OutputTruncation:      &s.OutputTruncation,
			OnFailure:             s.OnFailure,
			Expectation:           toMap(s.Expectation),
			OutputCapture:         toMap(s.OutputCapture),
			EnvJSON:               toMap(s.Env),
		})
	}

	results, err := execution.TestShellAction(c.Request.Context(), steps, conn, p.Variables)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Test failed", err.Error())
		return
	}

	restresponse.RestSuccess(c, results)
}

func validateShellScript(c *gin.Context) {
	var body struct {
		ShellPath  string `json:"shellPath"`
		ScriptText string `json:"scriptText"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	body.ShellPath = strings.TrimSpace(body.ShellPath)
	if body.ShellPath == "" {
		body.ShellPath = "/bin/sh"
	}
	ok, msg := shellSyntaxCheck(c.Request.Context(), body.ShellPath, body.ScriptText)
	restresponse.RestSuccess(c, gin.H{"ok": ok, "message": msg})
}

func shellSyntaxCheck(ctx context.Context, shellPath string, script string) (bool, string) {
	dir := os.TempDir()
	f, err := os.CreateTemp(dir, "cx-script-*.sh")
	if err != nil {
		return false, fmt.Sprintf("tempfile: %v", err)
	}
	fname := f.Name()
	_, _ = f.WriteString(script)
	_ = f.Close()
	defer func() { _ = os.Remove(fname) }()

	cmd := exec.CommandContext(ctx, shellPath, "-n", fname)
	out, err := cmd.CombinedOutput()
	if err != nil {
		s := strings.TrimSpace(string(out))
		if s == "" {
			s = err.Error()
		}
		return false, s
	}
	return true, "ok"
}
