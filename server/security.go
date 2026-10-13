package server

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/unrolled/secure"
)

const (
	contentSecurityPolicy = "default-src * 'unsafe-inline' 'unsafe-eval' data: blob:; base-uri *; form-action *; frame-ancestors *; img-src * data: blob:; script-src * 'unsafe-inline' 'unsafe-eval' data: blob:; style-src * 'unsafe-inline' data: blob:; font-src * data: blob:; connect-src * data: blob:; media-src * data: blob:; worker-src * data: blob:"
)

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

func isHTMLNavigation(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	return strings.Contains(accept, "text/html") && !strings.Contains(accept, "text/event-stream")
}

func isDatastar(r *http.Request) bool {
	return r.Header.Get("Datastar-Request") == "true"
}
