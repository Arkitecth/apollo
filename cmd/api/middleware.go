package main

import (
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/Arkitecth/apollo/internal/data"
	"github.com/Arkitecth/apollo/validator"
	"github.com/tomasen/realip"
	"golang.org/x/time/rate"
)

func (app *application) recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				w.Header().Set("Connection", "Close")
				app.serverErrorResponse(w, r, fmt.Errorf("%s", err))
			}
		}()

		next.ServeHTTP(w, r)
	})
}

func (app *application) rateLimit(next http.Handler) http.Handler {
	type client struct {
		limiter  *rate.Limiter
		lastSeen time.Time
	}
	var (
		mu      sync.Mutex
		clients = make(map[string]*client)
	)

	go func() {
		for {
			time.Sleep(time.Minute)

			mu.Lock()

			for ip, client := range clients {
				if time.Since(client.lastSeen) > 3*time.Minute {
					delete(clients, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if app.config.limiter.enabled {
			ip := realip.FromRequest(r)
			mu.Lock()
			if _, ok := clients[ip]; !ok {
				rateLimiter := rate.NewLimiter(rate.Limit(app.config.limiter.rps), app.config.limiter.burst)
				clients[ip] = &client{limiter: rateLimiter}
			}

			clients[ip].lastSeen = time.Now()

			if !clients[ip].limiter.Allow() {
				mu.Unlock()
				app.rateLimitExceededResponse(w, r)
				return
			}
			mu.Unlock()
		}
		next.ServeHTTP(w, r)
	})
}

func (app *application) authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Authorization")

		header := r.Header.Get("Authorization")

		if header == "" {
			r = app.setUserContext(r, data.AnomynousUser)
			next.ServeHTTP(w, r)
			return
		}

		headerParts := strings.Split(header, " ")

		if len(headerParts) != 2 || headerParts[0] != "Bearer" {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		token := headerParts[1]

		v := validator.New()

		if data.ValidatePlaintext(v, token); !v.Valid() {
			app.invalidAuthenticationTokenResponse(w, r)
			return
		}

		user, err := app.models.UserModel.GetUserFromToken(data.ScopeAuthentication, token)
		if err != nil {
			switch {
			case errors.Is(err, data.ErrRecordNotFound):
				app.invalidAuthenticationTokenResponse(w, r)
			default:
				app.serverErrorResponse(w, r, err)
			}
			return
		}

		r = app.setUserContext(r, user)

		next.ServeHTTP(w, r)
	})
}

func (app *application) requireAuthenticatedUser(next http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserContext(r)
		if user.IsAnonymous() {
			app.authenticationRequiredResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

}

func (app *application) requireActivatedUser(next http.HandlerFunc) http.HandlerFunc {
	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserContext(r)
		if !user.Activated {
			app.inactiveAccountResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	})

	return app.requireAuthenticatedUser(fn)
}

func (app *application) requireAuthorizedUser(code string, next http.HandlerFunc) http.HandlerFunc {
	fn := func(w http.ResponseWriter, r *http.Request) {
		user := app.getUserContext(r)

		permissions, err := app.models.PermissionModel.GetAllForUser(user.ID)
		if err != nil {
			app.serverErrorResponse(w, r, err)
			return
		}

		if !permissions.Include(code) {
			app.notAuthorizedResponse(w, r)
			return
		}

		next.ServeHTTP(w, r)
	}

	return app.requireActivatedUser(fn)

}

func (app *application) enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Vary", "Origin")

		w.Header().Set("Vary", "Access-Control-Request-Method")

		origin := r.Header.Get("Origin")

		if origin != "" {
			for i := range len(app.config.cors.trustedOrigins) {
				if origin == app.config.cors.trustedOrigins[i] {
					w.Header().Set("Access-Control-Allow-Origin", origin)

					if r.Method == http.MethodOptions && r.Header.Get("Access-Control-Request-Method") != "" {
						w.Header().Set("Access-Control-Allow-Methods", "OPTIONS, PATCH, PUT, DELETE")
						w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

						w.WriteHeader(http.StatusOK)
						return
					}

					break
				}

			}

		}

		next.ServeHTTP(w, r)
	})

}

type metricsWriter struct {
	wrappedWriter http.ResponseWriter
	headerWritten bool
	status        int
}

func NewMetricsWriter(w http.ResponseWriter) *metricsWriter {
	return &metricsWriter{
		wrappedWriter: w,
		status:        http.StatusOK,
	}
}

func (mw *metricsWriter) Header() http.Header {
	return mw.wrappedWriter.Header()
}

func (mw *metricsWriter) Write(p []byte) (int, error) {
	mw.headerWritten = true
	return mw.wrappedWriter.Write(p)
}

func (mw *metricsWriter) WriteHeader(statusCode int) {
	mw.wrappedWriter.WriteHeader(statusCode)

	if !mw.headerWritten {
		mw.status = statusCode
		mw.headerWritten = true
	}
}
func (mw *metricsWriter) Unwrap() http.ResponseWriter {
	return mw.wrappedWriter
}

func (app *application) metrics(next http.Handler) http.Handler {
	var (
		totalRequestsReceived           = expvar.NewInt("total_requests_received")
		totalResponsesSent              = expvar.NewInt("total_responses_sent")
		totalProcessingTimeMicroseconds = expvar.NewInt("total_processing_time_μs")
		totalResponsesSentByStatus      = expvar.NewMap("total_responses_sent_by_status")
	)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		mw := NewMetricsWriter(w)

		totalRequestsReceived.Add(1)

		next.ServeHTTP(mw, r)

		totalResponsesSent.Add(1)

		totalResponsesSentByStatus.Add(strconv.Itoa(mw.status), 1)

		duration := time.Since(start).Nanoseconds()
		totalProcessingTimeMicroseconds.Set(duration)
	})
}
