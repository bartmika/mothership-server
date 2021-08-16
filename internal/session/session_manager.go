package session

import (
	// "fmt"
	"context"
	"encoding/json"
	"github.com/go-redis/redis/v8"
	"time"

	"github.com/bartmika/mothership-server/internal/models"
)

type SessionManager struct {
	rdb *redis.Client
}

func New() *SessionManager {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // Default
		Password: "",               // no password set
		DB:       0,                // use default DB
	})
	return &SessionManager{
		rdb: rdb,
	}
}

func (sm *SessionManager) SaveUser(ctx context.Context, sessionUuid string, user *models.User, d time.Duration) error {
	userBin, err := json.Marshal(user)
	if err != nil {
		return err
	}
	err = sm.rdb.Set(ctx, sessionUuid, userBin, d).Err()
	if err != nil {
		return err
	}
	return nil
}

func (sm *SessionManager) GetUser(ctx context.Context, sessionUuid string) (*models.User, error) {
	userString, err := sm.rdb.Get(ctx, sessionUuid).Result()
	if err == redis.Nil {
		// fmt.Println("key2 does not exist")
		return nil, nil
	} else if err != nil {
		return nil, err
	} else {
		userBin := []byte(userString)
		user := &models.User{}
		err = json.Unmarshal(userBin, user)
		if user.Id == 0 {
			return nil, err
		}
		return user, err
	}
}
