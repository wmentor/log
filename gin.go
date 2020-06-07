package log

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func GinLogger() gin.HandlerFunc {

	return func(c *gin.Context) {

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		c.Next()

		str := fmt.Sprintf("%s %s %s %d \"%s\"", c.ClientIP(), c.Request.Method, path, c.Writer.Status(), c.Request.UserAgent())

		switch {
		case c.Writer.Status() >= http.StatusBadRequest && c.Writer.Status() < http.StatusInternalServerError:
			Warn(str)

		case c.Writer.Status() >= http.StatusInternalServerError:
			Error(str)

		default:
			Info(str)
		}

		if len(c.Errors) > 0 {
			Error(c.Errors.String())
		}

	}
}
