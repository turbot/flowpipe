package middleware

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"github.com/unrolled/secure"
)

func SecurityMiddleware(ctx context.Context) gin.HandlerFunc {
	options := secure.Options{
		// In development, many options are turned off automatically
		IsDevelopment: viper.GetString("url.base") == "http://localhost:7103",

		// Redirect HTTP to HTTPS
		SSLRedirect:          false,
		SSLTemporaryRedirect: false,
		SSLHost:              viper.GetString("web.secure.ssl_host"),

		// Set HSTS header, telling the browser to always use SSL. 1 year expiration.
		// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Strict-Transport-Security
		// See https://hstspreload.org/
		STSSeconds:           31536000,
		STSIncludeSubdomains: true,
		STSPreload:           true,

		// Only allow the site in a frame within it's own domain
		// See https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/X-Frame-Options
		CustomFrameOptionsValue: "SAMEORIGIN",

		// Tell the browser the MIME type is accurate, no need to sniff or change it
		ContentTypeNosniff: true,

		// Careful sending of referrer data, following recommendation at
		// https://scotthelme.co.uk/a-new-security-header-referrer-policy/
		ReferrerPolicy: "strict-origin-when-cross-origin",

		// Permissions policy is required by sites like securityheaders.io, even though
		// they don't really mind what is set. Make a minimal setting for now.
		PermissionsPolicy: "geolocation 'self'",
	}

	slog.Debug("Security middleware options", "IsDevelopment", options.IsDevelopment, "SSLHost", options.SSLHost, "STSSeconds", options.STSSeconds, "STSIncludeSubdomains", options.STSIncludeSubdomains, "STSPreload", options.STSPreload, "CustomFrameOptionsValue", options.CustomFrameOptionsValue, "ContentTypeNosniff", options.ContentTypeNosniff, "ReferrerPolicy", options.ReferrerPolicy, "PermissionsPolicy", options.PermissionsPolicy)

	secureMiddleware := secure.New(options)

	return func() gin.HandlerFunc {
		return func(c *gin.Context) {
			err := secureMiddleware.Process(c.Writer, c.Request)
			// If there was an error, do not continue.
			if err != nil {
				c.Abort()
				return
			}
			// Avoid header rewrite if response is a redirection.
			if status := c.Writer.Status(); status > 300 && status < 399 {
				c.Abort()
			}
		}
	}()
}
