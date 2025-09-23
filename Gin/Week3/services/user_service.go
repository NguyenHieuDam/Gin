package services

import (
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"WEEK3/models"
)

type UserService struct {
	db *gorm.DB
}

func NewUserService(db *gorm.DB) *UserService {
	return &UserService{db: db}
}

// Register creates a new user
func (s *UserService) Register(req *models.UserRequest) (*models.UserResponse, error) {
	// Check if user already exists
	var existingUser models.User
	if err := s.db.Where("username = ? OR email = ?", req.Username, req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("username hoặc email đã tồn tại")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("lỗi khi mã hóa mật khẩu")
	}

	// Create user
	user := &models.User{
		Username:  req.Username,
		Email:     req.Email,
		Password:  string(hashedPassword),
		IsOnline:  false,
		LastSeen:  time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, errors.New("lỗi khi tạo người dùng")
	}

	response := user.ToResponse()
	return &response, nil
}

// Login authenticates a user
func (s *UserService) Login(req *models.LoginRequest) (*models.UserResponse, error) {
	var user models.User
	if err := s.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("tên đăng nhập hoặc mật khẩu không đúng")
		}
		return nil, errors.New("lỗi khi đăng nhập")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		return nil, errors.New("tên đăng nhập hoặc mật khẩu không đúng")
	}

    // Update online status and last seen
    now := time.Now()
    s.db.Model(&user).Updates(map[string]interface{}{
        "is_online": true,
        "last_seen": now,
    })

    // Reflect updates in the returned response
    user.IsOnline = true
    user.LastSeen = now
    response := user.ToResponse()
	return &response, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(id uint) (*models.UserResponse, error) {
	var user models.User
	if err := s.db.First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("không tìm thấy người dùng")
		}
		return nil, errors.New("lỗi khi lấy thông tin người dùng")
	}

	response := user.ToResponse()
	return &response, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserService) GetUserByUsername(username string) (*models.UserResponse, error) {
	var user models.User
	if err := s.db.Where("username = ?", username).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("không tìm thấy người dùng")
		}
		return nil, errors.New("lỗi khi lấy thông tin người dùng")
	}

	response := user.ToResponse()
	return &response, nil
}

// UpdateUserStatus updates user online status
func (s *UserService) UpdateUserStatus(id uint, isOnline bool) error {
	updates := map[string]interface{}{
		"is_online": isOnline,
		"last_seen": time.Now(),
	}

	if err := s.db.Model(&models.User{}).Where("id = ?", id).Updates(updates).Error; err != nil {
		return errors.New("lỗi khi cập nhật trạng thái người dùng")
	}

	return nil
}

// GetAllUsers returns all users
func (s *UserService) GetAllUsers() ([]models.UserResponse, error) {
	var users []models.User
	if err := s.db.Find(&users).Error; err != nil {
		return nil, errors.New("lỗi khi lấy danh sách người dùng")
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}

// GetOnlineUsers returns online users
func (s *UserService) GetOnlineUsers() ([]models.UserResponse, error) {
	var users []models.User
	if err := s.db.Where("is_online = ?", true).Find(&users).Error; err != nil {
		return nil, errors.New("lỗi khi lấy danh sách người dùng trực tuyến")
	}

	var responses []models.UserResponse
	for _, user := range users {
		responses = append(responses, user.ToResponse())
	}

	return responses, nil
}
