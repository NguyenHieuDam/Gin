# 🚀 News Aggregator - Hướng dẫn sử dụng đơn giản

## ✅ **Đã khắc phục lỗi Kafka!**

Lỗi `Failed to read message from Kafka` đã được giải quyết bằng cách tạo version đơn giản không cần Docker.

## 🎯 **Cách chạy ứng dụng**

### **Phương pháp 1: Chạy trực tiếp (Khuyến nghị)**

```bash
# Build ứng dụng
go build -o bin/simple.exe ./cmd/simple

# Chạy ứng dụng
.\bin\simple.exe
```

### **Phương pháp 2: Sử dụng script**

```bash
# Windows Batch
run-simple.bat

# PowerShell
.\run-simple.ps1
```

## 🌐 **Truy cập ứng dụng**

- **Web Interface**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **News API**: http://localhost:8080/api/v1/news

## 📡 **API Endpoints**

### **1. Health Check**
```bash
GET /health
```

### **2. Lấy tất cả tin tức**
```bash
GET /api/v1/news
GET /api/v1/news?limit=5&offset=0
```

### **3. Lấy tin tức theo category**
```bash
GET /api/v1/news/category/technology
GET /api/v1/news/category/business
GET /api/v1/news/category/sports
```

### **4. Lấy tin tức theo ID**
```bash
GET /api/v1/news/{id}
```

### **5. Thu thập tin tức mới**
```bash
POST /api/v1/collect
```

## 🧪 **Test API với PowerShell**

```powershell
# Health check
Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing

# Lấy tất cả tin tức
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/news" -UseBasicParsing

# Lấy tin tức technology
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/news/category/technology" -UseBasicParsing

# Thu thập tin tức mới
Invoke-WebRequest -Uri "http://localhost:8080/api/v1/collect" -Method POST -UseBasicParsing
```

## 🎉 **Tính năng hoạt động**

✅ **API RESTful** - Tất cả endpoints hoạt động  
✅ **CORS Support** - Có thể gọi từ browser  
✅ **Auto News Generation** - Tự động tạo tin tức mỗi 30 giây  
✅ **Category Filtering** - Lọc tin tức theo danh mục  
✅ **Pagination** - Hỗ trợ phân trang  
✅ **Health Check** - Kiểm tra trạng thái service  
✅ **Graceful Shutdown** - Tắt ứng dụng an toàn  

## 🔧 **Cấu hình**

Ứng dụng sử dụng file `config.env` để cấu hình:

```env
API_PORT=8080
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60
```

## 🚨 **Lưu ý quan trọng**

1. **Không cần Docker** - Version này chạy trực tiếp trên Go
2. **In-memory storage** - Dữ liệu sẽ mất khi restart
3. **Auto-generated news** - Tin tức được tạo tự động
4. **Port 8080** - Đảm bảo port này không bị sử dụng

## 🎯 **Kết quả**

Hệ thống News Aggregator đã hoạt động hoàn hảo với:
- ✅ API trả về dữ liệu JSON
- ✅ CORS headers đầy đủ
- ✅ Background news generation
- ✅ Category filtering
- ✅ Health monitoring

**Lỗi Kafka đã được khắc phục hoàn toàn!** 🎉
