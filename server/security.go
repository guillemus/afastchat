package server

import (
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/httprate"
	"github.com/unrolled/secure"
)

const (
	tooManyAttemptsMessage = "Too many attempts. Try again later."
	contentSecurityPolicy  = "default-src 'self'; base-uri 'self'; form-action 'self'; frame-ancestors 'none'; img-src 'self' data:; script-src 'self' 'unsafe-eval' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com https://unpkg.com; font-src 'self' https://fonts.gstatic.com data:; connect-src 'self'"
)

func newAuthRateLimiter() *httprate.RateLimiter {
	return httprate.NewRateLimiter(
		5,
		10*time.Minute,
		httprate.WithResponseHeaders(httprate.ResponseHeaders{}),
	)
}

func newCrossOriginProtection() *http.CrossOriginProtection {
	protection := http.NewCrossOriginProtection()
	protection.SetDenyHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slog.Warn(
			"cross-origin request rejected",
			"path", r.URL.Path,
			"origin", r.Header.Get("Origin"),
			"sec_fetch_site", r.Header.Get("Sec-Fetch-Site"),
		)
		http.Error(w, "forbidden", http.StatusForbidden)
	}))
	return protection
}

func newSecureMiddleware(dev bool) *secure.Secure {
	return secure.New(secure.Options{
		IsDevelopment:         dev,
		STSSeconds:            31536000,
		STSIncludeSubdomains:  true,
		FrameDeny:             true,
		ContentTypeNosniff:    true,
		ReferrerPolicy:        "strict-origin-when-cross-origin",
		ContentSecurityPolicy: contentSecurityPolicy,
	})
}

func (x *Server) authRateLimited(w http.ResponseWriter, r *http.Request, email string) bool {
	if x.authIPLimiter.OnLimit(w, r, clientIP(r)) {
		return true
	}
	if email != "" && x.authEmailLimiter.OnLimit(w, r, email) {
		return true
	}
	return false
}

func isHTMLNavigation(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html") && !strings.Contains(accept, "text/event-stream")
}

func clientIP(r *http.Request) string {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}
