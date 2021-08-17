package controllers

import (
	"context"
	"errors"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/golang/protobuf/ptypes/empty"
	tspb "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/google/uuid"
	"github.com/nakabonne/tstorage"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/bartmika/mothership-server/internal/models"
	"github.com/bartmika/mothership-server/internal/utils"
	pb "github.com/bartmika/mothership-server/proto"
)

func (s *Controller) Register(ctx context.Context, in *pb.RegistrationReq) (*pb.RegistrationRes, error) {
	doesExist, err := s.tenantRepo.CheckIfExistsByName(ctx, in.Company)
	if err != nil {
		return nil, err
	}
	if doesExist {
		return nil, errors.New("Company name is not unique")
	}

	// Check to see if the email exists and if it does then return error,
	// else continue with the registration.
	doesExist, err = s.userRepo.CheckIfExistsByEmail(ctx, in.Email)
	if err != nil {
		return nil, err
	}
	if doesExist {
		return nil, errors.New("Email is not unique")
	}

	t := &models.Tenant{
		Uuid:         uuid.NewString(),
		Name:         in.Company,
		State:        models.TenantActiveState,
		Timezone:     in.Timezone,
		CreatedTime:  time.Now(),
		ModifiedTime: time.Now(),
	}
	err = s.tenantRepo.Insert(ctx, t)
	if err != nil {
		return nil, err
	}

	t, err = s.tenantRepo.GetByUuid(ctx, t.Uuid)

	passwordPlain := strings.TrimSpace(in.Password)
	passwordHash, err := utils.HashPassword(passwordPlain)
	if err != nil {
		return nil, err
	}

	u := &models.User{
		TenantId:          t.Id,
		Uuid:              uuid.NewString(),
		Email:             in.Email,
		FirstName:         in.FirstName,
		LastName:          in.LastName,
		State:             models.UserActiveState,
		Timezone:          in.Timezone,
		CreatedTime:       time.Now(),
		ModifiedTime:      time.Now(),
		PasswordHash:      passwordHash,
		PasswordAlgorithm: "bcrypt",
		RoleId:            models.UserAdminRoleId,
	}
	err = s.userRepo.Insert(ctx, u)
	if err != nil {
		return nil, err
	}

	//TODO: WRITE A NICE EXPLANATION
	partitionDuration := time.Duration(24) * time.Hour
	writeTimeout := time.Duration(60) * time.Second
	tid := strconv.FormatUint(t.Id, 10)
	storage, _ := tstorage.NewStorage(
		tstorage.WithDataPath("tsdb/"+tid),
		tstorage.WithTimestampPrecision(tstorage.Seconds),
		tstorage.WithPartitionDuration(partitionDuration),
		tstorage.WithWriteTimeout(writeTimeout),
	)
	s.storageMap[t.Id] = storage
	log.Println("TSDB ready for tenant id #", t.Id)

	return &pb.RegistrationRes{
		Message: "You have been successfully registered. Please login to begin using the system.",
	}, nil
}

func (s *Controller) Login(ctx context.Context, in *pb.LoginReq) (*pb.LoginRes, error) {
	email := strings.TrimSpace(in.Email)
	passwordPlain := strings.TrimSpace(in.Password)

	user, err := s.userRepo.GetByEmail(ctx, email)
	if err != nil {
		return nil, errors.New("Email or password are incorrect")
	}
	if user == nil {
		return nil, errors.New("Email or password are incorrect")
	}

	if utils.CheckPasswordHash(passwordPlain, user.PasswordHash) == false {
		return nil, errors.New("Email or password are incorrect")
	}

	sessionUuid := uuid.NewString()
	sessionExpiryTime := time.Hour * 24 * 7 // 1 week

	err = s.manager.SaveUser(ctx, sessionUuid, user, sessionExpiryTime)
	if err != nil {
		return nil, err
	}

	b := []byte(s.hmacSecret)
	accessToken, refreshToken, err := utils.GenerateJWTTokenPair(b, sessionUuid, sessionExpiryTime)

	return &pb.LoginRes{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (s *Controller) RefreshToken(ctx context.Context, in *pb.RefreshTokenReq) (*pb.RefreshTokenRes, error) {
	return &pb.RefreshTokenRes{}, status.Errorf(codes.Unimplemented, "TODO: implement")
}

func (s *Controller) InsertTimeSeriesDatum(ctx context.Context, in *pb.TimeSeriesDatumReq) (*empty.Empty, error) {
	// Get our authenticated user.
	user := ctx.Value("user").(*models.User)

	// Lookup the dedicated time-series storage instance for our particular tenant.
	storage := s.storageMap[user.TenantId]

	// Generate our labels, if there are any.
	labels := []tstorage.Label{}
	for _, label := range in.Labels {
		labels = append(labels, tstorage.Label{Name: label.Name, Value: label.Value})
	}

	// Generate our datapoint.
	dataPoint := tstorage.DataPoint{Timestamp: in.Timestamp.Seconds, Value: in.Value}

	err := storage.InsertRows([]tstorage.Row{
		{
			Metric:    in.Metric,
			Labels:    labels,
			DataPoint: dataPoint,
		},
	})

	return &empty.Empty{}, err
}

// Utility function which will return the `User` stored in our session for the
// associated `access token` which was provided by the incoming context
// (which was included by the client when making the `InsertTimeSeriesData`
// RPC request). In theory this function could be removed and somesort of
// `ClientSreamInterceptor` be writte instead but for now use this utility
// function.
func (s *Controller) getUserFromInsertTimeSeriesDataMiddleware(stream pb.Mothership_InsertTimeSeriesDataServer) (*models.User, error) {
	ctx, cancel := context.WithCancel(stream.Context())
	defer cancel()

	md, ok := metadata.FromIncomingContext(ctx) // get context from stream
	if !ok {
		return nil, status.Errorf(codes.InvalidArgument, "Retrieving metadata is failed")
    }

	authHeader, ok := md["authorization"]
	if !ok {
		return nil, status.Errorf(codes.Unauthenticated, "Authorization token is not supplied")
    }

	token := authHeader[0]

	sessionUuid, err := utils.ProcessBearerToken([]byte(s.hmacSecret), token)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	user, err := s.manager.GetUser(ctx, sessionUuid)
	if err != nil {
		return nil, status.Errorf(codes.Internal, err.Error())
	}

	return user, nil
}

func (s *Controller) InsertTimeSeriesData(stream pb.Mothership_InsertTimeSeriesDataServer) error {
	// Extract the `User` account associated with the incoming context 'access
	// token' value sent by the client with this RPC request.
    user, err := s.getUserFromInsertTimeSeriesDataMiddleware(stream)
	if err != nil {
		return err
	}

	// Lookup the dedicated time-series storage instance for our particular tenant.
	storage := s.storageMap[user.TenantId]

	// DEVELOPERS NOTE:
	// If you don't understand how server side streaming works using gRPC then
	// please visit the documentation to get an understanding:
	// https://grpc.io/docs/languages/go/basics/#server-side-streaming-rpc-1

	// Wait and receieve the stream from the client.
	for {
		datum, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&empty.Empty{})
		}
		if err != nil {
			return err
		}

		// Generate our labels, if there are any.
		labels := []tstorage.Label{}
		for _, label := range datum.Labels {
			labels = append(labels, tstorage.Label{Name: label.Name, Value: label.Value})
		}

		// Generate our datapoint.
		dataPoint := tstorage.DataPoint{Timestamp: datum.Timestamp.Seconds, Value: datum.Value}

		err = storage.InsertRows([]tstorage.Row{
			{
				Metric:    datum.Metric,
				Labels:    labels,
				DataPoint: dataPoint,
			},
		})
	}

	return nil
}

func (s *Controller) InsertBulkTimeSeriesData(ctx context.Context, in *pb.BulkTimeSeriesDataReq) (*empty.Empty, error) {
	// Get our authenticated user.
	user := ctx.Value("user").(*models.User)

	// Lookup the dedicated time-series storage instance for our particular tenant.
	storage := s.storageMap[user.TenantId]

	for _, datum := range in.Data {
		// Generate our labels, if there are any.
		labels := []tstorage.Label{}
		for _, label := range datum.Labels {
			labels = append(labels, tstorage.Label{Name: label.Name, Value: label.Value})
		}

		// Generate our datapoint.
		dataPoint := tstorage.DataPoint{Timestamp: datum.Timestamp.Seconds, Value: datum.Value}

		err := storage.InsertRows([]tstorage.Row{
			{
				Metric:    datum.Metric,
				Labels:    labels,
				DataPoint: dataPoint,
			},
		})
		if err != nil {
			return &empty.Empty{}, err
		}
	}

	return &empty.Empty{}, nil
}

func (s *Controller) SelectBulkTimeSeriesData(ctx context.Context, in *pb.FilterReq) (*pb.SelectBulkRes, error) {
	// Get our authenticated user.
	user := ctx.Value("user").(*models.User)

	// Lookup the dedicated time-series storage instance for our particular tenant.
	storage := s.storageMap[user.TenantId]

    // The results variable to return.
	results := []*pb.DataPointRes{}

	// Generate our labels, if there are any.
	labels := []tstorage.Label{}
	for _, label := range in.Labels {
		labels = append(labels, tstorage.Label{Name: label.Name, Value: label.Value})
	}

	points, err := storage.Select(in.Metric, labels, in.Start.Seconds, in.End.Seconds)
	if err != nil {
		log.Println("SelectTimeSeriesData | storage.Select | err", err)
		return &pb.SelectBulkRes{DataPoints: results}, nil
	}

	for _, point := range points {
		ts := &tspb.Timestamp{
			Seconds: point.Timestamp,
			Nanos:   0,
		}
		dataPoint := &pb.DataPointRes{Value: point.Value, Timestamp: ts}
		results = append(results, dataPoint)
	}

	return &pb.SelectBulkRes{DataPoints: results}, nil
}
