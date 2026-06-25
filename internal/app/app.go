package app

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Michealshodipo56/jarcade-backend/internal/auth"
	"github.com/Michealshodipo56/jarcade-backend/internal/config"
	"github.com/Michealshodipo56/jarcade-backend/internal/db"
	"github.com/Michealshodipo56/jarcade-backend/internal/handlers"
	"github.com/Michealshodipo56/jarcade-backend/internal/middleware"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func Run() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	store, err := db.New(cfg)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer store.Close()

	tokens := auth.NewTokenService(cfg.JWTSecret)
	authSvc := auth.NewService(store, tokens)
	authHandler := handlers.NewAuthHandler(authSvc, cfg.FrontendURL, cfg.ExposeResetLink, cfg.GoogleClientID)
	uploadHandler := handlers.NewUploadHandler(store)
	loginLimiter := middleware.NewLoginRateLimiter(10, 15*time.Minute)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(middleware.SecurityHeaders)
	r.Use(middleware.CORS(cfg))

	r.Get("/", handlers.Health)
	r.Get("/api/health", handlers.Health)

	r.Route("/api/auth", func(r chi.Router) {
		r.Post("/signup", authHandler.Signup)
		r.With(loginLimiter.Middleware).Post("/login", authHandler.Login)
		r.Post("/logout", authHandler.Logout)
		r.Post("/forgot-password", authHandler.ForgotPassword)
		r.Post("/reset-password", authHandler.ResetPassword)
		r.Post("/google", authHandler.GoogleLogin)
		r.With(middleware.Auth(tokens)).Get("/me", authHandler.Me)
	})

	r.Route("/api/uploads", func(r chi.Router) {
		r.Get("/", uploadHandler.List)
		r.With(middleware.Auth(tokens)).Get("/mine", uploadHandler.Mine)
		r.With(middleware.Auth(tokens)).Post("/", uploadHandler.Create)
		r.With(middleware.Auth(tokens)).Delete("/{id}", uploadHandler.Delete)
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       15 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	go func() {
		log.Printf("JARCADE API listening on :%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("shutdown: %v", err)
	}
}
