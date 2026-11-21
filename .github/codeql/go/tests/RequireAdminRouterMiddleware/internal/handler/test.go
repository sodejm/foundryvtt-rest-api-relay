package handler

import (
	appmw "github.com/ThreeHats/foundryvtt-rest-api-relay/go-relay/internal/middleware"
	otherhandler "github.com/ThreeHats/foundryvtt-rest-api-relay/go-relay/internal/otherhandler"
	"github.com/go-chi/chi/v5"
	"net/http"
)

func handler(http.ResponseWriter, *http.Request)           {}
func TestReportHandler(http.ResponseWriter, *http.Request) {}

func AdminRouter() chi.Router {
	r := chi.NewRouter()
	r.MethodFunc("GET", "/api/direct-method", handler)                                     // BAD: MethodFunc path is argument 1
	r.Get("/api/direct", handler)                                                          // BAD: direct API registration
	r.Get("/api/tests/report", TestReportHandler)                                          // exact direct exception
	r.Get("/api/tests/report", otherhandler.TestReportHandler)                             // BAD: foreign handler
	r.Get("/apiary", handler)                                                              // non-API prefix
	r.Route("/api/direct-route", func(direct chi.Router) { direct.Get("/leaf", handler) }) // BAD
	r.Mount("/api/direct-mount", DirectRouter())                                           // BAD
	r.Route("/other", func(other chi.Router) {
		other.Route("/api", func(outside chi.Router) { outside.Get("/public", handler) })
	})
	r.Get("/", handler) // intentional public UI, outside API
	r.Route("/auth", func(auth chi.Router) {
		auth.Post("/login", handler)  // credential exchange, outside API
		auth.Post("/logout", handler) // own cookie/CSRF checks, outside API
		auth.With(appmw.RequireAdmin()).Get("/me", handler)
	})
	r.Group(func(api chi.Router) {
		api.Use(appmw.Other())
		api.Route("/api", func(routes chi.Router) {
			routes.Get("/tests/report", otherhandler.TestReportHandler) // BAD: foreign same-name handler
			routes.Get("/tests/report", TestReportHandler)              // exact intentional IP-only exception
			routes.Post("/tests/report", TestReportHandler)             // BAD: method not exempt
			routes.Get("/tests/report", handler)                        // BAD: another handler is not exempt
			routes.Get("/unguarded", handler)                           // BAD
			routes.Group(func(protected chi.Router) {
				protected.Use(appmw.RequireAdmin())
				alias := protected
				alias.Get("/safe-alias", handler)
				protected.Route("/nested", func(nested chi.Router) {
					nested.Group(func(inner chi.Router) { inner.Mount("/safe", SharedRouter()) })
				})
			})
			routes.Group(func(sibling chi.Router) {
				sibling.Mount("/unsafe", SharedRouter()) // BAD: sibling must not inherit guard
			})
			routes.Group(func(ambiguous chi.Router) {
				candidate := ambiguous
				if len("x") > 0 {
					candidate = chi.NewRouter()
				}
				candidate.Use(appmw.RequireAdmin())
				ambiguous.Get("/ambiguous", handler) // BAD: guard may refer to another router
			})
			routes.With(appmw.RequireAdmin()).Get("/safe-with", handler)
			routes.With(appmw.Other()).Get("/unsafe-with", handler) // BAD
			mounted := ProtectedRouter()
			routes.Mount("/self-protected", mounted)
			routes.Group(func(conditional chi.Router) {
				if len("x") > 0 {
					conditional.Use(appmw.RequireAdmin())
				}
				conditional.Get("/conditional", handler) // BAD: conditional guard
			})
			routes.Group(func(late chi.Router) {
				late.Get("/late", handler) // BAD: guard after registration
				late.Use(appmw.RequireAdmin())
			})
			routes.Route("/different", func(other chi.Router) {
				other.Get("/tests/report", TestReportHandler) // BAD: exception path not exact
			})
		})
	})
	return r
}

func SharedRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/leaf", handler) // BAD: one mount path lacks middleware
	return r
}
func ProtectedRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(appmw.RequireAdmin())
	r.Get("/leaf", handler)
	return r
}
func UnrelatedRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/public", handler) // outside AdminRouter
	return r
}

func DirectRouter() chi.Router {
	r := chi.NewRouter()
	r.Get("/leaf", handler) // BAD: direct /api/... mount
	return r
}
