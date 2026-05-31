package rest

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chiMw "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/victorotene80/medilog-api/internal/interfaces/http/rest/handler"
	appmw "github.com/victorotene80/medilog-api/internal/interfaces/middleware"
)

type Router struct {
	mux *chi.Mux

	Logger                  *zap.Logger
	AuthMiddleware          *appmw.AuthMiddleware
	RateLimiter             *appmw.RateLimiter
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
}

func NewRouter(
	logger *zap.Logger,
	authMiddleware *appmw.AuthMiddleware,
	rateLimiter *appmw.RateLimiter,
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

) *Router {
	return &Router{
		mux:                     chi.NewRouter(),
		Logger:                  logger,
		AuthMiddleware:          authMiddleware,
		RateLimiter:             rateLimiter,
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
	}
}

func (rt *Router) Setup() http.Handler {
	rt.mux.Use(chiMw.RequestID)
	rt.mux.Use(chiMw.RealIP)
	rt.mux.Use(appmw.PanicRecovery(rt.Logger))
	rt.mux.Use(appmw.RequestMetadata)
	rt.mux.Use(chiMw.Logger)

	rt.mux.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"*"},
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

	// Real server health check.
	// Endpoint:
	// GET /health
	rt.mux.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	rt.mux.Route("/api/v1", func(r chi.Router) {
		r.Route("/auth", func(r chi.Router) {
			r.With(rt.limit(
				appmw.RateLimitRule{
					Name:      "auth_register_ip",
					Limit:     5,
					Window:    time.Hour,
					KeyFields: []string{"ip"},
				},
				appmw.RateLimitRule{
					Name:       "auth_register_identity",
					Limit:      3,
					Window:     time.Hour,
					KeyFields:  []string{"ip"},
					BodyFields: []string{"email", "phone"},
				},
			)).Post("/register", rt.AuthHandler.CreateUser)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:       "auth_login_identity",
				Limit:      5,
				Window:     10 * time.Minute,
				KeyFields:  []string{"ip", "device"},
				BodyFields: []string{"email", "phone"},
			})).Post("/login", rt.AuthHandler.Login)

			r.With(rt.limit(appmw.RateLimitRule{
				Name:      "auth_google_ip",
				Limit:     10,
				Window:    10 * time.Minute,
				KeyFields: []string{"ip", "device"},
			})).Post("/google", rt.AuthHandler.GoogleLogin)

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
				Limit:      50,
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
				KeyFields:  []string{"ip", "device"},
				BodyFields: []string{"recipient", "purpose"},
			})).Post("/otp/verify-onboarding", rt.OTPHandler.VerifyOnboardingOTP)

			r.Group(func(r chi.Router) {
				r.Use(rt.AuthMiddleware.Handle)

				r.Post("/logout", rt.AuthHandler.Logout)
				r.Post("/change-password", rt.AuthHandler.ChangePassword)
			})
		})

		r.Route("/reference", func(r chi.Router) {
			r.Get("/countries", rt.ReferenceHandler.ListCountries)

			// Master allergy reference list.
			// GET is public. Mutations (POST/PUT/DELETE) are auth-protected —
			// swap AuthMiddleware.Handle for an admin middleware once it exists.
			//
			// Endpoints:
			// GET    /api/v1/reference/allergies
			// POST   /api/v1/reference/allergies
			// PUT    /api/v1/reference/allergies/{id}
			// DELETE /api/v1/reference/allergies/{id}
			r.Route("/allergies", func(r chi.Router) {
				r.Get("/", rt.AllergyHandler.ListAllergies)

				r.Group(func(r chi.Router) {
					r.Use(rt.AuthMiddleware.Handle)
					r.Post("/", rt.AllergyHandler.AddAllergy)
					r.Put("/{id}", rt.AllergyHandler.UpdateAllergy)
					r.Delete("/{id}", rt.AllergyHandler.DeleteAllergy)
				})
			})

			// Master fun facts reference list.
			// GET is public. Mutations (POST/PUT/DELETE) are auth-protected —
			// swap AuthMiddleware.Handle for an admin middleware once it exists.
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
					r.Use(rt.AuthMiddleware.Handle)
					r.Post("/", rt.FunFactHandler.CreateFunFact)
					r.Put("/{id}", rt.FunFactHandler.UpdateFunFact)
					r.Delete("/{id}", rt.FunFactHandler.DeleteFunFact)
				})
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(rt.AuthMiddleware.Handle)

			// Endpoint:
			// GET /api/v1/users/me
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", rt.UserHandler.GetMe)
			})

			// Endpoint:
			// POST /api/v1/emergency-contacts
			r.Route("/emergency-contacts", func(r chi.Router) {
				r.Post("/", rt.EmergencyContactHandler.CreateEmergencyContact)
			})
		})

		r.Group(func(r chi.Router) {
			r.Use(rt.AuthMiddleware.Handle)
			r.Use(rt.AuthMiddleware.RequireOnboardingCompleted)

			// User personal allergy records.
			//
			// Endpoints:
			// GET    /api/v1/health/allergies
			// POST   /api/v1/health/allergies
			// DELETE /api/v1/health/allergies/{publicId}
			r.Route("/health/allergies", func(r chi.Router) {
				r.Get("/", rt.UserAllergyHandler.ListUserAllergies)
				r.Post("/", rt.UserAllergyHandler.AddUserAllergies)
				r.Delete("/{publicId}", rt.UserAllergyHandler.DeleteUserAllergy)
			})

			// Future full-access routes:
			//
			r.Route("/medications", func(r chi.Router) {
				r.Get("/", rt.MedicationHandler.ListMedications)
				r.Post("/", rt.MedicationHandler.CreateMedication)
				r.Get("/{publicId}", rt.MedicationHandler.GetMedication)
				r.Put("/{publicId}", rt.MedicationHandler.UpdateMedication)
				r.Delete("/{publicId}", rt.MedicationHandler.DeleteMedication)
				r.Patch("/{publicId}/complete", rt.MedicationHandler.CompleteMedication)
				r.Post("/{publicId}/adherence", rt.MedicationHandler.LogAdherence)
			})
			r.Route("/visits", func(r chi.Router) {
				r.Get("/", rt.VisitHandler.ListVisits)
				r.Post("/", rt.VisitHandler.CreateVisit)
				r.Get("/{publicId}", rt.VisitHandler.GetVisit)
				r.Put("/{publicId}", rt.VisitHandler.UpdateVisit)
				r.Delete("/{publicId}", rt.VisitHandler.DeleteVisit)
			})

			r.Route("/ai/conversations", func(r chi.Router) {
				r.Get("/", rt.AIHandler.ListConversations)
				r.Post("/", rt.AIHandler.CreateConversation)
				r.Get("/{publicId}", rt.AIHandler.GetConversation)
				r.Post("/{publicId}/messages", rt.AIHandler.SendMessage)
				r.Patch("/{publicId}/archive", rt.AIHandler.ArchiveConversation)
			})

			r.Route("/drugs", func(r chi.Router) {
				r.Post("/verify", rt.ScanHandler.VerifyDrug)
				r.Get("/scans", rt.ScanHandler.ListDrugScans)
				r.Get("/scans/{publicId}", rt.ScanHandler.GetDrugScan)
			})

			r.Get("/dashboard", rt.DashboardHandler.GetDashboard)
		})
	})

	return rt.mux
}

func (rt *Router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rt.mux.ServeHTTP(w, r)
}

func (rt *Router) limit(rules ...appmw.RateLimitRule) func(http.Handler) http.Handler {
	if rt.RateLimiter == nil {
		return func(next http.Handler) http.Handler {
			return next
		}
	}

	return rt.RateLimiter.Limit(rules...)
}
