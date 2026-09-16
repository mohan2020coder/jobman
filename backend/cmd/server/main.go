package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jobman/backend/config"
	"github.com/jobman/backend/database"
	"github.com/jobman/backend/internal/auth"
	"github.com/jobman/backend/internal/business"
	"github.com/jobman/backend/internal/customers"
	"github.com/jobman/backend/internal/dashboard"
	"github.com/jobman/backend/internal/jobs"
	"github.com/jobman/backend/internal/payments"
	"github.com/jobman/backend/internal/receipts"
	"github.com/jobman/backend/internal/realtime"
	"github.com/jobman/backend/internal/technicians"
	"github.com/jobman/backend/internal/users"
	"github.com/jobman/backend/middleware"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := database.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()

	if err := database.NewMigrator(pool).Run(ctx); err != nil {
		log.Fatalf("migrations: %v", err)
	}
	log.Printf("database ready")

	if cfg.AppEnv != "production" {
		if err := database.SeedPasswords(ctx, pool, "password123"); err != nil {
			log.Printf("seed passwords: %v", err)
		}
	}

	h := buildHandler(cfg, pool)

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           h,
		ReadHeaderTimeout: 10 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("server listening on :%s", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server: %v", err)
		}
	}()

	<-ctx.Done()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)
	log.Printf("server stopped")
}

func buildHandler(cfg *config.Config, pool *pgxpool.Pool) http.Handler {
	usersRepo := users.NewRepository(pool)
	businessRepo := business.NewRepository(pool)
	customersRepo := customers.NewRepository(pool)
	techRepo := technicians.NewRepository(pool)
	paymentsRepo := payments.NewRepository(pool)
	jobsRepo := jobs.NewRepository(pool)
	receiptsRepo := receipts.NewRepository(pool)
	dashboardRepo := dashboard.NewRepository(pool)

	authSvc := auth.NewService(pool, usersRepo, businessRepo, auth.Config{
		JWTSecret:     []byte(cfg.JWTSecret),
		JWTExpiration: cfg.JWTExpiration,
	})
	businessSvc := business.NewService(businessRepo)
	customersSvc := customers.NewService(customersRepo)
	techSvc := technicians.NewService(pool, techRepo, usersRepo)
	paymentsSvc := payments.NewService(paymentsRepo)
	receiptsSvc := receipts.NewService(receiptsRepo)
	dashboardSvc := dashboard.NewService(dashboardRepo)

	receiptsSvc.SetJobDataProvider(&jobDataProvider{jobs: jobsRepo, payments: paymentsRepo})
	receiptsSvc.SetBusinessInfo(func(ctx context.Context, id string) (string, string, error) {
		b, err := businessRepo.GetByID(ctx, id)
		if err != nil {
			return "", "", err
		}
		phone := ""
		if b.Phone != nil {
			phone = *b.Phone
		}
		return b.Name, phone, nil
	})

	hub := realtime.NewHub()
	jobsSvc := jobs.NewService(pool, jobsRepo, paymentsRepo, techSvc, hub)
	jobsSvc.SetReceiptIssuer(&receiptsAdapter{inner: receipts.NewIssuer(receiptsRepo)})
	paymentsSvc.SetAssignmentChecker(jobsSvc.EnsureAssigned)

	authHandler := auth.NewHandler(authSvc)
	businessHandler := business.NewHandler(businessSvc)
	customersHandler := customers.NewHandler(customersSvc)
	techHandler := technicians.NewHandler(techSvc)
	jobsHandler := jobs.NewHandler(jobsSvc)
	paymentsHandler := payments.NewHandler(paymentsSvc)
	receiptsHandler := receipts.NewHandler(receiptsSvc)
	dashboardHandler := dashboard.NewHandler(dashboardSvc)

	customersHandler.SetJobsLister(func(ctx context.Context, businessID, customerID string) (any, error) {
		return jobsSvc.ListByCustomer(ctx, businessID, customerID)
	})
	techHandler.SetJobsLister(func(ctx context.Context, businessID, technicianID string) (any, error) {
		return jobsSvc.ListByTechnician(ctx, businessID, technicianID)
	})

	loginLimiter := middleware.NewRateLimiter(20, time.Minute)
	registerLimiter := middleware.NewRateLimiter(10, time.Minute)

	mux := http.NewServeMux()
	api := cfg.APIPrefix

	// --- Public routes ---
	mux.Handle("POST "+api+"/auth/register", middleware.RateLimit(registerLimiter, nil)(http.HandlerFunc(authHandler.Register)))
	mux.Handle("POST "+api+"/auth/login", middleware.RateLimit(loginLimiter, nil)(http.HandlerFunc(authHandler.Login)))
	mux.HandleFunc("GET "+api+"/ws", realtime.ServeWS(hub, techSvc, []byte(cfg.JWTSecret)))
	mux.HandleFunc("GET "+api+"/public/receipts/{token}", receiptsHandler.PublicJSON)
	mux.HandleFunc("GET /r/{token}", receiptsHandler.PublicPage)

	// --- Authenticated routes ---
	authed := http.NewServeMux()
	ro := func(roles ...string) func(http.HandlerFunc) http.HandlerFunc {
		if len(roles) == 0 {
			return func(h http.HandlerFunc) http.HandlerFunc { return h }
		}
		return func(h http.HandlerFunc) http.HandlerFunc {
			return middleware.RequireRole(roles...)(h).ServeHTTP
		}
	}

	authed.HandleFunc("GET "+api+"/auth/me", authHandler.Me)
	authed.HandleFunc("POST "+api+"/auth/logout", authHandler.Logout)
	authed.HandleFunc("GET "+api+"/business/me", businessHandler.Me)
	authed.HandleFunc("PUT "+api+"/business/me", ro("OWNER")(businessHandler.Update))

	authed.HandleFunc("GET "+api+"/customers", ro("OWNER", "ADMIN")(customersHandler.List))
	authed.HandleFunc("POST "+api+"/customers", ro("OWNER", "ADMIN")(customersHandler.Create))
	authed.HandleFunc("GET "+api+"/customers/{id}", ro("OWNER", "ADMIN")(customersHandler.Get))
	authed.HandleFunc("PUT "+api+"/customers/{id}", ro("OWNER", "ADMIN")(customersHandler.Update))
	authed.HandleFunc("DELETE "+api+"/customers/{id}", ro("OWNER", "ADMIN")(customersHandler.Delete))
	authed.HandleFunc("GET "+api+"/customers/{id}/jobs", ro("OWNER", "ADMIN")(customersHandler.Jobs))

	authed.HandleFunc("GET "+api+"/technicians", ro("OWNER", "ADMIN")(techHandler.List))
	authed.HandleFunc("POST "+api+"/technicians", ro("OWNER", "ADMIN")(techHandler.Create))
	authed.HandleFunc("GET "+api+"/technicians/{id}", ro("OWNER", "ADMIN")(techHandler.Get))
	authed.HandleFunc("PUT "+api+"/technicians/{id}", ro("OWNER", "ADMIN")(techHandler.Update))
	authed.HandleFunc("DELETE "+api+"/technicians/{id}", ro("OWNER", "ADMIN")(techHandler.Delete))
	authed.HandleFunc("GET "+api+"/technicians/{id}/jobs", ro("OWNER", "ADMIN")(techHandler.Jobs))

	authed.HandleFunc("GET "+api+"/jobs", ro("OWNER", "ADMIN")(jobsHandler.List))
	authed.HandleFunc("POST "+api+"/jobs", ro("OWNER", "ADMIN")(jobsHandler.Create))
	authed.HandleFunc("GET "+api+"/jobs/{id}", jobsHandler.Get)
	authed.HandleFunc("PUT "+api+"/jobs/{id}", ro("OWNER", "ADMIN")(jobsHandler.Update))
	authed.HandleFunc("DELETE "+api+"/jobs/{id}", ro("OWNER", "ADMIN")(jobsHandler.Delete))

	authed.HandleFunc("POST "+api+"/jobs/{id}/accept", ro("TECHNICIAN")(jobsHandler.Accept))
	authed.HandleFunc("POST "+api+"/jobs/{id}/on-the-way", ro("TECHNICIAN")(jobsHandler.OnTheWay))
	authed.HandleFunc("POST "+api+"/jobs/{id}/start", ro("TECHNICIAN")(jobsHandler.Start))
	authed.HandleFunc("POST "+api+"/jobs/{id}/complete", ro("TECHNICIAN")(jobsHandler.Complete))
	authed.HandleFunc("POST "+api+"/jobs/{id}/cancel", ro("OWNER", "ADMIN")(jobsHandler.CancelAction))

	authed.HandleFunc("GET "+api+"/jobs/{id}/payments", paymentsHandler.ListByJob)
	authed.HandleFunc("POST "+api+"/jobs/{id}/payments", paymentsHandler.Create)
	authed.HandleFunc("GET "+api+"/jobs/{id}/receipt", receiptsHandler.GetByJob)

	authed.HandleFunc("GET "+api+"/technician/jobs", ro("TECHNICIAN")(jobsHandler.MyJobs))
	authed.HandleFunc("GET "+api+"/dashboard/today", ro("OWNER", "ADMIN")(dashboardHandler.Today))

	authed.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		httpapiWriteJSON(w, 404, `{"error":{"code":"RESOURCE_NOT_FOUND","message":"Endpoint not found."}}`)
	})

	mux.Handle("/", middleware.RequireAuth([]byte(cfg.JWTSecret))(
		authed,
	))

	return middleware.Logging(middleware.Cors(cfg.CORSOrigins)(mux))
}

func httpapiWriteJSON(w http.ResponseWriter, status int, body string) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write([]byte(body))
}

// receiptsAdapter implements jobs.ReceiptIssuer over the receipts package.
type receiptsAdapter struct {
	inner *receipts.Issuer
}

func (a *receiptsAdapter) IssueTx(ctx context.Context, tx database.Querier, businessID, jobID string) (*jobs.IssuedReceipt, error) {
	rec, err := a.inner.IssueTx(ctx, tx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	return &jobs.IssuedReceipt{
		ID:            rec.ID,
		ReceiptNumber: rec.ReceiptNumber,
		PublicToken:   rec.PublicToken,
		IssuedAt:      rec.IssuedAt,
	}, nil
}

// jobDataProvider implements receipts.JobDataProvider over job + payment data.
type jobDataProvider struct {
	jobs     *jobs.Repository
	payments *payments.Repository
}

func (p *jobDataProvider) PublicJobData(ctx context.Context, businessID, jobID string) (*receipts.JobSnapshot, error) {
	job, err := p.jobs.GetByID(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	items, err := p.jobs.GetItems(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}
	paymentRows, err := p.payments.ListByJob(ctx, businessID, jobID)
	if err != nil {
		return nil, err
	}

	snap := &receipts.JobSnapshot{
		CustomerName:       job.Customer.Name,
		ServiceType:        job.ServiceType,
		ProblemDescription: job.ProblemDescription,
		FinalAmount:        job.FinalAmount,
		CompletedAt:        job.CompletedAt,
	}
	for _, it := range items {
		snap.Items = append(snap.Items, receipts.JobSnapshotItem{
			Description: it.Description,
			Quantity:    it.Quantity,
			TotalPrice:  it.TotalPrice,
		})
	}
	for _, pay := range paymentRows {
		snap.Payments = append(snap.Payments, receipts.PaymentLine{
			Amount: pay.Amount,
			Method: pay.Method,
			Status: pay.Status,
		})
	}
	return snap, nil
}