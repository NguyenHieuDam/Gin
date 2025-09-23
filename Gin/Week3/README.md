# Chat App - Ứng dụng Trò chuyện Thời gian thực

Ứng dụng chat thời gian thực được xây dựng bằng Go, sử dụng WebSocket, PostgreSQL và Redis.

## 🚀 Tính năng

- ✅ **Trò chuyện thời gian thực** - Gửi và nhận tin nhắn ngay lập tức
- ✅ **WebSocket** - Giao tiếp hai chiều nhanh chóng
- ✅ **Theo dõi trạng thái hiện diện** - Hiển thị người dùng trực tuyến
- ✅ **Lưu trữ lịch sử tin nhắn** - Với Redis cho truy cập nhanh
- ✅ **Giới hạn tốc độ (Rate Limiting)** - Ngăn chặn spam
- ✅ **Giao diện web hiện đại** - Responsive và thân thiện với người dùng
- ✅ **Quản lý phòng trò chuyện** - Hỗ trợ nhiều phòng
- ✅ **Tìm kiếm tin nhắn** - Tìm kiếm trong lịch sử
- ✅ **Typing indicators** - Hiển thị ai đang gõ

## 🛠️ Công nghệ sử dụng

### Backend
- **Go 1.21+** - Ngôn ngữ lập trình chính
- **Gin** - Web framework
- **GORM** - ORM cho database
- **PostgreSQL** - Database chính
- **Redis** - Cache và real-time data
- **Gorilla WebSocket** - WebSocket implementation

### Frontend
- **HTML5/CSS3** - Giao diện người dùng
- **Vanilla JavaScript** - Logic frontend
- **Font Awesome** - Icons
- **Responsive Design** - Tương thích mobile

## 📋 Yêu cầu hệ thống

- Go 1.21 hoặc cao hơn
- PostgreSQL 12+
- Redis 6+
- Modern web browser

## 🔧 Cài đặt và Chạy

### 1. Clone repository
```bash
git clone <repository-url>
cd chat-app
```

### 2. Cài đặt dependencies
```bash
go mod tidy
```

### 3. Cài đặt và cấu hình Database

#### PostgreSQL
```sql
-- Tạo database
CREATE DATABASE chatapp;

-- Tạo user (optional)
CREATE USER chatapp_user WITH PASSWORD 'your_password';
GRANT ALL PRIVILEGES ON DATABASE chatapp TO chatapp_user;
```

#### Redis
```bash
# Cài đặt Redis (Ubuntu/Debian)
sudo apt update
sudo apt install redis-server

# Khởi động Redis
sudo systemctl start redis-server
sudo systemctl enable redis-server
```

### 4. Cấu hình Environment Variables

Tạo file `.env` hoặc set environment variables:

```bash
# Database
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=123456
export DB_NAME=chatapp
export DB_SSLMODE=disable

# Redis
export REDIS_HOST=localhost
export REDIS_PORT=6379
export REDIS_PASSWORD=
export REDIS_DB=0

# Server
export PORT=8080
export GIN_MODE=debug
```

### 5. Chạy ứng dụng
```bash
go run main.go
```

Ứng dụng sẽ chạy tại: `http://localhost:8080`

## 📖 API Documentation

### Authentication Endpoints

#### Đăng ký
```http
POST /api/v1/auth/register
Content-Type: application/json

{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "password123"
}
```

#### Đăng nhập
```http
POST /api/v1/auth/login
Content-Type: application/json

{
  "username": "john_doe",
  "password": "password123"
}
```

### Message Endpoints

#### Lấy tin nhắn gần đây
```http
GET /api/v1/messages/{roomId}/recent?limit=20
```

#### Tìm kiếm tin nhắn
```http
GET /api/v1/messages/{roomId}/search?q=keyword&limit=20
```

### WebSocket Endpoints

#### Kết nối WebSocket
```http
GET /api/v1/ws/?user_id=1&username=john_doe&room_id=general
```

#### Lấy danh sách người trực tuyến
```http
GET /api/v1/ws/general/users
```

### Health Check
```http
GET /health
```

## 🔌 WebSocket Message Format

### Gửi tin nhắn
```json
{
  "type": "message",
  "data": {
    "content": "Xin chào mọi người!",
    "room_id": "general"
  }
}
```

### Typing indicator
```json
{
  "type": "typing"
}
```

### Ping/Pong
```json
{
  "type": "ping"
}
```

## 📁 Cấu trúc Project

```
chat-app/
├── main.go                 # Entry point
├── config.go              # Configuration
├── go.mod                 # Go modules
├── go.sum                 # Go modules checksum
├── README.md              # Documentation
├── db/                    # Database connections
│   ├── postgres.go        # PostgreSQL connection
│   └── redis.go           # Redis connection
├── handlers/              # HTTP handlers
│   ├── auth_handler.go    # Authentication handlers
│   ├── message_handler.go # Message handlers
│   ├── ws_handler.go      # WebSocket handlers
│   └── handler.go         # Main handler
├── middleware/            # Middleware
│   └── rate_limit.go      # Rate limiting
├── models/                # Data models
│   ├── user.go           # User model
│   └── message.go        # Message model
├── services/              # Business logic
│   ├── user_service.go    # User service
│   ├── message_service.go # Message service
│   └── presence_service.go # Presence service
├── websocket/             # WebSocket implementation
│   ├── client.go          # WebSocket client
│   └── hub.go             # WebSocket hub
└── static/                # Frontend files
    ├── index.html         # Main HTML
    ├── style.css          # CSS styles
    └── app.js             # JavaScript
```

## 🚀 Deployment

### Docker (Recommended)

Tạo `Dockerfile`:
```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/static ./static
CMD ["./main"]
```

Tạo `docker-compose.yml`:
```yaml
version: '3.8'
services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: chatapp
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: 123456
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

volumes:
  postgres_data:
  redis_data:
```

Chạy với Docker:
```bash
docker-compose up -d
```

## 🔒 Security Features

- **Rate Limiting** - Ngăn chặn spam và DDoS
- **Input Validation** - Kiểm tra dữ liệu đầu vào
- **Password Hashing** - Mã hóa mật khẩu với bcrypt
- **CORS Protection** - Cấu hình CORS phù hợp
- **Connection Limits** - Giới hạn kết nối WebSocket

## 🧪 Testing

```bash
# Chạy tests
go test ./...

# Test coverage
go test -cover ./...

# Benchmark tests
go test -bench=. ./...
```

## 📈 Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Metrics Endpoints
- `/health` - Health status
- `/ping` - Simple ping endpoint

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

## 📝 License

This project is licensed under the MIT License.

## 🆘 Troubleshooting

### Common Issues

1. **Database connection failed**
   - Kiểm tra PostgreSQL đang chạy
   - Verify database credentials
   - Check network connectivity

2. **Redis connection failed**
   - Kiểm tra Redis đang chạy
   - Verify Redis configuration
   - Check firewall settings

3. **WebSocket connection failed**
   - Kiểm tra port 8080 available
   - Verify CORS settings
   - Check browser console for errors

### Logs

```bash
# View application logs
tail -f /var/log/chat-app.log

# Docker logs
docker-compose logs -f app
```

## 📞 Support

Nếu gặp vấn đề, vui lòng:
1. Check logs
2. Verify configuration
3. Create issue trên GitHub
4. Contact development team

---

**Chúc bạn sử dụng ứng dụng vui vẻ! 🎉**
