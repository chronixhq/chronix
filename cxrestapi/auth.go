package cxrestapi

import (
	"chronix/cmd/app/rpc"
	activitypkg "chronix/internal/activity"
	cxuserpkg "chronix/internal/cxuser"
	"chronix/internal/db/models"
	notifypkg "chronix/internal/notify"
	"fmt"
	"log/slog"
	"math/rand"
	"strings"
	"time"

	"github.com/dan-sherwin/go-rest-api-server/restresponse"
	"github.com/dan-sherwin/go-utilities"
	"github.com/gin-gonic/gin"
)

type (
	AuthCommandDef struct {
		AdminCode AdminCodeCommand `cmd:"" help:"Get an admin login code" name:"adminCode"`
	}
	AdminCodeCommand struct{}

	AuthClaim struct {
		UserID  int64  `json:"userId"`
		Admin   bool   `json:"admin"`
		AuthKey string `json:"authKey"`
		Email   string `json:"email"`
		Phone   string `json:"phone"`
		Name    string `json:"name"`
	}
)

var (
	ActiveAdminCode = ""
	//jwtSecret       = []byte("6o0H4*881$&Jz9!rIt5d0mp2T%q1NB1A$*%N#2y4R9DMjRkZt!l762yN0$hkdfoo")
	jwtSecret = []byte("6o0H4*881$&Jz9!rIt5d0mp2T%q1NB1A$*%N#2y4R9DMjRkZt!l762yN0$hkd$mw")
)

func init() {
	_ = rpc.RegisterName("AdminCode", &AdminCodeCommand{})
}

func authRouter(r *gin.Engine) {
	r.GET("/auth/settings/:code", authAdminCode)
	r.POST("/auth/login", userLogin)
}

func userLogin(c *gin.Context) {
	var loginData struct{ Email, Password string }
	if err := c.BindJSON(&loginData); err != nil {
		restresponse.RestErrorRespond(c, restresponse.BadRequest, "Invalid JSON")
		return
	}
	user, err := cxuserpkg.Login(loginData.Email, loginData.Password)
	if err != nil {
		restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid credentials")
		return
	}
	// Record activity
	_ = activitypkg.RecordUserActivity(user.ID, "Logged in", "", c.ClientIP(), c.Request.UserAgent())
	tokenRefresh(c, user)
}

// refreshAuthCookie generates a new JWT based on the provided user and sets it as a cookie.
// It also revokes the old auth key (if present) and registers the new one.
// This function does NOT write any API response, so it can be used in endpoints
// that need to rotate the token but return their own payload.
func refreshAuthCookie(c *gin.Context, user *cxuserpkg.CxUser) {
	claim := AuthClaim{
		UserID:  user.ID,
		Admin:   user.Admin,
		AuthKey: fmt.Sprintf("%d:%d:%d", user.ID, user.Sv, time.Now().UnixMilli()),
		Email:   user.Email,
		Phone:   utilities.PtrVal(user.Phone),
		Name:    user.Name,
	}
	jwt, err := utilities.GenerateJWT(claim, time.Hour*24*7, jwtSecret)
	if err != nil {
		slog.Error("Error generating JWT", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error generating JWT")
		return
	}
	if len(user.AuthKey) > 0 {
		delete(authKeys, user.AuthKey)
	}
	authKeys[claim.AuthKey] = struct{}{}
	c.SetCookie(
		"cxtoken",
		jwt,
		60*60*24*7,
		"/",
		"",
		false,
		true,
	)
}

func tokenRefresh(c *gin.Context, user *cxuserpkg.CxUser) {
	refreshAuthCookie(c, user)
	restresponse.RestSuccess(c, gin.H{"status": "logged in"})
}

func logout(c *gin.Context) {
	user := userFromGinContext(c)
	RevokeAuthKey(user.AuthKey)
	c.SetCookie("cxtoken", "", -1, "/", "", false, true)
	_ = activitypkg.RecordUserActivity(user.ID, "Logged out", "", c.ClientIP(), c.Request.UserAgent())
	restresponse.RestSuccess(c, gin.H{"message": "Logged out"})
}

func fullLogout(c *gin.Context) {
	_ = userFromGinContext(c).IncrementSv()
	logout(c)
}

func authAdminCode(c *gin.Context) {
	if !strings.EqualFold(strings.ReplaceAll(c.Param("code"), " ", ""), strings.ReplaceAll(ActiveAdminCode, " ", "")) {
		slog.Error("invalid admin code", "op", "auth")
		restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid admin code")
		return
	}
	claim := AuthClaim{
		UserID:  0,
		Admin:   true,
		AuthKey: "",
		Email:   "",
		Phone:   "",
		Name:    "",
	}
	jwt, err := utilities.GenerateJWT(claim, time.Minute*10, jwtSecret)
	if err != nil {
		slog.Error("Error generating JWT", "error", err)
		restresponse.RestErrorRespond(c, restresponse.Internal, "Error generating JWT")
		return
	}
	c.SetCookie(
		"cxtoken",
		jwt,
		60*10,
		"/",
		"",
		false,
		true,
	)
	ActiveAdminCode = ""
	// Record admin login with userId 0
	_ = activitypkg.RecordUserActivity(0, "Logged in as Admin", "", c.ClientIP(), c.Request.UserAgent())
	// Emit security alert: admin login code used (Always On)
	data := map[string]any{"ip": c.ClientIP()}
	notifypkg.TryCreateNotification(notifypkg.CategorySecurity, notifypkg.SeverityInfo, "Admin login code used", nil, &data)
	restresponse.RestSuccess(c, gin.H{"status": "logged in"})
}

func (a *AdminCodeCommand) Run() error {
	var code string
	err := rpc.Call("AdminCode.GetAdminCode", &struct{}{}, &code)
	if err != nil {
		slog.Error("Error getting admin code", "error", err)
		fmt.Println("Error getting admin code")
		return err
	}
	fmt.Printf("Admin login code: %s\nUse this temporary admin code to login to Chronix.\nThis code is good for 10 minutes.\n", code)
	return nil
}

func (a *AdminCodeCommand) GetAdminCode(_ *struct{}, code *string) error {
	GenerateAdminCode()
	*code = ActiveAdminCode
	return nil
}

func GenerateAdminCode() {
	var genstr = func() string {
		const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
		result := make([]byte, 4)
		for i := range result {
			result[i] = charset[rand.Intn(len(charset))]
		}
		return string(result)
	}
	part1 := genstr()
	part2 := genstr()
	part3 := genstr()
	ActiveAdminCode = fmt.Sprintf("%s %s %s", part1, part2, part3)
	// Record activity
	_ = activitypkg.RecordUserActivity(0, "Admin login code generated", "", "", "")
	go func() {
		time.Sleep(time.Minute * 10)
		ActiveAdminCode = ""
	}()
}

func AuthGinMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		cxToken, err := c.Cookie("cxtoken")
		if err != nil {
			restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid credentials")
			c.Abort()
			return
		}
		var claims AuthClaim
		err = utilities.ExtractJwtClaimsInto(cxToken, jwtSecret, &claims)
		if err != nil {
			restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid credentials")
			c.Abort()
			return
		}
		if claims.UserID != 0 {
			if _, ok := authKeys[claims.AuthKey]; !ok {
				restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid credentials")
				c.Abort()
				return
			}
		}
		c.Set("user", &cxuserpkg.CxUser{
			CxUser: models.CxUser{
				ID:    claims.UserID,
				Email: claims.Email,
				Name:  claims.Name,
				Phone: utilities.Ptr(claims.Phone),
				Admin: claims.Admin,
			},
			AuthKey: claims.AuthKey,
		})
		c.Next()
	}
}

func adminFunc(fn gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		user := userFromGinContext(c)
		if user.Admin {
			fn(c)
		} else {
			restresponse.RestErrorRespond(c, restresponse.Unauthenticated, "Invalid credentials")
			return
		}
	}
}

func userFromGinContext(c *gin.Context) *cxuserpkg.CxUser {
	return c.MustGet("user").(*cxuserpkg.CxUser)
}
