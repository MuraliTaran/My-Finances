package main

import (
	"context"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var MongoClient *mongo.Client
var Validator = validator.New()

func mongo_initializer() *mongo.Client {
	m_client, err := mongo.Connect(context.Background(), options.Client().ApplyURI("mongodb://localhost:27017"))
	if err != nil {
		log.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	err = m_client.Ping(ctx, nil)
	if err != nil {
		log.Fatal("unable to ping", err)
	}
	log.Println("Mongo client created")
	return m_client
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Credentials", "true")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Header("Access-Control-Allow-Methods", "POST, HEAD, PATCH, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}

func main() {
	MongoClient = mongo_initializer()

	router := gin.Default()
	router.Use(CORSMiddleware())

	router.POST("/Expense", CreateExpense())
	router.GET("/Expense/:uid", GetExpenses())
	router.PUT("/Expense/:id", UpdateExpense())
	router.DELETE("/Expense/:id", DeleteExpense())

	router.Run("localhost:8080")
}
