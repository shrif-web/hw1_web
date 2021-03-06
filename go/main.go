package main

import (
	"github.com/go-redis/redis/v8"
	"github.com/gin-gonic/gin"
	"crypto/sha256"
	"net/http"
	"context"
	"fmt"
	"os"
)

var ctx = context.Background()

var rdb = redis.NewClient(&redis.Options{
	Addr:	  "localhost:6379",
	Password: "", // no password set
	DB:		  0,  // use default DB
})

type Input struct {
	Text string `json:"String"`
}


func getSHA256(input string) string {
	hash := sha256.Sum256([]byte(input))
	out := fmt.Sprintf("%x", hash)
	return out
}

func sha256_get(c *gin.Context){
	q := c.Request.URL.Query().Get("hash_text");
	if len(q) <= 0 { c.HTML(http.StatusOK, "sha256.html", gin.H{}); return }

	value, err := rdb.Get(ctx, q).Result()
	if err == redis.Nil {
		html := "<p>Hash value not found, <a href=/go/sha256>back</a></p>"
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(html))
		return
	} else if err != nil {
		html := "<p>Something bad happend, <a href=/go/sha256>back</a></p>"
		c.Writer.WriteHeader(http.StatusOK)
		c.Writer.Write([]byte(html))
		panic(err)
		return
	}
	c.HTML(http.StatusOK, "sha256.html", gin.H{"hash_input": q, "unhashed_value": value})

}

func sha256_post(c *gin.Context){
	input := c.PostForm("text")
	if len(input) < 8 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Your input must be more than 8 chars"})
		return
	}
	hash := getSHA256(input)

	err := rdb.Set(ctx, hash, input, 0).Err()
	if err != nil {
		panic(err)
	}

	c.HTML(http.StatusOK, "sha256.html", gin.H{"input": input, "sha_value": hash})
}

func sha_post(c *gin.Context){
	var input Input
	c.BindJSON(&input)
	if len(input.Text) < 8 { c.JSON(http.StatusBadRequest, gin.H{"error": "Your input must be more than 8 chars"}); return }
	hash := getSHA256(input.Text)

	err := rdb.Set(ctx, hash, input.Text, 0).Err()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Saving in Redis failed"})
		panic(err)
	}

	c.JSON(http.StatusOK, gin.H{"sha256": hash})
}

func sha_get(c *gin.Context){
	q := c.Request.URL.Query().Get("hash_value");

	value, err := rdb.Get(ctx, q).Result();
        if err == redis.Nil {
		c.JSON(http.StatusOK, gin.H{"error": "Hash value not found"})
		return
        } else if err != nil {
		c.JSON(http.StatusOK, gin.H{"error": "Somthing bad happend"})
		panic(err)
                return
        }
	c.JSON(http.StatusOK, gin.H{"value": value})
}

func main() {
	PORT := os.Getenv("PORT")
	HOST_LAN_IP := os.Getenv("LAN_HOST_IP")
	LISTEN_ADDRESS := fmt.Sprintf("%s:%s", HOST_LAN_IP, PORT)

	r := gin.Default()
	r.LoadHTMLGlob("./templates/*")

	r.GET("/ping", func(c *gin.Context) { // For the sake of testing
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.GET("/sha256", sha256_get)
	r.POST("/sha256", sha256_post)
	r.POST("/sha", sha_post)
	r.GET("/sha", sha_get)

	r.Run(LISTEN_ADDRESS) // Listens on 0.0.0.0:8080
}
