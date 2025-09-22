package services

import (
	"context"
	"errors"
	"time"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
	"WEEK3/models"
)

type PresenceService struct {
	db    *gorm.DB
	redis *redis.Client
	ctx   context.Context
}

func NewPresenceService(db *gorm.DB, redis *redis.Client) *PresenceService {
	return &PresenceService{
		db:    db,
		redis: redis,
		ctx:   context.Background(),
	}
}

// SetUserOnline marks a user as online
func (s *PresenceService) SetUserOnline(userID uint, username, roomID string) error {
	// Update database
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_online": true,
		"last_seen": time.Now(),
	}).Error; err != nil {
		return errors.New("lỗi khi cập nhật trạng thái người dùng")
	}

	// Set in Redis with expiration
	key := "online_users:" + roomID
	field := string(rune(userID))
	
	// Add user to online set
	s.redis.SAdd(s.ctx, key, field)
	
	// Set user info in hash
	userInfoKey := "user_info:" + string(rune(userID))
	s.redis.HMSet(s.ctx, userInfoKey, map[string]interface{}{
		"username": username,
		"room_id":  roomID,
		"joined_at": time.Now().Unix(),
	})

	// Set expiration for both keys (5 minutes)
	s.redis.Expire(s.ctx, key, 5*time.Minute)
	s.redis.Expire(s.ctx, userInfoKey, 5*time.Minute)

	return nil
}

// SetUserOffline marks a user as offline
func (s *PresenceService) SetUserOffline(userID uint, roomID string) error {
	// Update database
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"is_online": false,
		"last_seen": time.Now(),
	}).Error; err != nil {
		return errors.New("lỗi khi cập nhật trạng thái người dùng")
	}

	// Remove from Redis
	key := "online_users:" + roomID
	field := string(rune(userID))
	s.redis.SRem(s.ctx, key, field)

	// Remove user info
	userInfoKey := "user_info:" + string(rune(userID))
	s.redis.Del(s.ctx, userInfoKey)

	return nil
}

// GetOnlineUsersInRoom returns online users in a specific room
func (s *PresenceService) GetOnlineUsersInRoom(roomID string) ([]models.UserResponse, error) {
	key := "online_users:" + roomID
	members, err := s.redis.SMembers(s.ctx, key).Result()
	if err != nil {
		return nil, errors.New("lỗi khi lấy danh sách người dùng trực tuyến")
	}

	var users []models.UserResponse
	for _, member := range members {
		userID := uint(member[0]) // Convert first character to uint
		userInfoKey := "user_info:" + string(member[0])
		
		userInfo, err := s.redis.HGetAll(s.ctx, userInfoKey).Result()
		if err == nil && len(userInfo) > 0 {
			user := models.UserResponse{
				ID:       userID,
				Username: userInfo["username"],
				IsOnline: true,
			}
			users = append(users, user)
		}
	}

	return users, nil
}

// GetOnlineUsersCount returns the count of online users in a room
func (s *PresenceService) GetOnlineUsersCount(roomID string) (int, error) {
	key := "online_users:" + roomID
	count, err := s.redis.SCard(s.ctx, key).Result()
	if err != nil {
		return 0, errors.New("lỗi khi đếm số người dùng trực tuyến")
	}
	return int(count), nil
}

// UpdateUserHeartbeat updates user's last activity time
func (s *PresenceService) UpdateUserHeartbeat(userID uint, roomID string) error {
	// Update last seen in database
	if err := s.db.Model(&models.User{}).Where("id = ?", userID).Update("last_seen", time.Now()).Error; err != nil {
		return errors.New("lỗi khi cập nhật thời gian hoạt động")
	}

	// Extend Redis expiration
	key := "online_users:" + roomID
	field := string(rune(userID))
	
	// Check if user is still in the set
	exists, err := s.redis.SIsMember(s.ctx, key, field).Result()
	if err != nil {
		return err
	}

	if exists {
		// Extend expiration
		s.redis.Expire(s.ctx, key, 5*time.Minute)
		
		userInfoKey := "user_info:" + string(rune(userID))
		s.redis.Expire(s.ctx, userInfoKey, 5*time.Minute)
	}

	return nil
}

// CleanupExpiredUsers removes users who haven't sent heartbeat
func (s *PresenceService) CleanupExpiredUsers() error {
	// This would typically be called by a background job
	// For now, Redis TTL handles expiration automatically
	return nil
}

// IsUserOnline checks if a user is online
func (s *PresenceService) IsUserOnline(userID uint, roomID string) (bool, error) {
	key := "online_users:" + roomID
	field := string(rune(userID))
	
	exists, err := s.redis.SIsMember(s.ctx, key, field).Result()
	if err != nil {
		return false, errors.New("lỗi khi kiểm tra trạng thái người dùng")
	}

	return exists, nil
}

// GetUserLastSeen returns user's last seen time
func (s *PresenceService) GetUserLastSeen(userID uint) (time.Time, error) {
	var user models.User
	if err := s.db.Select("last_seen").First(&user, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return time.Time{}, errors.New("không tìm thấy người dùng")
		}
		return time.Time{}, errors.New("lỗi khi lấy thông tin người dùng")
	}

	return user.LastSeen, nil
}
