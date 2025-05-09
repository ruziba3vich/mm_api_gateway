package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/go-delve/delve/pkg/config"
	"github.com/ruziba3vich/mm_api_getway/genprotos/genprotos/article_protos"
	"github.com/ruziba3vich/mm_api_getway/genprotos/genprotos/user_protos"
	handler "github.com/ruziba3vich/mm_api_getway/internal/http"
	logger "github.com/ruziba3vich/prodonik_lgger"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	app := fx.New(
		fx.Provide(
			config.LoadConfig,
			newLogger,
			newUserServiceClient,
			newArticleServiceClient,
			handler.NewHTTPHandler,
			registerRoutes,
		),
		fx.Invoke(registerHooks),
	)

	app.Run()
}

// Register application lifecycle hooks
func registerHooks(
	lc fx.Lifecycle,
	grpcServer *grpc.Server,
	cfg *config.Config,
) {
	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			log.Println("Starting logging service...")

			listener, err := net.Listen("tcp", fmt.Sprintf(":%s", "7878"))
			if err != nil {
				return fmt.Errorf("failed to listen on port %s: %s", "7878", err.Error())
			}

			log.Printf("gRPC server listening on port %s", "7878") // TODO: pass server port via config

			go func() {
				if err := grpcServer.Serve(listener); err != nil {
					log.Fatalf("Failed to serve gRPC: %v", err)
				}
			}()

			log.Println("Logging service started")
			return nil
		},
		OnStop: func(context.Context) error {
			log.Println("Stopping logging service...")

			grpcServer.GracefulStop()
			log.Println("Logging service stopped")
			return nil
		},
	})

	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
		<-signals

		log.Println("Received shutdown signal")
	}()
}

func newUserServiceClient(cfg *config.Config, logger *logger.Logger) (user_protos.UserServiceClient, error) {
	conn, err := grpc.NewClient("localhost:7373", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to Python Executor Service", map[string]any{"error": err})
		return nil, err
	}
	logger.Info("Connected to gRPC service", map[string]any{"address": "localhost:7373"}) // TODO: pass these data from config
	return user_protos.NewUserServiceClient(conn), nil
}

func newArticleServiceClient(cfg *config.Config, logger *logger.Logger) (article_protos.ArticleServiceClient, error) {
	conn, err := grpc.NewClient("localhost:7171", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		logger.Error("Failed to connect to Python Executor Service", map[string]any{"error": err})
		return nil, err
	}
	logger.Info("Connected to gRPC service", map[string]any{"address": "localhost:7171"}) // TODO: pass these data from config
	return article_protos.NewArticleServiceClient(conn), nil
}

func newLogger() (*logger.Logger, error) {
	return logger.NewLogger("/app/logs/article_service.log")
}

func registerRoutes(handler *handler.HTTPHandler, r *gin.Engine) {
	handler.RegisterRoutes(r)
}
