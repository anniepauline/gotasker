package main

import (
	"fmt"
	"net/http"
	"time"

	"strconv"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-contrib/cors" // <-- Add this import
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var db *gorm.DB
var jwtKey = []byte("secret")

type User struct {
	gorm.Model
	Username string `gorm:"unique"`
	Password string
	Theme    string `gorm:"default:light"` // light or dark

}

type Task struct {
	gorm.Model
	Title    string
	UserID   uint
	Status   string    `gorm:"default:todo"` // todo, in_progress, done
	DueDate  time.Time `gorm:"default:null"`
	Priority string    `gorm:"default:medium"` // low, medium, high

}

type Claims struct {
	UserID uint
	jwt.StandardClaims
}
type TaskHistory struct {
	gorm.Model
	TaskID     uint
	ChangedBy  uint
	Field      string
	OldValue   string
	NewValue   string
	ChangeTime time.Time
}

type Tag struct {
	gorm.Model
	Name   string
	UserID uint
}

type TaskTag struct {
	TaskID uint
	TagID  uint
}
type TaskTimeLog struct {
	gorm.Model
	TaskID   uint
	UserID   uint
	Duration time.Duration
	Note     string
}

func main() {
	router := gin.Default()

	// CORS Middleware configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://example.com"},                      // specify allowed origins
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},            // specify allowed methods
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"}, // specify allowed headers
		AllowCredentials: true,
	}))

	dsn := "host=localhost user=postgres password=admin dbname=gotasker port=5432 sslmode=disable"
	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("Failed to connect to database")
	}
	db = database
	db.AutoMigrate(&User{}, &Task{})

	r := gin.Default()

	r.POST("/register", Register)
	r.POST("/login", Login)

	auth := r.Group("/")
	auth.Use(AuthMiddleware())
	{
		auth.POST("/tasks", CreateTask)
		auth.GET("/tasks", GetTasks)
		auth.PUT("/tasks/:id", UpdateTask)
		auth.DELETE("/tasks/:id", DeleteTask)
		auth.POST("/logout", Logout)
		auth.GET("/tasks/due-soon", GetDueSoonTasks)
		auth.GET("/tasks/stats", GetTaskStats)

	}

	r.Run(":8080")
}

func Register(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(input.Password), 10)
	input.Password = string(hash)
	db.Create(&input)
	c.JSON(http.StatusOK, gin.H{"message": "registered"})
}

func Login(c *gin.Context) {
	var input User
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	var user User
	db.Where("username = ?", input.Username).First(&user)
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid credentials"})
		return
	}

	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &Claims{
		UserID: user.ID,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString(jwtKey)

	c.JSON(http.StatusOK, gin.H{"token": tokenString})
}

func CreateTask(c *gin.Context) {
	userID := c.MustGet("userID").(uint)
	var input struct {
		Title   string    `json:"title"`
		Status  string    `json:"status"`
		DueDate time.Time `json:"due_date"` // "2006-01-02T15:04:05Z"
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task := Task{
		Title:   input.Title,
		UserID:  userID,
		Status:  input.Status,
		DueDate: input.DueDate,
	}
	db.Create(&task)
	c.JSON(http.StatusOK, task)
}

func GetTasks(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// Pagination
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")
	page, _ := strconv.Atoi(pageStr)
	limit, _ := strconv.Atoi(limitStr)
	offset := (page - 1) * limit

	// Query filters
	search := c.Query("search")
	sort := c.DefaultQuery("sort", "desc") // asc or desc
	status := c.Query("status")
	due := c.Query("due") // yyyy-mm-dd

	query := db.Model(&Task{}).Where("user_id = ?", userID)

	if search != "" {
		query = query.Where("title ILIKE ?", "%"+search+"%")
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if due != "" {
		dueDate, err := time.Parse("2006-01-02", due)
		if err == nil {
			query = query.Where("due_date = ?", dueDate)
		}
	}

	var total int64
	query.Count(&total)

	var tasks []Task
	query.Order("created_at " + sort).Offset(offset).Limit(limit).Find(&tasks)

	c.JSON(http.StatusOK, gin.H{
		"tasks":       tasks,
		"total":       total,
		"page":        page,
		"limit":       limit,
		"total_pages": (total + int64(limit) - 1) / int64(limit),
	})
}

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")
		if tokenString == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		// Print the token for debugging
		fmt.Println("Token from header:", tokenString)

		// Remove the 'Bearer ' prefix if it exists
		if len(tokenString) > 7 && tokenString[:7] == "Bearer " {
			tokenString = tokenString[7:]
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		// Print the token details for debugging
		if err != nil {
			fmt.Println("Error parsing token:", err)
		}

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("userID", claims.UserID)
		c.Next()
	}
}
func UpdateTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("userID").(uint)

	var task Task
	if err := db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}

	var input struct {
		Title   string    `json:"title"`
		Status  string    `json:"status"`
		DueDate time.Time `json:"due_date"`
	}
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.Title = input.Title
	task.Status = input.Status
	task.DueDate = input.DueDate
	db.Save(&task)

	c.JSON(http.StatusOK, task)
}

func DeleteTask(c *gin.Context) {
	taskID := c.Param("id")
	userID := c.MustGet("userID").(uint)
	var task Task
	if err := db.Where("id = ? AND user_id = ?", taskID, userID).First(&task).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "task not found"})
		return
	}
	db.Model(&task).Update("deleted_at", time.Now())

	c.JSON(http.StatusOK, gin.H{"message": "task deleted"})
}

func Logout(c *gin.Context) {
	// Invalidate token logic, maybe by adding it to a blacklist (JWT token blacklist can be stored in Redis or database)
	c.JSON(http.StatusOK, gin.H{"message": "logged out"})
}

func GetTasksWithPagination(c *gin.Context) {
	userID := c.MustGet("userID").(uint)

	// Parse query params
	pageStr := c.DefaultQuery("page", "1")
	limitStr := c.DefaultQuery("limit", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 {
		limit = 10
	}

	offset := (page - 1) * limit

	var tasks []Task
	db.Where("user_id = ?", userID).Offset(offset).Limit(limit).Find(&tasks)

	c.JSON(http.StatusOK, tasks)
}

func GetTaskStats(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	var total, completed, pending int64

	db.Model(&Task{}).Where("user_id = ?", userID).Count(&total)
	db.Model(&Task{}).Where("user_id = ? AND status = ?", userID, "completed").Count(&completed)
	db.Model(&Task{}).Where("user_id = ? AND status = ?", userID, "pending").Count(&pending)

	c.JSON(http.StatusOK, gin.H{
		"total_tasks": total,
		"completed":   completed,
		"pending":     pending,
	})
}

func GetDueSoonTasks(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	now := time.Now()
	threeDaysLater := now.Add(72 * time.Hour)

	var tasks []Task
	err := db.Where("user_id = ? AND due_date BETWEEN ? AND ?", userID, now, threeDaysLater).
		Order("due_date ASC").
		Find(&tasks).Error

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch tasks"})
		return
	}

	c.JSON(http.StatusOK, tasks)
}
