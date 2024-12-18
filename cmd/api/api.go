package api

import (
	"database/sql"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/sikozonpc/ecom/service/cart"
	"github.com/sikozonpc/ecom/service/order"
	"github.com/sikozonpc/ecom/service/page"
	"github.com/sikozonpc/ecom/service/rating"
	"github.com/sikozonpc/ecom/service/search"
	"github.com/sikozonpc/ecom/service/student"
	"github.com/sikozonpc/ecom/service/teacher"
	"github.com/sikozonpc/ecom/service/user"
	"github.com/sikozonpc/ecom/types"
	// "github.com/sikozonpc/ecom/docs"
)

type APIServer struct {
	addr string
	db   *sql.DB
}


func NewAPIServer(addr string, db *sql.DB) *APIServer {
    return &APIServer{
        addr: addr,
        db:   db,
    }
}


func (s *APIServer) Start() error {
	router := mux.NewRouter()
	router.Use(LoggingMiddiware)
	subrouter := router.PathPrefix("/api/v1").Subrouter()



	// Registering user routes
	techStore := teacher.NewStore(s.db)
	userStore := user.NewStore(s.db) 
	userHandler := user.NewHandler(userStore, s.db, techStore)
	userHandler.AuthRoutes(subrouter)


	// Registering teacher routes
	teacherStore := teacher.NewStore(s.db)
	teacherHandler := teacher.NewHandler(teacherStore, userStore)
	teacherHandler.TeachRoutes(subrouter)

	// Registering the search routes
	searchStore := search.NewStore(s.db)
	searchHandler := search.NewHandler(searchStore)
	searchHandler.SearchRoutes(subrouter)

	// Registering the cart routes
	cartStore := cart.NewStore(s.db)
	cartHandler := cart.NewHandler(cartStore, userStore)
	cartHandler.CartRoutes(subrouter)

	// Registering the order routes
	orderStore := order.NewStore(s.db)
	orderHandler := order.NewHandler(orderStore, userStore, cartStore)
	orderHandler.OrderRoutes(subrouter)

	// Registering the Student routes
	studentStore := student.NewStore(s.db)
	studentHandler := student.NewHandler(studentStore, userStore)
	studentHandler.StudentRoutes(subrouter)

	// Registering the rating routes
	ratingStore := rating.NewStore(s.db)
	ratingHandler := rating.NewHandler(ratingStore, userStore)
	ratingHandler.RatingRoutes(subrouter)

	// Registering the page routes
	pageStore := page.NewStore(s.db)
	pageHandler := page.NewHandler(pageStore, userStore, teacherStore, ratingStore)
	pageHandler.PageRoutes(subrouter)

	log.Println("Starting On ", s.addr)
	return http.ListenAndServe(s.addr, router)
}


func LoggingMiddiware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		status := types.StatusRecorder{
			ResponseWriter: w,
            StatusCode:     http.StatusOK,
		}
		wrappedWriter := &status
		next.ServeHTTP(wrappedWriter, r)

		log.Printf("Method: %s, Endpoint: %s, Status: %d, Duration: %s", r.Method, r.URL.Path, wrappedWriter.StatusCode, time.Since(start))
	})
}