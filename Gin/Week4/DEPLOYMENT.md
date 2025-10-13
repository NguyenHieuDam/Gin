# Hướng dẫn Deploy News Aggregator lên Cloud

## 🚀 Tổng quan

Hệ thống News Aggregator có thể được deploy lên nhiều nền tảng cloud khác nhau. Tài liệu này hướng dẫn deploy lên các platform phổ biến.

## 📋 Yêu cầu hệ thống

- **RAM**: Tối thiểu 2GB, khuyến nghị 4GB+
- **CPU**: 2 cores trở lên
- **Storage**: 10GB trống
- **Network**: Port 80, 443 (HTTP/HTTPS)

## 🌐 Các phương án Deploy

### 1. Deploy lên VPS/Cloud Server

#### 1.1 Chuẩn bị server
```bash
# Cập nhật hệ thống
sudo apt update && sudo apt upgrade -y

# Cài đặt Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sudo sh get-docker.sh
sudo usermod -aG docker $USER

# Cài đặt Docker Compose
sudo curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose

# Logout và login lại để áp dụng group docker
```

#### 1.2 Deploy ứng dụng
```bash
# Clone repository
git clone <your-repo-url>
cd news-aggregator

# Deploy production
docker-compose -f docker-compose.production.yml up -d

# Kiểm tra logs
docker-compose -f docker-compose.production.yml logs -f
```

#### 1.3 Cấu hình domain (tùy chọn)
```bash
# Cài đặt Nginx
sudo apt install nginx -y

# Tạo file cấu hình
sudo nano /etc/nginx/sites-available/news-aggregator

# Nội dung file:
server {
    listen 80;
    server_name your-domain.com;
    
    location / {
        proxy_pass http://localhost:80;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

# Kích hoạt site
sudo ln -s /etc/nginx/sites-available/news-aggregator /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx
```

### 2. Deploy lên DigitalOcean

#### 2.1 Tạo Droplet
1. Đăng nhập DigitalOcean
2. Tạo Droplet mới:
   - **Image**: Ubuntu 22.04 LTS
   - **Size**: Basic $12/month (2GB RAM, 1 CPU)
   - **Region**: Gần nhất với người dùng
   - **Authentication**: SSH Key

#### 2.2 Cấu hình Droplet
```bash
# SSH vào droplet
ssh root@your-droplet-ip

# Cài đặt Docker và Docker Compose (như hướng dẫn trên)
# Clone và deploy ứng dụng
```

#### 2.3 Cấu hình Firewall
```bash
# Mở port cần thiết
ufw allow 22    # SSH
ufw allow 80    # HTTP
ufw allow 443   # HTTPS
ufw enable
```

### 3. Deploy lên AWS EC2

#### 3.1 Tạo EC2 Instance
1. Đăng nhập AWS Console
2. Tạo EC2 Instance:
   - **AMI**: Ubuntu Server 22.04 LTS
   - **Instance Type**: t3.small (2 vCPU, 2GB RAM)
   - **Security Group**: Mở port 22, 80, 443
   - **Key Pair**: Tạo hoặc chọn key pair

#### 3.2 Cấu hình EC2
```bash
# SSH vào instance
ssh -i your-key.pem ubuntu@your-ec2-ip

# Cài đặt Docker và deploy (như hướng dẫn trên)
```

#### 3.3 Cấu hình Security Group
- **Inbound Rules**:
  - SSH (22): Your IP
  - HTTP (80): 0.0.0.0/0
  - HTTPS (443): 0.0.0.0/0

### 4. Deploy lên Google Cloud Platform

#### 4.1 Tạo VM Instance
```bash
# Sử dụng gcloud CLI
gcloud compute instances create news-aggregator \
    --image-family=ubuntu-2204-lts \
    --image-project=ubuntu-os-cloud \
    --machine-type=e2-small \
    --zone=asia-southeast1-a \
    --tags=http-server,https-server
```

#### 4.2 Cấu hình Firewall
```bash
# Mở port HTTP/HTTPS
gcloud compute firewall-rules create allow-http-https \
    --allow tcp:80,tcp:443 \
    --source-ranges 0.0.0.0/0 \
    --target-tags http-server,https-server
```

### 5. Deploy lên Heroku (Limited)

**Lưu ý**: Heroku có giới hạn về persistent storage, không phù hợp cho Kafka và Redis.

#### 5.1 Chuẩn bị
```bash
# Cài đặt Heroku CLI
# Tạo Procfile
echo "web: ./api" > Procfile
```

#### 5.2 Deploy
```bash
# Login Heroku
heroku login

# Tạo app
heroku create your-news-aggregator

# Deploy
git push heroku main
```

### 6. Deploy lên Railway

#### 6.1 Chuẩn bị
1. Đăng nhập Railway
2. Connect GitHub repository
3. Tạo project mới

#### 6.2 Cấu hình
```yaml
# railway.json
{
  "build": {
    "builder": "DOCKERFILE",
    "dockerfilePath": "Dockerfile.production"
  },
  "deploy": {
    "startCommand": "./api",
    "healthcheckPath": "/health",
    "healthcheckTimeout": 100
  }
}
```

## 🔧 Cấu hình Production

### 1. Environment Variables
```bash
# Tạo file .env.production
REDIS_URL=redis:6379
KAFKA_BROKER=kafka:29092
KAFKA_TOPIC=news-updates
API_PORT=8080
COLLECTOR_PORT=8081
RATE_LIMIT_REQUESTS=1000
RATE_LIMIT_WINDOW=60
NEWS_API_KEY=your_actual_api_key
```

### 2. SSL/HTTPS Setup
```bash
# Sử dụng Let's Encrypt
sudo apt install certbot python3-certbot-nginx -y

# Lấy SSL certificate
sudo certbot --nginx -d your-domain.com

# Auto-renewal
sudo crontab -e
# Thêm dòng: 0 12 * * * /usr/bin/certbot renew --quiet
```

### 3. Monitoring và Logging
```bash
# Cài đặt monitoring tools
docker run -d \
  --name=prometheus \
  -p 9090:9090 \
  prom/prometheus

# Log rotation
sudo nano /etc/logrotate.d/docker
# Nội dung:
/var/lib/docker/containers/*/*.log {
  rotate 7
  daily
  compress
  size=1M
  missingok
  delaycompress
  copytruncate
}
```

## 📊 Performance Tuning

### 1. Redis Optimization
```bash
# Trong redis.conf
maxmemory 256mb
maxmemory-policy allkeys-lru
save 900 1
save 300 10
save 60 10000
```

### 2. Kafka Optimization
```bash
# Trong kafka server.properties
num.network.threads=8
num.io.threads=8
socket.send.buffer.bytes=102400
socket.receive.buffer.bytes=102400
socket.request.max.bytes=104857600
```

### 3. Go Application
```bash
# Set GOMAXPROCS
export GOMAXPROCS=2

# Memory optimization
export GOGC=100
```

## 🔒 Security Best Practices

### 1. Firewall Configuration
```bash
# UFW configuration
ufw default deny incoming
ufw default allow outgoing
ufw allow ssh
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable
```

### 2. Docker Security
```bash
# Run containers as non-root
# Use specific image tags
# Regular security updates
docker system prune -a
```

### 3. API Security
- Rate limiting đã được cấu hình
- CORS headers
- Input validation
- HTTPS only trong production

## 📈 Scaling

### 1. Horizontal Scaling
```bash
# Scale API service
docker-compose -f docker-compose.production.yml up -d --scale news-api=3

# Load balancer configuration
upstream news_api {
    server news-api-1:8080;
    server news-api-2:8080;
    server news-api-3:8080;
}
```

### 2. Database Scaling
- Redis Cluster cho high availability
- Kafka partitioning cho throughput cao
- External database (PostgreSQL) cho persistence

## 🚨 Troubleshooting

### 1. Common Issues
```bash
# Check container status
docker-compose ps

# View logs
docker-compose logs news-api
docker-compose logs news-collector

# Check resource usage
docker stats

# Restart services
docker-compose restart news-api
```

### 2. Performance Issues
```bash
# Check memory usage
free -h
docker system df

# Check disk space
df -h

# Monitor network
netstat -tulpn
```

### 3. Health Checks
```bash
# API health
curl http://localhost/health

# Collector health
curl http://localhost:8081/health

# Redis connection
docker exec redis redis-cli ping

# Kafka topics
docker exec kafka kafka-topics --bootstrap-server localhost:9092 --list
```

## 📝 Maintenance

### 1. Regular Updates
```bash
# Update system
sudo apt update && sudo apt upgrade -y

# Update Docker images
docker-compose pull
docker-compose up -d

# Clean up
docker system prune -a
```

### 2. Backup Strategy
```bash
# Backup Redis data
docker exec redis redis-cli BGSAVE

# Backup configuration
tar -czf backup-$(date +%Y%m%d).tar.gz docker-compose.production.yml config.env
```

### 3. Monitoring
- Set up alerts cho CPU, Memory, Disk usage
- Monitor API response times
- Track error rates
- Log analysis

## 💰 Cost Estimation

### VPS/Cloud Server
- **DigitalOcean**: $12-24/month
- **AWS EC2**: $15-30/month
- **Google Cloud**: $10-25/month
- **Vultr**: $6-12/month

### Domain & SSL
- **Domain**: $10-15/year
- **SSL**: Free (Let's Encrypt)

### Total: ~$15-35/month cho một instance cơ bản

## 🎯 Kết luận

Hệ thống News Aggregator có thể được deploy dễ dàng lên nhiều nền tảng cloud khác nhau. Chọn phương án phù hợp với ngân sách và yêu cầu kỹ thuật của bạn.

Để bắt đầu, khuyến nghị sử dụng DigitalOcean hoặc VPS với Docker Compose cho đơn giản và hiệu quả.
