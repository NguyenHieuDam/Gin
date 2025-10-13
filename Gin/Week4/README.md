# News Aggregator System

Hệ thống tổng hợp tin tức đơn giản với Go, sử dụng microservices architecture, Kafka, và Redis.

## 🏗️ Kiến trúc hệ thống

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│  News Collector │    │   News API      │    │   Kafka UI      │
│   (Port 8081)   │    │   (Port 8080)   │    │   (Port 8082)   │
└─────────┬───────┘    └─────────┬───────┘    └─────────────────┘
          │                      │
          │                      │
          ▼                      ▼
    ┌─────────────┐        ┌─────────────┐
    │    Kafka    │        │    Redis    │
    │ (Port 9092) │        │ (Port 6379) │
    └─────────────┘        └─────────────┘
```

## 🚀 Tính năng

- **News Collector Service**: Thu thập tin tức từ các nguồn khác nhau
- **News API Service**: Cung cấp API REST với rate limiting
- **Web Interface**: Giao diện web đẹp mắt với HTML/CSS/JavaScript
- **Kafka**: Xử lý tin tức bất đồng bộ
- **Redis**: Lưu trữ và cache tin tức
- **Rate Limiting**: Giới hạn tốc độ API (100 requests/phút)
- **Docker Compose**: Triển khai toàn bộ hệ thống
- **Production Ready**: Cấu hình sẵn sàng cho production deployment

## 📋 Yêu cầu hệ thống

- Docker & Docker Compose
- Go 1.21+ (nếu chạy local)
- 4GB RAM tối thiểu

## 🛠️ Cài đặt và chạy

### 1. Clone repository
```bash
git clone <repository-url>
cd news-aggregator
```

### 2. Chạy với Docker Compose
```bash
# Khởi động toàn bộ hệ thống
docker-compose up -d

# Xem logs
docker-compose logs -f

# Dừng hệ thống
docker-compose down
```

### 3. Truy cập giao diện web
Mở trình duyệt và truy cập: **http://localhost:8080**

Giao diện web bao gồm:
- 📰 Hiển thị tin tức với layout đẹp mắt
- 🔍 Tìm kiếm tin tức theo từ khóa
- 📂 Lọc tin tức theo danh mục
- 📊 Thống kê real-time
- 🔄 Thu thập tin tức thủ công
- 📱 Responsive design cho mobile

### 4. Chạy local (tùy chọn)

#### Cài đặt dependencies
```bash
go mod download
```

#### Khởi động Redis và Kafka
```bash
# Sử dụng Docker Compose chỉ cho infrastructure
docker-compose up -d redis kafka zookeeper
```

#### Chạy services
```bash
# Terminal 1: News Collector
go run cmd/collector/main.go

# Terminal 2: News API
go run cmd/api/main.go
```

## 🌐 Web Interface

### Giao diện chính
- **URL**: http://localhost:8080
- **Tính năng**:
  - Hiển thị tin tức dạng grid với card đẹp mắt
  - Tìm kiếm real-time
  - Lọc theo danh mục (Công nghệ, Kinh doanh, Thể thao)
  - Phân trang
  - Modal xem chi tiết bài viết
  - Thu thập tin tức thủ công
  - Thống kê real-time

### Responsive Design
- Tối ưu cho desktop, tablet và mobile
- Dark/Light theme
- Animations mượt mà
- Loading states

## 📡 API Endpoints

### News API Service (Port 8080)

#### 1. Health Check
```bash
GET http://localhost:8080/health
```

#### 2. Lấy tin tức mới nhất
```bash
GET http://localhost:8080/api/v1/news?limit=10&offset=0
```

**Response:**
```json
{
  "articles": [
    {
      "id": "news_1234567890",
      "title": "Breaking: Technology News Update",
      "content": "This is a sample technology news article content...",
      "source": "TechNews",
      "url": "https://example.com/tech-news",
      "published_at": "2024-01-01T12:00:00Z",
      "created_at": "2024-01-01T12:00:00Z",
      "category": "technology"
    }
  ],
  "total": 1,
  "page": 1,
  "limit": 10
}
```

#### 3. Lấy tin tức theo ID
```bash
GET http://localhost:8080/api/v1/news{id}/
```

#### 4. Lấy tin tức theo category
```bash
GET http://localhost:8080/api/v1/news/category/technology?limit=10&offset=0
```

### News Collector Service (Port 8081)

#### 1. Health Check
```bash
GET http://localhost:8081/health
```

#### 2. Thu thập tin tức thủ công
```bash
POST http://localhost:8081/collect
```

## 🔧 Cấu hình

Chỉnh sửa file `config.env`:

```env
# Database
REDIS_URL=localhost:6379

# Kafka
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=news-updates

# API Configuration
API_PORT=8080
COLLECTOR_PORT=8081

# Rate Limiting
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# News Sources
NEWS_API_KEY=your_news_api_key_here
```

## 📊 Monitoring

### Kafka UI
Truy cập http://localhost:8082 để xem Kafka topics và messages.

### Health Checks
- News API: http://localhost:8080/health
- News Collector: http://localhost:8081/health

## 🧪 Testing

### Test API với curl
```bash
# Lấy tin tức mới nhất
curl "http://localhost:8080/api/v1/news?limit=5"

# Thu thập tin tức thủ công
curl -X POST "http://localhost:8081/collect"

# Test rate limiting
for i in {1..105}; do curl "http://localhost:8080/api/v1/news"; done
```

### Test với Postman
Import collection từ file `postman_collection.json` (nếu có).

## 🔒 Rate Limiting

- **Giới hạn**: 100 requests/phút per IP
- **Headers trả về**:
  - `X-RateLimit-Limit`: Giới hạn tối đa
  - `X-RateLimit-Remaining`: Số requests còn lại
  - `X-RateLimit-Reset`: Thời gian reset

## 🏃‍♂️ Scaling

### Horizontal Scaling
```bash
# Scale News API service
docker-compose up -d --scale news-api=3

# Scale News Collector service
docker-compose up -d --scale news-collector=2
```

### Load Balancing
Thêm Nginx hoặc HAProxy để load balance giữa các instances.

## 🐛 Troubleshooting

### 1. Services không start
```bash
# Check logs
docker-compose logs news-api
docker-compose logs news-collector

# Check health
curl http://localhost:8080/health
curl http://localhost:8081/health
```

### 2. Kafka connection issues
```bash
# Check Kafka status
docker-compose logs kafka

# Check topic
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

### 3. Redis connection issues
```bash
# Check Redis status
docker-compose logs redis

# Test Redis connection
docker exec redis redis-cli ping
```

## 📁 Cấu trúc project

```
news-aggregator/
├── cmd/
│   ├── api/           # News API service
│   └── collector/     # News Collector service
├── internal/
│   ├── config/        # Configuration
│   ├── kafka/         # Kafka producer/consumer
│   ├── middleware/    # Rate limiting middleware
│   ├── models/        # Data models
│   └── redis/         # Redis client
├── docker-compose.yml
├── Dockerfile.api
├── Dockerfile.collector
├── go.mod
├── go.sum
└── README.md
```

## 🚀 Deploy lên Production

### Deploy nhanh
```bash
# Sử dụng script deploy tự động
./deploy.sh production

# Hoặc deploy thủ công
docker-compose -f docker-compose.production.yml up -d
```

### Deploy lên Cloud
Xem hướng dẫn chi tiết trong file [DEPLOYMENT.md](DEPLOYMENT.md) để deploy lên:
- DigitalOcean
- AWS EC2
- Google Cloud Platform
- VPS/Cloud Server

### Tính năng Production
- Nginx reverse proxy với SSL
- Rate limiting nâng cao
- Health checks
- Auto-restart
- Logging và monitoring
- Security headers

## 🚀 Mở rộng

### Thêm nguồn tin tức mới
1. Cập nhật `collectFromExternalSources()` trong `cmd/collector/main.go`
2. Thêm logic parse cho nguồn mới
3. Transform data về format `NewsArticle`

### Thêm authentication
1. Thêm JWT middleware
2. Implement user management
3. Add API keys cho external access

### Thêm database
1. Thay thế Redis bằng PostgreSQL/MongoDB
2. Implement data persistence
3. Add data migration scripts

### Cải thiện giao diện
1. Thêm PWA (Progressive Web App)
2. Implement offline support
3. Thêm dark/light theme toggle
4. Add more animations và transitions

## 📝 License

MIT License

## 🤝 Contributing

1. Fork the repository
2. Create feature branch
3. Commit changes
4. Push to branch
5. Create Pull Request
