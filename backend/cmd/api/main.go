package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/joho/godotenv"

	"manufactpro/backend/internal/auth"
	"manufactpro/backend/internal/bom"
	"manufactpro/backend/internal/db"
	"manufactpro/backend/internal/material"
	"manufactpro/backend/internal/product"
	"manufactpro/backend/internal/procurement"
	"manufactpro/backend/internal/production"
	"manufactpro/backend/internal/reports"
	"manufactpro/backend/internal/stock"
	"manufactpro/backend/internal/supplier"
	syncpkg "manufactpro/backend/internal/sync"
	"manufactpro/backend/internal/user"
	"manufactpro/backend/internal/warehouse"
	"manufactpro/backend/pkg/response"
)

func main() {
	_ = godotenv.Load()

	ctx := context.Background()
	if err := db.Connect(ctx); err != nil {
		log.Fatalf("db connect: %v", err)
	}
	defer db.Pool.Close()

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:5173", "http://localhost:4173"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
	}))

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		response.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Mount("/api/auth", auth.Routes())
	r.Mount("/api/materials", material.Routes())
	r.Mount("/api/suppliers", supplier.Routes())
	r.Mount("/api/products", product.Routes())
	r.Mount("/api/variants", product.VariantRoutes())
	r.Mount("/api", warehouse.Routes())
	r.Mount("/api/users", user.Routes())

	// BOM — variant sub-routes share /api/variants prefix with product, use inline
	r.With(auth.JWTMiddleware).Get("/api/variants/{variantId}/bom", bom.HandleList)
	r.With(auth.JWTMiddleware).Post("/api/variants/{variantId}/bom", bom.HandleCreate)
	r.With(auth.JWTMiddleware).Put("/api/bom/{id}", bom.HandleUpdate)
	r.With(auth.JWTMiddleware).Delete("/api/bom/{id}", bom.HandleDelete)

	// Stock
	r.With(auth.JWTMiddleware).Get("/api/stock", stock.HandleCurrentStock)
	r.With(auth.JWTMiddleware).Get("/api/stock/movements", stock.HandleListMovements)
	r.With(auth.JWTMiddleware).Post("/api/stock/transfer", stock.HandleTransfer)
	r.With(auth.JWTMiddleware).Post("/api/stock/adjustment", stock.HandleAdjustment)

	// Reports
	r.With(auth.JWTMiddleware).Get("/api/reports/dashboard", reports.HandleDashboard)
	r.With(auth.JWTMiddleware).Get("/api/reports/hpp-margin", reports.HandleHppMargin)
	r.With(auth.JWTMiddleware).Get("/api/reports/production-trend", reports.HandleProductionTrend)

	// Procurement
	r.With(auth.JWTMiddleware).Get("/api/procurement/purchase-orders", procurement.HandleList)
	r.With(auth.JWTMiddleware).Post("/api/procurement/purchase-orders", procurement.HandleCreate)
	r.With(auth.JWTMiddleware).Get("/api/procurement/purchase-orders/{id}", procurement.HandleGetDetail)
	r.With(auth.JWTMiddleware).Patch("/api/procurement/purchase-orders/{id}/send", procurement.HandleSend)
	r.With(auth.JWTMiddleware).Post("/api/procurement/purchase-orders/{id}/receive", procurement.HandleReceive)
	r.With(auth.JWTMiddleware).Patch("/api/procurement/purchase-orders/{id}/cancel", procurement.HandleCancel)

	// Sync — offline queue replay
	r.With(auth.JWTMiddleware).Post("/api/sync", syncpkg.HandleSync)

	// Production
	r.With(auth.JWTMiddleware).Get("/api/production/orders", production.HandleList)
	r.With(auth.JWTMiddleware).Post("/api/production/orders", production.HandleCreate)
	r.With(auth.JWTMiddleware).Post("/api/production/orders/{id}/start", production.HandleStart)
	r.With(auth.JWTMiddleware).Post("/api/production/orders/{id}/output", production.HandleOutput)
	r.With(auth.JWTMiddleware).Patch("/api/production/orders/{id}/cancel", production.HandleCancel)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Printf("ManufactPro API listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
