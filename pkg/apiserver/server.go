package apiserver

import (
	"context"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"hornbill/pkg/auth"
	"hornbill/pkg/pb"
	"hornbill/pkg/rpcconn"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/swaggo/files"       // swagger embed files
	"github.com/swaggo/gin-swagger" // gin-swagger middleware
	_ "hornbill/docs"
)

type Daemon struct {
	Id           string           `json:"id"`
	Address      string           `json:"address"`
	Connection   *grpc.ClientConn `json:"-"`
	DaemonClient pb.DaemonClient  `json:"-"`
}

type DaemonList []Daemon

type Server struct {
	DaemonList        DaemonList
	GinServer         *gin.Engine
	AuthService       *auth.Service
	UserMaxTTLMinutes int64
}

func NewServer(addressList []string) (*Server, error) {
	server := Server{}

	userMaxTTLMinutesString := os.Getenv("USER_MAX_TTL_MINUTES")
	userMaxTTLMinutes, err := strconv.Atoi(userMaxTTLMinutesString)
	if err != nil {
		log.Fatalln("USER_MAX_TTL_MINUTES not providing or failed to parse", userMaxTTLMinutesString)
	}
	server.UserMaxTTLMinutes = int64(userMaxTTLMinutes)

	err = server.InitDaemonConnection(addressList)
	if err != nil {
		return nil, err
	}

	err = server.InitGin()
	if err != nil {
		return nil, err
	}

	return &server, nil
}

func (s *Server) GetDaemon(id string) *Daemon {
	for _, reg := range s.DaemonList {
		if reg.Id == id {
			return &reg
		}
	}
	return nil
}

func (s *Server) InitDaemonConnection(addressList []string) error {
	s.DaemonList = make([]Daemon, len(addressList))
	for i, address := range addressList {
		client, err := rpcconn.NewClient(address)
		if err != nil {
			return err
		}

		s.DaemonList[i] = Daemon{
			Id:           address,
			Address:      address,
			Connection:   client,
			DaemonClient: pb.NewDaemonClient(client),
		}
	}
	return nil
}

func (s *Server) InitGin() error {
	s.GinServer = gin.Default()
	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = strings.Split(os.Getenv("CORS_ALLOW_ORIGINS"), ",")
	corsConfig.AllowCredentials = true
	//s.GinServer.Use(cors.Default())
	s.GinServer.Use(cors.New(corsConfig))
	authRoutes := s.GinServer.Group("/auth")
	apiRoutes := s.GinServer.Group("/api")
	apiListenAddr := os.Getenv("API_LISTEN_ADDR")

	authService, err := auth.NewAuthService(context.TODO(), auth.LoadAuthServiceConfigEnv())
	s.AuthService = authService
	if err != nil {
		return err
	}
	authService.BindPaths(authRoutes)
	s.GinServer.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	s.GinServer.GET("/ready", ReadinessProbe)

	apiRoutes.Use(authService.Middleware())
	apiRoutes.POST("/login/:server", s.HandleDaemonLogin)
	apiRoutes.POST("/logout/:server", s.HandleDaemonLogout)
	apiRoutes.GET("/list", s.HandleDaemonList)
	apiRoutes.GET("/userinfo", authService.HandlePathInfo)

	err = s.GinServer.Run(apiListenAddr)
	if err != nil {
		return err
	}
	return nil
}

func (s *Server) PingAll(ctx context.Context) error {
	for _, daemonReg := range s.DaemonList {
		_, err := daemonReg.DaemonClient.Ping(ctx, &pb.Empty{})
		if err != nil {
			return fmt.Errorf("failed to ping %s: %v", daemonReg.Address, err)
		}
	}
	return nil
}
