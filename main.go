package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var (
	rdb *redis.Client
)

func init() {
	rdb = redis.NewClient(&redis.Options{
		Addr:     "redis-19897.c281.us-east-1-2.ec2.cloud.redislabs.com:19897",
		Username: "default",
		Password: "Z58DxJjDvlEpxCtBbmGbtpLT0rOwNSpe",
		DB:       0,
	})
}

type User struct {
	Username string `json:"username"`
	Wins     int    `json:"wins"`
}

// RegisterUser registers a new user
func RegisterUser(c *gin.Context) {
	var user User
	if err := c.BindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add or update the user data in Redis
	err := rdb.HSet(c, "users", user.Username, user.Wins).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered/updated successfully", "data": user})
}

func GetUser(c *gin.Context) {
	username := c.Param("username")

	// Retrieve the user details from Redis
	wins, err := rdb.HGet(c, "users", username).Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Check if the user exists
	if wins == "" {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	// Parse wins to integer
	var winsInt int
	fmt.Sscanf(wins, "%d", &winsInt)

	// Create and return the user object
	user := User{
		Username: username,
		Wins:     winsInt,
	}

	c.JSON(http.StatusOK, user)
}

// RecordWin records a win for a user
func RecordWin(c *gin.Context) {
	username := c.Param("username")

	// Increment the number of wins for the user
	err := rdb.HIncrBy(c, "users", username, 1).Err()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "win recorded successfully"})
}

// GetLeaderboard retrieves the leaderboard
func GetLeaderboard(c *gin.Context) {
	// Retrieve all users and their wins
	users, err := rdb.HGetAll(c, "users").Result()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var leaderboard []User
	for username, winsStr := range users {
		wins := 0
		fmt.Sscanf(winsStr, "%d", &wins)
		leaderboard = append(leaderboard, User{Username: username, Wins: wins})
	}

	c.JSON(http.StatusOK, leaderboard)
}

func main() {
	// Initialize Gin router
	router := gin.Default()

	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"https://cardgame-fe.vercel.app"} // Add your frontend URL here
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept"}
	router.Use(cors.New(config))

	// Routes
	router.POST("/register", RegisterUser)
	router.POST("/record-win/:username", RecordWin)
	router.GET("/leaderboard", GetLeaderboard)
	router.GET("/user/:username", GetUser)

	// Start the server
	if err := router.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
