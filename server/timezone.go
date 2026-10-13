package server

import (
	"net/http"
	"sync"
	"time"
	_ "time/tzdata"

	"afastchat.com/views"
)

var timeZones sync.Map

func loadTimeZone(name string) *time.Location {
	if name == "" || len(name) > 64 {
		return nil
	}
	if loc, ok := timeZones.Load(name); ok {
		return loc.(*time.Location)
	}
	loc, err := time.LoadLocation(name)
	if err != nil {
		return nil
	}
	timeZones.Store(name, loc)
	return loc
}

func withTimeZone(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if cookie, err := r.Cookie("tz"); err == nil {
			if loc := loadTimeZone(cookie.Value); loc != nil {
				r = r.WithContext(views.WithLocation(r.Context(), loc))
			}
		}
		next.ServeHTTP(w, r)
	})
}
