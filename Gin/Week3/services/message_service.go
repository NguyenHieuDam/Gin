package services

import (
	"encoding/json"
	"errors"

	"WEEK3/models"

	"github.com/go-redis/redis/v8"
	"gorm.io/gorm"
)

type MessageService struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewMessageService(db *gorm.DB, redis *redis.Client) *MessageService {
	return &MessageService{db: db, redis: redis}
}

// SendMessage creates a new message
func (s *MessageService) SendMessage(userID uint, username, content, roomID string) (*models.MessageResponse, error) {
	// Validate content
	if len(content) == 0 || len(content) > 1000 {
		return nil, errors.New("nội dung tin nhắn không hợp lệ")
	}

	// Create message
	message := &models.Message{
		UserID:   userID,
		Username: username,
		Content:  content,
		RoomID:   roomID,
	}

	if err := s.db.Create(message).Error; err != nil {
		return nil, errors.New("lỗi khi gửi tin nhắn")
	}

	// Save to Redis for real-time access
	s.saveMessageToRedis(message)

	response := message.ToResponse()
	return &response, nil
}

// GetMessages retrieves messages for a room
func (s *MessageService) GetMessages(roomID string, limit, offset int) ([]models.MessageResponse, error) {
	var messages []models.Message
	query := s.db.Where("room_id = ?", roomID).Order("created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	if err := query.Find(&messages).Error; err != nil {
		return nil, errors.New("lỗi khi lấy tin nhắn")
	}

	var responses []models.MessageResponse
	for _, message := range messages {
		responses = append(responses, message.ToResponse())
	}

	return responses, nil
}

// GetRecentMessages retrieves recent messages for a room
func (s *MessageService) GetRecentMessages(roomID string, limit int) ([]models.MessageResponse, error) {
	return s.GetMessages(roomID, limit, 0)
}

// GetMessageByID retrieves a specific message
func (s *MessageService) GetMessageByID(id uint) (*models.MessageResponse, error) {
	var message models.Message
	if err := s.db.First(&message, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("không tìm thấy tin nhắn")
		}
		return nil, errors.New("lỗi khi lấy tin nhắn")
	}

	response := message.ToResponse()
	return &response, nil
}

// DeleteMessage deletes a message (only by the author)
func (s *MessageService) DeleteMessage(id, userID uint) error {
	var message models.Message
	if err := s.db.First(&message, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("không tìm thấy tin nhắn")
		}
		return errors.New("lỗi khi lấy tin nhắn")
	}

	// Check if user is the author
	if message.UserID != userID {
		return errors.New("bạn không có quyền xóa tin nhắn này")
	}

	if err := s.db.Delete(&message).Error; err != nil {
		return errors.New("lỗi khi xóa tin nhắn")
	}

	// Remove from Redis
	s.removeMessageFromRedis(id, message.RoomID)

	return nil
}

// saveMessageToRedis saves message to Redis for real-time access
func (s *MessageService) saveMessageToRedis(message *models.Message) {
	ctx := s.redis.Context()
	key := "messages:" + message.RoomID

	// Add to sorted set with timestamp as score
	score := float64(message.CreatedAt.Unix())
	member, _ := json.Marshal(message.ToResponse())

	s.redis.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	})

	// Keep only last 100 messages per room
	s.redis.ZRemRangeByRank(ctx, key, 0, -101)
}

// getMessagesFromRedis retrieves recent messages from Redis
func (s *MessageService) getMessagesFromRedis(roomID string, limit int) ([]models.MessageResponse, error) {
	ctx := s.redis.Context()
	key := "messages:" + roomID

	// Get recent messages from sorted set
	members, err := s.redis.ZRevRange(ctx, key, 0, int64(limit-1)).Result()
	if err != nil {
		return nil, err
	}

	var messages []models.MessageResponse
	for _, member := range members {
		var message models.MessageResponse
		if err := json.Unmarshal([]byte(member), &message); err == nil {
			messages = append(messages, message)
		}
	}

	return messages, nil
}

// removeMessageFromRedis removes message from Redis
func (s *MessageService) removeMessageFromRedis(messageID uint, roomID string) {
	ctx := s.redis.Context()
	key := "messages:" + roomID

	// Get all messages to find and remove the specific one
	members, err := s.redis.ZRange(ctx, key, 0, -1).Result()
	if err != nil {
		return
	}

	for _, member := range members {
		var message models.MessageResponse
		if err := json.Unmarshal([]byte(member), &message); err == nil {
			if message.ID == messageID {
				s.redis.ZRem(ctx, key, member)
				break
			}
		}
	}
}

// SearchMessages searches messages by content
func (s *MessageService) SearchMessages(roomID, query string, limit int) ([]models.MessageResponse, error) {
	var messages []models.Message
	err := s.db.Where("room_id = ? AND content ILIKE ?", roomID, "%"+query+"%").Order("created_at DESC").Limit(limit).Find(&messages).Error

	if err != nil {
		return nil, errors.New("lỗi khi tìm kiếm tin nhắn")
	}

	var responses []models.MessageResponse
	for _, message := range messages {
		responses = append(responses, message.ToResponse())
	}

	return responses, nil
}
