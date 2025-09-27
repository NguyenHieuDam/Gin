# Ứng dụng Chat Real-time với Gin Golang

Ứng dụng trò chuyện thời gian thực được phát triển bằng Gin Golang theo mô hình MVC.

## Tính năng

- ✅ **Trò chuyện thời gian thực**: Người dùng có thể gửi và nhận tin nhắn theo thời gian thực
- ✅ **WebSockets**: Giao tiếp hai chiều nhanh chóng
- ✅ **Theo dõi sự hiện diện**: Hiển thị người dùng trực tuyến
- ✅ **Lưu trữ lịch sử tin nhắn**: Sử dụng Redis để lưu trữ tin nhắn
- ✅ **Giới hạn tốc độ**: Rate limiting để ngăn chặn spam
- ✅ **Giao diện đẹp**: UI hiện đại với Bootstrap và Vue.js

## Cấu trúc dự án

```
chat-app/
├── controllers/          # Controllers xử lý logic
│   ├── auth_controller.go
│   ├── message_controller.go
│   └── websocket_controller.go
├── models/              # Models định nghĩa cấu trúc dữ liệu
│   ├── user.go
│   ├── message.go
│   └── websocket.go
├── middleware/          # Middleware
│   └── rate_limiter.go
├── services/           # Services
│   └── redis_service.go
├── templates/          # HTML templates
│   └── index.html
├── main.go            # Entry point
├── go.mod            # Go modules
└── README.md         # Documentation
```

## Yêu cầu hệ thống

- Go 1.21+
- Redis (tùy chọn, ứng dụng vẫn hoạt động nếu không có Redis)

## Cài đặt và chạy

### 1. Clone repository
```bash
git clone <repository-url>
cd chat-app
```

### 2. Cài đặt dependencies
```bash
go mod tidy
```

### 3. Cài đặt Redis (tùy chọn)
```bash
# Ubuntu/Debian
sudo apt-get install redis-server

# macOS
brew install redis

# Windows
# Tải Redis từ https://redis.io/download
```

### 4. Chạy ứng dụng
```bash
go run main.go
```

Ứng dụng sẽ chạy trên `http://localhost:8080`

## Cấu hình môi trường

Bạn có thể cấu hình các biến môi trường sau:

```bash
export PORT=8080                    # Port server (mặc định: 8080)
export REDIS_ADDR=localhost:6379    # Redis address (mặc định: localhost:6379)
export REDIS_PASSWORD=              # Redis password (mặc định: rỗng)
```

## API Endpoints

### Authentication
- `POST /api/auth/login` - Đăng nhập
- `GET /api/auth/user` - Lấy thông tin user
- `GET /api/auth/users` - Lấy danh sách users

### Messages
- `POST /api/messages` - Gửi tin nhắn
- `GET /api/messages` - Lấy lịch sử tin nhắn
- `GET /api/messages/online` - Lấy danh sách users trực tuyến

### WebSocket
- `GET /ws?token=<token>` - Kết nối WebSocket

## Rate Limiting

Ứng dụng có các giới hạn tốc độ sau:

- **Message sending**: 10 tin nhắn/phút
- **WebSocket connections**: 60 kết nối/phút
- **General API**: 5 requests/phút

## WebSocket Message Types

- `message` - Tin nhắn chat
- `user_joined` - User tham gia
- `user_left` - User rời khỏi
- `typing` - Đang nhập
- `stop_typing` - Dừng nhập
- `ping/pong` - Heartbeat

## Tính năng nâng cao

### Typing Indicators
Ứng dụng hiển thị khi ai đó đang nhập tin nhắn.

### Online Users
Hiển thị danh sách người dùng đang trực tuyến.

### Message History
Lưu trữ lịch sử tin nhắn trong Redis với TTL 24 giờ.

### Auto-reconnection
WebSocket tự động kết nối lại khi mất kết nối.

## Phát triển

### Thêm tính năng mới

1. **Models**: Định nghĩa cấu trúc dữ liệu trong `models/`
2. **Controllers**: Xử lý logic trong `controllers/`
3. **Services**: Logic nghiệp vụ trong `services/`
4. **Middleware**: Middleware tùy chỉnh trong `middleware/`

### Testing

```bash
# Chạy tests
go test ./...

# Test với coverage
go test -cover ./...
```

## Production Deployment

### Docker
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/templates ./templates
CMD ["./main"]
```

### Environment Variables
```bash
export PORT=8080
export REDIS_ADDR=redis:6379
export REDIS_PASSWORD=your_redis_password
```

## Troubleshooting

### Redis Connection Issues
- Kiểm tra Redis có đang chạy không
- Kiểm tra cấu hình `REDIS_ADDR` và `REDIS_PASSWORD`
- Ứng dụng vẫn hoạt động nếu không có Redis (không lưu lịch sử)

### WebSocket Issues
- Kiểm tra firewall có chặn WebSocket không
- Kiểm tra proxy có hỗ trợ WebSocket không

### Rate Limiting
- Nếu bị giới hạn, chờ trong thời gian window được cấu hình
- Có thể điều chỉnh rate limits trong `middleware/rate_limiter.go`

## License

MIT License
