// 🚀 Fiber is an Express inspired web framework written in Go with 💖
// 📌 API Documentation: https://fiber.wiki
// 📝 Github Repository: https://github.com/gofiber/fiber

package redirect

import (
	"regexp"
	"strconv"
	"strings"

	"github.com/gofiber/fiber"
)

// Config ...
type Config struct {
	// Filter defines a function to skip middleware.
	// Optional. Default: nil
	Filter func(*fiber.Ctx) bool
	// Rules defines the URL path rewrite rules. The values captured in asterisk can be
	// retrieved by index e.g. $1, $2 and so on.
	// Required. Example:
	// "/old":              "/new",
	// "/api/*":            "/$1",
	// "/js/*":             "/public/javascripts/$1",
	// "/users/*/orders/*": "/user/$1/order/$2",
	Rules map[string]string
	// The status code when redirecting
	// This is ignored if Redirect is disabled
	// Optional. Default: 302 Temporary Redirect
	StatusCode int
	
	rulesRegex map[*regexp.Regexp]string
}

// New ...
func New(config ...Config) func(*fiber.Ctx) {
	// Init config
	var cfg Config
	if len(config) > 0 {
		cfg = config[0]
	}
	if cfg.StatusCode == 0 {
		cfg.StatusCode = 302 // Temporary Redirect
	}
	cfg = config[0]
	cfg.rulesRegex = map[*regexp.Regexp]string{}
	// Initialize
	for k, v := range cfg.Rules {
		k = strings.Replace(k, "*", "(.*)", -1)
		k = k + "$"
		cfg.rulesRegex[regexp.MustCompile(k)] = v
	}
	// Middleware function
	return func(c *fiber.Ctx) {
		// Filter request to skip middleware
		if cfg.Filter != nil && cfg.Filter(c) {
			c.Next()
			return
		}
		// Rewrite
		for k, v := range cfg.rulesRegex {
			replacer := captureTokens(k, c.Path())
			if replacer != nil {
				c.Redirect(replacer.Replace(v), cfg.StatusCode)
				return
			}
		}
		c.Next()
	}
}

// https://github.com/labstack/echo/blob/master/middleware/rewrite.go
func captureTokens(pattern *regexp.Regexp, input string) *strings.Replacer {
	groups := pattern.FindAllStringSubmatch(input, -1)
	if groups == nil {
		return nil
	}
	values := groups[0][1:]
	replace := make([]string, 2*len(values))
	for i, v := range values {
		j := 2 * i
		replace[j] = "$" + strconv.Itoa(i+1)
		replace[j+1] = v
	}
	return strings.NewReplacer(replace...)
}
