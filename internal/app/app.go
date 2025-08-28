package app

import (
	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
	"google.golang.org/grpc"
	"log"
	"lootor_posts/gen/go/posts"
	"lootor_posts/internal/config"
	"lootor_posts/internal/core/grpcserver"
	"lootor_posts/internal/core/models"
	"lootor_posts/internal/core/repositories"
	"lootor_posts/internal/core/services"
	"lootor_posts/pkg/database"

	"net"
	"time"
)

type App struct {
	Echo *echo.Echo
}

func NewEchoApp(cfg *config.Config) (*App, error) {
	_ = godotenv.Load()

	e := echo.New()

	e.GET("/swagger/posts*", echoSwagger.WrapHandler)
	e.Server.MaxHeaderBytes = 1 << 20
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{
			"https://dev.lootor.me",
			"https://lootor.me",
			"https://www.lootor.me",
			"https://www.dev.lootor.me",
			"http://localhost:3000",
			"http://localhost:4173",
			"*",
		},
		AllowMethods: []string{
			echo.GET,
			echo.POST,
			echo.PUT,
			echo.DELETE,
			echo.OPTIONS,
		},
		AllowHeaders: []string{
			echo.HeaderOrigin,
			echo.HeaderContentType,
			echo.HeaderAccept,
			echo.HeaderAuthorization,
			"X-Requested-With",
		},
		AllowCredentials: true,
		MaxAge:           86400,
	}))

	// db init
	db, err := database.InitDB(&cfg.Database)
	if err != nil {
		if err = database.Reconnect(&cfg.Database); err != nil {
			log.Printf("Reconnection failed: %v", err)
		}
	}

	postsRepo := repositories.NewPostsRepository(db)
	reactRepo := repositories.NewPostReactionsRepository(db)

	err = db.AutoMigrate(&models.Posts{}, &models.PostReactions{})
	if err != nil {
		return nil, err
	}

	postsService := services.NewPostsService(postsRepo, reactRepo)

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 60 * time.Second,
	}))

	grpcServer := grpcserver.NewPostsService(postsService)

	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatal("failed to listen:", err)
	}

	s := grpc.NewServer()
	posts.RegisterPostsServiceServer(s, grpcServer)

	log.Println("Starting gRPC server on port 50051")
	if err := s.Serve(lis); err != nil {
		log.Fatal("failed to serve:", err)
	}

	return &App{Echo: e}, nil
}
