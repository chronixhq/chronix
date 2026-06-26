package cxrestapi

import (
	"chronix"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/gin-gonic/gin"
)

func helpRouter(app *gin.Engine) {
	app.GET("/help", getHelpMarkdown)
}

func getHelpMarkdown(c *gin.Context) {
	content, err := chronix.HelpMarkdown.ReadFile("docs/help.md")
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Failed to read help documentation", err.Error())
		return
	}

	markdown := string(content)
	// Trim front matter if it exists
	if strings.HasPrefix(markdown, "---") {
		parts := strings.SplitN(markdown, "---", 3)
		if len(parts) == 3 {
			markdown = strings.TrimSpace(parts[2])
		}
	}

	restresponse.RestSuccess(c, gin.H{"markdown": markdown})
}
