package bootstrap

import (
	"context"
	"net"
	"net/http"
	"user-service/config"
	delivery_grpc "user-service/delivery/grpc"
	"user-service/interceptor"
	"user-service/pkg"
	userpb "user-service/proto/v1"
	"user-service/repository"
	"user-service/usecase"

	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

type App struct {
	server *http.Server
	jwtService *pkg.JWTService
	grpcServer *grpc.Server
	cfg *config.Config
	pgxPool *pgxpool.Pool
}
func NewApp(cfg *config.Config) *App{
	jwt := pkg.NewJWTService(cfg.JWTConfig.Secret, cfg.JWTConfig.AccessTTL, cfg.JWTConfig.RefreshTTL)
	app := &App{
		cfg: cfg,
		jwtService: jwt ,
	}
	return app
}
func (a *App) SetupGRPC(ctx context.Context) error {
	
	publicMethod := map[string]bool{
		"/proto.v1.UserService/Register": true,
		"/proto.v1.UserService/Login": true,
		"/proto.v1.UserService/RefreshAccessToken": true,
	}
	publicRoles := map[string][]string{
		"/proto.v1.UserService/Admin" : {"admin"},
	}
	a.grpcServer = grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			interceptor.RecoveryInterceptor(),
			interceptor.AuthInterceptor(a.jwtService, publicMethod),
			interceptor.RBACInterceptor(publicRoles),
		),
	)
	if a.cfg.AppConfig.Env == "development" {
		reflection.Register(a.grpcServer)
	}
	err := a.initDatabase(ctx)
	if err != nil {
		return err
	}
	userRepo := repository.NewUserRepository(a.pgxPool)
	uc := usecase.NewUserUsecase(userRepo, a.jwtService)
	uh := delivery_grpc.NewUserHandler(uc)
	userpb.RegisterUserServiceServer(a.grpcServer, uh)
	return nil
}
func (a *App) Run(lis net.Listener) error {
	if err := a.grpcServer.Serve(lis); err != nil {
		return err
	}
	return nil
}

func (a *App) GracefulStop(){
	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}
	if a.pgxPool != nil {
		a.pgxPool.Close()
	}
}

func (a *App) Stop(){
	if a.grpcServer != nil {
		a.grpcServer.Stop()
	}
	if a.pgxPool != nil {
		a.pgxPool.Close()
	}
}
func (a *App) initDatabase(ctx context.Context) error{
	if a.pgxPool != nil {
		return nil
	}
	var err error
	a.pgxPool, err = config.ConnectDatabase(ctx, a.cfg)
	if err != nil {
		return err
	}
	return nil
}