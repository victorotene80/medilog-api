package rest

import (
	"io"
	"net/http"
	"time"

	"github.com/victorotene80/medilog-api/docs"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/handler"
	appmw "github.com/victorotene80/medilog-api/internal/interfaces/middleware"
)

type Router struct {
	mux *chi.Mux

	Logger                  *zap.Logger
	AuthMiddleware          *appmw.AuthMiddleware
	AdminMiddleware         *appmw.AdminMiddleware
	RateLimiter             *appmw.RateLimiter
	TelemetryMiddleware     *appmw.TelemetryMiddleware
	CORSOrigins             []string
	AuthHandler             *handler.AuthHandler
	OTPHandler              *handler.OTPHandler
	UserHandler             *handler.UserHandler
	UserAllergyHandler      *handler.UserAllergyHandler
	EmergencyContactHandler *handler.EmergencyContactHandler
	MedicationHandler       *handler.MedicationHandler
	VisitHandler            *handler.VisitHandler
	AIHandler               *handler.AIHandler
	ScanHandler             *handler.ScanHandler
	ReferenceHandler        *handler.ReferenceHandler
	AllergyHandler          *handler.AllergyHandler
	FunFactHandler          *handler.FunFactHandler
	DashboardHandler        *handler.DashboardHandler
	SupportTicketHandler    *handler.SupportTicketHandler
	FeedbackHandler         *handler.FeedbackHandler
	NotificationHandler     *handler.NotificationHandler
	AuditLogHandler         *handler.AuditLogHandler
	HealthHandler           *handler.HealthHandler
	SchedulerHandler        *handler.SchedulerHandler
	IsLive                  bool
}

func NewRouter(
	logger *zap.Logger,
	authMiddleware *appmw.AuthMiddleware,
	adminMiddleware *appmw.AdminMiddleware,
	rateLimiter *appmw.RateLimiter,
	telemetryMiddleware *appmw.TelemetryMiddleware,
	corsOrigins []string,
	authHandler *handler.AuthHandler,
	otpHandler *handler.OTPHandler,
	userHandler *handler.UserHandler,
	userAllergyHandler *handler.UserAllergyHandler,
	emergencyContactHandler *handler.EmergencyContactHandler,
	medicationHandler *handler.MedicationHandler,
	visitHandler *handler.VisitHandler,
	aiHandler *handler.AIHandler,
	referenceHandler *handler.ReferenceHandler,
	allergyHandler *handler.AllergyHandler,
	scanHandler *handler.ScanHandler,
	funFactHandler *handler.FunFactHandler,
	dashboardHandler *handler.DashboardHandler,
	supportTicketHandler *handler.SupportTicketHandler,
	feedbackHandler *handler.FeedbackHandler,
	notificationHandler *handler.NotificationHandler,
	auditLogHandler *handler.AuditLogHandler,
	healthHandler *handler.HealthHandler,
	schedulerHandler *handler.SchedulerHandler,
	isLive bool,
) *Router {
	return &Router{
		mux:                     chi.NewRouter(),
		Logger:                  logger,
		AuthMiddleware:          authMiddleware,
		AdminMiddleware:         adminMiddleware,
		RateLimiter:             rateLimiter,
		TelemetryMiddleware:     telemetryMiddleware,
		CORSOrigins:             corsOrigins,
		AuthHandler:             authHandler,
		OTPHandler:              otpHandler,
		UserHandler:             userHandler,
		UserAllergyHandler:      userAllergyHandler,
		EmergencyContactHandler: emergencyContactHandler,
		MedicationHandler:       medicationHandler,
		VisitHandler:            visitHandler,
		AIHandler:               aiHandler,
		ReferenceHandler:        referenceHandler,
		AllergyHandler:          allergyHandler,
		ScanHandler:             scanHandler,
		FunFactHandler:          funFactHandler,
		DashboardHandler:        dashboardHandler,
		SupportTicketHandler:    supportTicketHandler,
		FeedbackHandler:         feedbackHandler,
		NotificationHandler:     notificationHandler,
		AuditLogHandler:         auditLogHandler,
		HealthHandler:           healthHandler,
		SchedulerHandler:        schedulerHandler,
		IsLive:                  isLive,
	}
}

func (rt *Router) Setup() http.Handler {
	rt.mux.Use(chiMw.RequestID)
	rt.mux.Use(chiMw.RealIP)
	rt.mux.Use(appmw.PanicRecovery(rt.Logger))
	rt.mux.Use(appmw.RequestMetadata)
	rt.mux.Use(appmw.NewPrometheusMetrics().Handle)
	rt.mux.Use(chiMw.Logger)
	rt.mux.Use((&appmw.SecurityHeaders{}).Handle)
	rt.mux.Use((&appmw.BodyLimit{}).Handle)

	if rt.TelemetryMiddleware != nil {
		rt.mux.Use(rt.TelemetryMiddleware.Handle)
	}

	rt.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: rt.CORSOrigins,
		AllowedMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodOptions,
		},
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-Device-ID",
			"X-Device-Name",
			"X-Device-Fingerprint",
		},
		ExposedHeaders: []string{
			"Link",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
			"X-RateLimit-Reset",
			"Retry-After",
		},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	rt.mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rt.mux.Get("/health/ready", rt.HealthHandler.Ready)

	// Driven by Cloud Scheduler once a minute, replacing the in-process ticker
	// that Cloud Run's CPU throttling made unreliable. On the root mux rather
	// than under /api/v1 because it is not part of the public API and must not
	// pick up AuthMiddleware — the caller is a machine with a shared secret,
	// not a user with a session.
	//
	// Unregistered entirely when SCHEDULER_ENABLED is false, so a misconfigured
	// environment 404s here instead of quietly accepting ticks and doing
	// nothing with them.
	//
	// The rate limit is a blast-radius cap, not a correctness control: the scan
	// is already idempotent behind an advisory lock. 10/min leaves room for a
	// Cloud Scheduler retry or a manual run while refusing a flood.
	if rt.SchedulerHandler != nil {
		rt.mux.With(rt.limit(appmw.RateLimitRule{
			Name:      "scheduler_tick_ip",
			Limit:     10,
			Window:    time.Minute,
			KeyFields: []string{"ip"},
		})).Post("/internal/scheduler/reminders/tick", rt.SchedulerHandler.TickReminders)
	}

	// Prometheus scrapes this at /metrics on the container port (see
	// observability/prometheus.yml), so it must stay on the root mux and off the
	// authenticated /api/v1 tree. It is not exposed publicly: docker-compose
	// binds the API to 127.0.0.1:8080 and Prometheus reaches it over the compose
	// network as api:8080.
	rt.mux.Get("/metrics", promhttp.Handler().ServeHTTP)

	// Serve the spec that swag compiled into the binary, not a file read off
	// disk.
	//
	// These used to be http.ServeFile(w, r, "docs/swagger.json"), which resolves
	// against the process working directory. The runtime Docker image copies
	// only the binary — there is no docs/ directory in it — so every deployed
	// environment served a 404 here and Swagger UI came up empty no matter how
	// recently the spec had been regenerated. main.go already blank-imports the
	// generated docs package, so the spec is present in the binary; it just was
	// not being used.
	//
	// Note the spec is embedded at BUILD time: after `make swagger` the service
	// must be rebuilt for the change to appear.
	serveSpec := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		// Never cache: a stale spec is exactly the failure this replaces.
		w.Header().Set("Cache-Control", "no-store, must-revalidate")
		_, _ = io.WriteString(w, docs.SwaggerInfo.ReadDoc())
	}

	rt.mux.Get("/swagger/doc.json", serveSpec)
	rt.mux.Get("/swagger.json", serveSpec)
	rt.mux.Get("/swagger/*", func(w http.ResponseWriter, r *http.Request) {
		// Override strict CSP for Swagger UI to allow CDN scripts and inline execution
		w.Header().Set("Content-Security-Policy", "default-src 'self' 'unsafe-inline' https://unpkg.com")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write([]byte(`<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>MediLog API - Swagger UI</title>
  <link rel="stylesheet" href="https://unpkg.com/swagger-ui-dist@5/swagger-ui.css" />
</head>
<body>
<div id="swagger-ui"></div>
<script src="https://unpkg.com/swagger-ui-dist@5/swagger-ui-bundle.js" crossorigin></script>
<script>
  window.onload = () => {
    window.ui = SwaggerUIBundle({
      url: '/swagger/doc.json',
      dom_id: '#swagger-ui',
    });
  };
</script>
</body>
</html>`))
	})

	rt.mux.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.With(rt.limit(
				appmw.RateLimitRule{
					Name:      "auth_register_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				},
				appmw.RateLimitRule{
					Name:       "auth_register_identity",
					Limit:      10,
					Window:     time.Hour,
					KeyFields:  []string{"ip"},
					BodyFields: []string{"email", "phone"},
				},
			)).Post("/register", rt.AuthHandler.CreateUser)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_login_identity",
				Limit:      5,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"email", "phone"},
			})).Post("/login", rt.AuthHandler.Login)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:      "auth_google_ip",
				Limit:     10,
				Window:    10 * time.Minute,
				KeyFields: []string{"ip"},
			})).Post("/google", rt.AuthHandler.GoogleLogin)

			r.With(rt.limit(
				appmw.RateLimitRule{
					Name:      "auth_login_otp_ip",
					Limit:     10,
					Window:    10 * time.Minute,
					KeyFields: []string{"ip"},
				},
				appmw.RateLimitRule{
					Name:       "auth_login_otp_identity",
					Limit:      5,
					Window:     10 * time.Minute,
					KeyFields:  []string{"ip"},
					BodyFields: []string{"recipient"},
				},
			)).Post("/login-otp", rt.AuthHandler.LoginViaOTP)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_forgot_password",
				Limit:      3,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"recipient"},
			})).Post("/forgot-password", rt.AuthHandler.ForgotPassword)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_reset_password",
				Limit:      5,
				Window:     15 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"recipient"},
			})).Post("/reset-password", rt.AuthHandler.ResetPassword)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_otp_request",
				Limit:      10,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"recipient", "purpose"},
			})).Post("/otp/request", rt.OTPHandler.RequestOTP)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_otp_verify",
				Limit:      5,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"recipient", "purpose"},
			})).Post("/otp/verify", rt.OTPHandler.VerifyOTP)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_onboarding_otp_verify",
				Limit:      5,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip"},
				BodyFields: []string{"recipient", "purpose"},
			})).Post("/otp/verify-onboarding", rt.OTPHandler.VerifyOnboardingOTP)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:      "auth_refresh_ip",
				Limit:     30,
				Window:    time.Hour,
				KeyFields: []string{"ip"},
			})).Post("/refresh", rt.AuthHandler.RefreshSession)

			r.Group(func(r chi.Router) {
				r.Use(rt.AuthMiddleware.Handle)

				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "auth_logout_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/logout", rt.AuthHandler.Logout)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "auth_change_password_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/change-password", rt.AuthHandler.ChangePassword)
			})
		})

		r.Route("/reference", func(r chi.Router) {
			// Apply a generous rate limit for reference data reads
			r.Use(rt.limit(appmw.RateLimitRule{
				Name:      "reference_read_ip",
				Limit:     100,
				Window:    time.Minute,
				KeyFields: []string{"ip"},
			}))

			r.Get("/countries", rt.ReferenceHandler.ListCountries)

			// Master allergy reference list.
			// GET is public. Mutations (POST/PUT/DELETE) require admin role.
			//
			// Endpoints:
			// GET    /api/v1/reference/allergies
			// POST   /api/v1/reference/allergies
			// PUT    /api/v1/reference/allergies/{id}
			// DELETE /api/v1/reference/allergies/{id}
			r.Route("/allergies", func(r chi.Router) {
				r.Get("/", rt.AllergyHandler.ListAllergies)

				r.Group(func(r chi.Router) {
					// AdminMiddleware only reads AuthContext out of the request
					// context; it does not authenticate. Without Handle above it,
					// every request here 401s before CheckAdminAccess is reached.
					r.Use(rt.AuthMiddleware.Handle)
					r.Use(rt.AdminMiddleware.Handle)
					r.Use(rt.limit(appmw.RateLimitRule{
						Name:      "ref_allergy_mutation_ip",
						Limit:     20,
						Window:    time.Hour,
						KeyFields: []string{"ip"},
					}))
					r.Post("/", rt.AllergyHandler.AddAllergy)
					r.Put("/{id}", rt.AllergyHandler.UpdateAllergy)
					r.Delete("/{id}", rt.AllergyHandler.DeleteAllergy)
				})
			})

			// Master fun facts reference list.
			// GET is public. Mutations (POST/PUT/DELETE) require admin role.
			//
			// Endpoints:
			// GET    /api/v1/reference/fun-facts
			// GET    /api/v1/reference/fun-facts/{id}
			// POST   /api/v1/reference/fun-facts
			// PUT    /api/v1/reference/fun-facts/{id}
			// DELETE /api/v1/reference/fun-facts/{id}
			r.Route("/fun-facts", func(r chi.Router) {
				r.Get("/", rt.FunFactHandler.ListFunFacts)
				r.Get("/{id}", rt.FunFactHandler.GetFunFact)

				r.Group(func(r chi.Router) {
					// AdminMiddleware only reads AuthContext out of the request
					// context; it does not authenticate. Without Handle above it,
					// every request here 401s before CheckAdminAccess is reached.
					r.Use(rt.AuthMiddleware.Handle)
					r.Use(rt.AdminMiddleware.Handle)
					r.Use(rt.limit(appmw.RateLimitRule{
						Name:      "ref_funfact_mutation_ip",
						Limit:     20,
						Window:    time.Hour,
						KeyFields: []string{"ip"},
					}))
					r.Post("/", rt.FunFactHandler.CreateFunFact)
					r.Put("/{id}", rt.FunFactHandler.UpdateFunFact)
					r.Delete("/{id}", rt.FunFactHandler.DeleteFunFact)
				})
			})
		})

		// Submitting feedback is open to anonymous callers — a user who cannot
		// sign in or finish onboarding is exactly the user with feedback to
		// give, and feedback.user_id is nullable for that reason. OptionalAuth
		// still attributes the feedback when a signed-in user submits it.
		// Reading a specific feedback entry back stays behind full auth.
		//
		// Endpoints:
		// POST /api/v1/feedback
		// GET  /api/v1/feedback/{feedbackId}
		r.Route("/feedback", func(r chi.Router) {
			r.With(rt.AuthMiddleware.OptionalAuth).
				With(rt.limit(appmw.RateLimitRule{
					Name:      "feedback_submit_ip",
					Limit:     5,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.FeedbackHandler.SubmitFeedback)

			r.Group(func(r chi.Router) {
				r.Use(rt.AuthMiddleware.Handle)
				r.Use(rt.AuthMiddleware.RequireOnboardingCompleted)

				r.Get("/{feedbackId}", rt.FeedbackHandler.GetFeedback)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(rt.AuthMiddleware.Handle)
			r.Use(rt.AuthMiddleware.RequireOnboardingCompleted)

			// Default allowance for this authenticated group. Rate limits were
			// previously attached per route with .With(rt.limit(...)), so any
			// route added without that call was silently unlimited — 22 of them
			// were, including unthrottled PATCH writes and /dashboard, which
			// fans out to three concurrent queries per call. Applying it here
			// inverts the default to "limited unless deliberately raised".
			// Tighter per-route rules below still apply on top of this.
			r.Use(rt.limit(appmw.RateLimitRule{
				Name:      "authenticated_default_user",
				Limit:     300,
				Window:    time.Minute,
				KeyFields: []string{"user"},
			}))

			// Endpoints:
			// GET    /api/v1/users/me
			// PATCH  /api/v1/users/me
			// DELETE /api/v1/users/me
			// GET    /api/v1/users/me/notification-preferences
			// PATCH  /api/v1/users/me/notification-preferences
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", rt.UserHandler.GetMe)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "user_update_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Patch("/me", rt.UserHandler.UpdateMe)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:       "delete_account_identity",
					Limit:      3,
					Window:     10 * time.Minute,
					KeyFields:  []string{"ip"},
					BodyFields: []string{"recipient"},
				})).Delete("/me", rt.AuthHandler.DeleteAccount)

				r.Get("/me/notification-preferences", rt.UserHandler.GetNotificationPreferences)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "preferences_update_ip",
					Limit:     60,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Patch("/me/notification-preferences", rt.UserHandler.UpdateNotificationPreferences)
			})

		})

		// Emergency contacts gate on RequireActiveUser, NOT
		// RequireOnboardingCompleted. Creating the first emergency contact is
		// what *completes* onboarding, so gating it on completed onboarding
		// locks every new user out of the only route that can finish theirs.
		//
		// Endpoints:
		// GET    /api/v1/emergency-contacts
		// POST   /api/v1/emergency-contacts
		// PUT    /api/v1/emergency-contacts/{publicId}
		// DELETE /api/v1/emergency-contacts/{publicId}
		r.Group(func(r chi.Router) {
			r.Use(rt.AuthMiddleware.Handle)
			r.Use(rt.AuthMiddleware.RequireActiveUser)

			// Default allowance for this authenticated group. Rate limits were
			// previously attached per route with .With(rt.limit(...)), so any
			// route added without that call was silently unlimited — 22 of them
			// were, including unthrottled PATCH writes and /dashboard, which
			// fans out to three concurrent queries per call. Applying it here
			// inverts the default to "limited unless deliberately raised".
			// Tighter per-route rules below still apply on top of this.
			r.Use(rt.limit(appmw.RateLimitRule{
				Name:      "authenticated_default_user",
				Limit:     300,
				Window:    time.Minute,
				KeyFields: []string{"user"},
			}))

			r.Route("/emergency-contacts", func(r chi.Router) {
				r.Get("/", rt.EmergencyContactHandler.ListEmergencyContacts)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "contact_create_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.EmergencyContactHandler.CreateEmergencyContact)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "contact_update_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Put("/{publicId}", rt.EmergencyContactHandler.UpdateEmergencyContact)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "contact_delete_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Delete("/{publicId}", rt.EmergencyContactHandler.DeleteEmergencyContact)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(rt.AuthMiddleware.Handle)
			r.Use(rt.AuthMiddleware.RequireOnboardingCompleted)

			// Default allowance for this authenticated group. Rate limits were
			// previously attached per route with .With(rt.limit(...)), so any
			// route added without that call was silently unlimited — 22 of them
			// were, including unthrottled PATCH writes and /dashboard, which
			// fans out to three concurrent queries per call. Applying it here
			// inverts the default to "limited unless deliberately raised".
			// Tighter per-route rules below still apply on top of this.
			r.Use(rt.limit(appmw.RateLimitRule{
				Name:      "authenticated_default_user",
				Limit:     300,
				Window:    time.Minute,
				KeyFields: []string{"user"},
			}))

			// User personal allergy records.
			//
			// Endpoints:
			// GET    /api/v1/health/allergies
			// POST   /api/v1/health/allergies
			// DELETE /api/v1/health/allergies/{publicId}
			r.Route("/health/allergies", func(r chi.Router) {
				r.Get("/", rt.UserAllergyHandler.ListUserAllergies)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "allergy_add_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.UserAllergyHandler.AddUserAllergies)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "allergy_update_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Put("/{publicId}", rt.UserAllergyHandler.UpdateUserAllergy)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "allergy_delete_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Delete("/{publicId}", rt.UserAllergyHandler.DeleteUserAllergy)
			})

			// Future full-access routes:
			//
			r.Route("/medications", func(r chi.Router) {
				r.Get("/", rt.MedicationHandler.ListMedications)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "medication_create_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.MedicationHandler.CreateMedication)
				r.Get("/{publicId}", rt.MedicationHandler.GetMedication)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "medication_update_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Put("/{publicId}", rt.MedicationHandler.UpdateMedication)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "medication_delete_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Delete("/{publicId}", rt.MedicationHandler.DeleteMedication)
				r.Patch("/{publicId}/complete", rt.MedicationHandler.CompleteMedication)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "medication_adherence_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/{publicId}/adherence", rt.MedicationHandler.LogAdherence)
			})
			r.Route("/visits", func(r chi.Router) {
				r.Get("/", rt.VisitHandler.ListVisits)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "visit_create_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.VisitHandler.CreateVisit)
				r.Get("/{publicId}", rt.VisitHandler.GetVisit)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "visit_update_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Put("/{publicId}", rt.VisitHandler.UpdateVisit)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "visit_delete_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Delete("/{publicId}", rt.VisitHandler.DeleteVisit)
			})

			// Polled by the chat screen, so the limit is generous.
			r.With(rt.limit(appmw.RateLimitRule{
				Name:      "ai_quota_read_ip",
				Limit:     120,
				Window:    time.Hour,
				KeyFields: []string{"ip"},
			})).Get("/ai/quota", rt.AIHandler.GetQuota)

			r.Route("/ai/conversations", func(r chi.Router) {
				r.Get("/", rt.AIHandler.ListConversations)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "ai_conversation_create_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.AIHandler.CreateConversation)
				r.Get("/{publicId}", rt.AIHandler.GetConversation)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "ai_conversation_update_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Patch("/{publicId}", rt.AIHandler.UpdateConversation)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "ai_message_send_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/{publicId}/messages", rt.AIHandler.SendMessage)
				r.Patch("/{publicId}/archive", rt.AIHandler.ArchiveConversation)
			})

			r.Route("/drugs", func(r chi.Router) {
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "drug_verify_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/verify", rt.ScanHandler.VerifyDrug)
				r.Get("/scans", rt.ScanHandler.ListDrugScans)
				r.Get("/scans/{publicId}", rt.ScanHandler.GetDrugScan)
			})

			r.Get("/dashboard", rt.DashboardHandler.GetDashboard)

			r.Route("/support/tickets", func(r chi.Router) {
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "ticket_create_ip",
					Limit:     5,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/", rt.SupportTicketHandler.CreateTicket)
				r.Get("/", rt.SupportTicketHandler.ListTickets)
				r.Get("/{ticketId}", rt.SupportTicketHandler.GetTicket)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "ticket_message_ip",
					Limit:     20,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Post("/{ticketId}/messages", rt.SupportTicketHandler.AddMessage)
			})

			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", rt.NotificationHandler.ListNotifications)
				r.Get("/{notificationId}", rt.NotificationHandler.GetNotification)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "notification_read_ip",
					Limit:     30,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Patch("/{notificationId}/read", rt.NotificationHandler.MarkRead)
				r.With(rt.limit(appmw.RateLimitRule{
					Name:      "notification_read_all_ip",
					Limit:     10,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				})).Patch("/read-all", rt.NotificationHandler.MarkAllRead)
			})

			r.Group(func(r chi.Router) {
				r.Use(rt.AdminMiddleware.Handle)
				r.Get("/audit-logs", rt.AuditLogHandler.ListAuditLogs)
			})
		})
	})

	return rt.mux
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}

func (rt *Router) limit(rules ...appmw.RateLimitRule) func(http.Handler) http.Handler {
	// Validated even when rate limiting is disabled, so a bad rule is caught in
	// development rather than first taking effect in production.
	if err := appmw.ValidateRules(rules...); err != nil {
		panic(err)
	}

	if !rt.IsLive || rt.RateLimiter == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return rt.RateLimiter.Limit(rules...)
}
