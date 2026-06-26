package cxrestapi

import (
	cxsettingspkg "chronix/internal/cxsettings"
	"chronix/internal/notifier"
	"fmt"
	"strconv"
	"strings"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

func getEmailSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings
	resp := gin.H{
		"smtpHost": utilities.PtrVal(s.SMTPHost),
		"smtpPort": func() string {
			if s.SMTPPort != nil {
				return fmt.Sprintf("%d", *s.SMTPPort)
			}
			return "587"
		}(),
		"secure":       utilities.PtrVal(s.SMTPSecure),
		"fromName":     utilities.PtrVal(s.SMTPFromName),
		"fromEmail":    utilities.PtrVal(s.SMTPFromEmail),
		"smtpLogin":    utilities.PtrVal(s.SMTPLogin),
		"smtpPassword": redactIfSet(s.SMTPPassword),
	}
	restresponse.RestSuccess(c, resp)
}

func putEmailSettings(c *gin.Context) {
	var body struct {
		SMTPHost     string `json:"smtpHost"`
		SMTPPort     string `json:"smtpPort"`
		Secure       string `json:"secure"`
		FromName     string `json:"fromName"`
		FromEmail    string `json:"fromEmail"`
		SMTPLogin    string `json:"smtpLogin"`
		SMTPPassword string `json:"smtpPassword"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.SMTPHost = utilities.Ptr(strings.TrimSpace(body.SMTPHost))
	if p := strings.TrimSpace(body.SMTPPort); p != "" {
		if n, err := strconv.Atoi(p); err == nil {
			cxsettingspkg.CxSettings.SMTPPort = utilities.Ptr(int64(n))
		}
	}
	cxsettingspkg.CxSettings.SMTPSecure = utilities.Ptr(strings.TrimSpace(body.Secure))
	cxsettingspkg.CxSettings.SMTPFromName = utilities.Ptr(strings.TrimSpace(body.FromName))
	cxsettingspkg.CxSettings.SMTPFromEmail = utilities.Ptr(strings.TrimSpace(body.FromEmail))
	cxsettingspkg.CxSettings.SMTPLogin = utilities.Ptr(strings.TrimSpace(body.SMTPLogin))
	if body.SMTPPassword != "" && body.SMTPPassword != "<redacted>" {
		cxsettingspkg.CxSettings.SMTPPassword = utilities.Ptr(body.SMTPPassword)
	}
	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func postTestEmailSettings(c *gin.Context) {
	var body struct {
		SMTPHost     string `json:"smtpHost"`
		SMTPPort     string `json:"smtpPort"`
		Secure       string `json:"secure"`
		FromName     string `json:"fromName"`
		FromEmail    string `json:"fromEmail"`
		SMTPLogin    string `json:"smtpLogin"`
		SMTPPassword string `json:"smtpPassword"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	port, _ := strconv.Atoi(body.SMTPPort)
	if port == 0 {
		port = 587
	}

	password := body.SMTPPassword
	if password == "<redacted>" {
		password = utilities.PtrVal(cxsettingspkg.CxSettings.SMTPPassword)
	}

	cfg := notifier.EmailConfig{
		Host:      body.SMTPHost,
		Port:      port,
		Login:     body.SMTPLogin,
		Password:  password,
		FromName:  body.FromName,
		FromEmail: body.FromEmail,
		Secure:    body.Secure,
	}

	err := notifier.SendEmail(cfg, []string{body.FromEmail}, "Chronix SMTP Test", "This is a test email from Chronix to verify your SMTP settings.", "")
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "SMTP test failed", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}

func getSMSSettings(c *gin.Context) {
	s := cxsettingspkg.CxSettings

	accountSid := utilities.PtrVal(s.TwilioAccountSid)
	if accountSid == "" {
		accountSid = utilities.PtrVal(s.TwilioUsername)
	}

	authToken := s.TwilioAPISecret
	if utilities.PtrVal(authToken) == "" {
		authToken = s.TwilioPassword
	}

	resp := gin.H{
		"provider":   utilities.PtrVal(s.SmsProvider),
		"fromNumber": utilities.PtrVal(s.SmsFromNumber),
		"accountSid": accountSid,
		"authToken":  redactIfSet(authToken),
	}
	restresponse.RestSuccess(c, resp)
}

func putSMSSettings(c *gin.Context) {
	var body struct {
		Provider   string `json:"provider"`
		FromNumber string `json:"fromNumber"`
		AccountSid string `json:"accountSid"`
		AuthToken  string `json:"authToken"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	cxsettingspkg.CxSettings.SmsProvider = utilities.Ptr(strings.TrimSpace(body.Provider))
	cxsettingspkg.CxSettings.SmsFromNumber = utilities.Ptr(strings.TrimSpace(body.FromNumber))
	cxsettingspkg.CxSettings.TwilioAccountSid = utilities.Ptr(strings.TrimSpace(body.AccountSid))
	if body.AuthToken != "" && body.AuthToken != "<redacted>" {
		cxsettingspkg.CxSettings.TwilioAPISecret = utilities.Ptr(body.AuthToken)
	}

	if err := cxsettingspkg.SetCxSettings(); err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error saving settings", err.Error())
		return
	}
	restresponse.RestSuccessNoContent(c)
}

func postTestSMSSettings(c *gin.Context) {
	var body struct {
		FromNumber string `json:"fromNumber"`
		AccountSid string `json:"accountSid"`
		AuthToken  string `json:"authToken"`
		TestNumber string `json:"testNumber"`
	}
	if err := c.BindJSON(&body); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}

	authToken := body.AuthToken
	if authToken == "<redacted>" {
		authToken = utilities.PtrVal(cxsettingspkg.CxSettings.TwilioAPISecret)
		if authToken == "" {
			authToken = utilities.PtrVal(cxsettingspkg.CxSettings.TwilioPassword)
		}
	}

	cfg := notifier.SMSConfig{
		AccountSid: body.AccountSid,
		AuthToken:  authToken,
		FromNumber: body.FromNumber,
	}

	testNumber := strings.TrimSpace(body.TestNumber)
	if testNumber == "" {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Test number is required")
		return
	}

	digitsOnly := ""
	for _, char := range testNumber {
		if char >= '0' && char <= '9' {
			digitsOnly += string(char)
		}
	}
	if len(digitsOnly) == 10 {
		testNumber = "+1" + digitsOnly
	} else if len(digitsOnly) == 11 && strings.HasPrefix(digitsOnly, "1") {
		testNumber = "+" + digitsOnly
	}

	err := notifier.SendSMS(cfg, []string{testNumber}, "Chronix SMS Test message.")
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Internal, "SMS test failed", err.Error())
		return
	}

	restresponse.RestSuccessNoContent(c)
}
