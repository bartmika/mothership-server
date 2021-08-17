package controllers

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nakabonne/tstorage"
	"google.golang.org/grpc"

	"github.com/bartmika/mothership-server/internal/models"
	"github.com/bartmika/mothership-server/internal/repositories"
	"github.com/bartmika/mothership-server/internal/session"
	pb "github.com/bartmika/mothership-server/proto"
)

type Controller struct {
	ipAddress   string
	port        int
	databaseUrl string
	hmacSecret  string
	dbpool      *pgxpool.Pool
	manager     *session.SessionManager
	grpcServer  *grpc.Server
	tenantRepo  models.TenantRepository
	userRepo    models.UserRepository
	storageMap  map[uint64]tstorage.Storage
	pb.MothershipServer
}

func New(ipAddress string, port int, databaseUrl string, hmacSecret string) *Controller {
	dbpool, err := pgxpool.Connect(context.Background(), databaseUrl)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		os.Exit(1)
	}

	tenantRepo := repositories.NewTenantRepo(dbpool)
	userRepo := repositories.NewUserRepo(dbpool)

	return &Controller{
		ipAddress:   ipAddress,
		port:        port,
		databaseUrl: databaseUrl,
		hmacSecret:  hmacSecret,
		dbpool:      dbpool,
		tenantRepo:  tenantRepo,
		userRepo:    userRepo,
		manager:     session.New(),
		grpcServer:  nil,
	}
}

// Function will consume the main runtime loop and run the business logic
// of the application.
func (s *Controller) RunMainRuntimeLoop() {
	// Open a TCP server to the specified localhost and environment variable
	// specified port number.
	lis, err := net.Listen("tcp", fmt.Sprintf("%v:%v", s.ipAddress, s.port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// Initialize our gRPC server using our TCP server.
	grpcServer := grpc.NewServer(
		withServerUnaryInterceptor(s),
	)

	// Save reference to our application state.
	s.grpcServer = grpcServer

	// For debugging purposes only.
	log.Printf("Server is running on port %v", s.port)

	//TODO: WRITE NICE EXPLANATION.
	storageMap := make(map[uint64]tstorage.Storage)
	tenantIds, err := s.tenantRepo.ListAllIds(context.Background())
	for _, tenantId := range tenantIds {
		tid := strconv.FormatUint(tenantId, 10)
		partitionDuration := time.Duration(24) * time.Hour
		writeTimeout := time.Duration(60) * time.Second
		storage, _ := tstorage.NewStorage(
			tstorage.WithDataPath("tsdb/"+tid),
			tstorage.WithTimestampPrecision(tstorage.Seconds),
			tstorage.WithPartitionDuration(partitionDuration),
			tstorage.WithWriteTimeout(writeTimeout),
		)
		storageMap[tenantId] = storage
		log.Println("TSDB ready for tenant id #" + tid)
	}
	s.storageMap = storageMap

	// Block the main runtime loop for accepting and processing gRPC requests.
	pb.RegisterMothershipServer(grpcServer, s)
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

// Function will tell the application to stop the main runtime loop when
// the process has been finished.
func (s *Controller) StopMainRuntimeLoop() {
	log.Printf("Starting graceful shutdown now...")

	// Shutdown our implementation sub-system.
	// Iterate through all the time-series data storage instances running.
	tenantIds, _ := s.tenantRepo.ListAllIds(context.Background())
	for _, tenantId := range tenantIds {
		// Finish our database operations running.
		storage := s.storageMap[tenantId]
		storage.Close()
		log.Printf("TSDB shutdown for tenant id #%v\n", tenantId)
	}

	// Finish our database operations running.
	defer s.dbpool.Close()

	// Finish any RPC communication taking place at the moment before
	// shutting down the gRPC server.
	s.grpcServer.GracefulStop()
}
