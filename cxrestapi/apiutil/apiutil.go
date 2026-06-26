package apiutil

import (
	cxuserpkg "chronix/internal/cxuser"
	"fmt"

	"github.com/gin-gonic/gin"
)

func Atoi64(s string) int64 {
	var v int64
	_, _ = fmt.Sscan(s, &v)
	return v
}

func UserFromGinContext(c *gin.Context) *cxuserpkg.CxUser {
	return c.MustGet("user").(*cxuserpkg.CxUser)
}
