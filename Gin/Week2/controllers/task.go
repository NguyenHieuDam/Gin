package controllers

import (
	"net/http"
	"week2/config"
	"week2/middlewares"
	"week2/models"

	"github.com/gin-gonic/gin"
)

// Lấy danh sách task
type TaskResponse struct {
	ID          uint   `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
}

func GetTasks(c *gin.Context) {
	var tasks []models.Task
	config.DB.Find(&tasks)

	var response []TaskResponse
	for _, task := range tasks {
		response = append(response, TaskResponse{
			ID:          task.ID,
			Title:       task.Title,
			Description: task.Description,
			Status:      task.Status,
		})
	}

	c.JSON(http.StatusOK, response)
}

// Tạo task mới
func CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if task.Status != "pending" && task.Status != "in-progress" && task.Status != "completed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Trạng thái không hợp lệ"})
		return
	}
	config.DB.Create(&task)
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
	}
	c.JSON(http.StatusCreated, response)
}

// Cập nhật task
func UpdateTask(c *gin.Context) {
	var task models.Task
	if err := config.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task không tồn tại"})
		return
	}

	var input models.Task
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	config.DB.Model(&task).Updates(input)
	response := TaskResponse{
		ID:          task.ID,
		Title:       task.Title,
		Description: task.Description,
		Status:      task.Status,
	}
	c.JSON(http.StatusOK, response)
}

// Xóa task
func DeleteTask(c *gin.Context) {
	//check admin mới được xóa task
	if !middlewares.CheckAdmin(c) {
		return
	}

	var task models.Task
	if err := config.DB.First(&task, c.Param("id")).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task không tồn tại"})
		return
	}

	config.DB.Delete(&task)
	c.JSON(http.StatusOK, gin.H{"message": "Xóa thành công"})
}
