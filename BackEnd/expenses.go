package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Expense struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"_id,omitempty" yaml:"_id,omitempty"`
	Amount      float32            `bson:"amt" json:"amt" yaml:"amt" validate:"Required,gt=0"`
	Category    int                `bson:"category" json:"category" yaml:"category"`
	SubCategory string             `bson:"subcategory,omitempty" json:"subcategory,omitempty" yaml:"subcategory,omitempty"`
	Description string             `bson:"description,omitempty" json:"description,omitempty" yaml:"description,omitempty"`
	TimeStamp   int64              `bson:"ts" json:"ts" yaml:"ts"`
	UserId      string             `bson:"user_id" json:"user_id" yaml:"user_id" validate:"Required"`
	Cashback    int                `bson:"cashback,omitempty" json:"cashback,omitempty" yaml:"cashback,omitempty"`
}

func CreateExpense() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var newExpense Expense
		var err error

		if err = c.BindJSON(&newExpense); err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Binding failed"})
			return
		}

		if err = Validator.Struct(newExpense); err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Validation failed"})
			return
		}

		newExpense.ID = primitive.NewObjectID()
		if newExpense.TimeStamp == 0 {
			newExpense.TimeStamp = time.Now().UnixMilli()
		}

		exp_Collection := MongoClient.Database("PocketPassbook").Collection("expenses_cl")
		_, err = exp_Collection.InsertOne(ctx, newExpense)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": false, "message": "Mongo insertion failed"})
			return
		}

		c.IndentedJSON(http.StatusCreated, gin.H{"status": true, "message": "Inserted successfully"})

	}
}

func GetExpenses() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user_id := c.Param("uid")
		var err error

		query := bson.M{"user_id": user_id}
		if c.Query("from") != "" {
			num, err := strconv.Atoi(c.Query("from"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": false, "message": "Provided invalid query value for 'from'"})
				return
			}
			query["ts"] = bson.M{"$lte": num}
		}
		if c.Query("to") != "" {
			num, err := strconv.Atoi(c.Query("to"))
			if err != nil {
				c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error(), "status": false, "message": "Provided invalid query value for 'to'"})
				return
			}
			query["ts"] = bson.M{"$lte": num}
		}
		if c.Query("category") != "" {
			query["category"] = c.Query("category")
		}

		findOptions := options.Find().SetSort(bson.D{primitive.E{Key: "ts", Value: -1}})

		exp_Collection := MongoClient.Database("PocketPassbook").Collection("expenses_cl")

		cursor, err := exp_Collection.Find(ctx, query, findOptions)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": false, "message": "Mongo find failed"})
			return
		}
		defer cursor.Close(ctx)

		var expenses []Expense
		if err = cursor.All(ctx, &expenses); err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Mongo cursor decoding failed"})
			return
		}

		c.IndentedJSON(http.StatusOK, gin.H{"status": true, "message": "Fetched expenses successfully", "data": expenses})
	}
}

func UpdateExpense() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error

		objID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Invalid ObjectID provided"})
			return
		}

		var updateExpense Expense

		if err = c.BindJSON(&updateExpense); err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Binding failed"})
			return
		}

		if err = Validator.Struct(updateExpense); err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Validation failed"})
			return
		}

		exp_Collection := MongoClient.Database("PocketPassbook").Collection("expenses_cl")
		result, err := exp_Collection.UpdateByID(ctx, objID, updateExpense)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": false, "message": "Mongo update failed"})
			return
		}
		if result.MatchedCount == 0 {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": false, "message": "Document not found"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"status": true, "message": "Updated successfully"})
	}
}

func DeleteExpense() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		var err error

		objID, err := primitive.ObjectIDFromHex(c.Param("id"))
		if err != nil {
			c.IndentedJSON(http.StatusExpectationFailed, gin.H{"error": err.Error(), "status": false, "message": "Invalid ObjectID provided"})
			return
		}

		exp_Collection := MongoClient.Database("PocketPassbook").Collection("expenses_cl")
		result, err := exp_Collection.DeleteOne(ctx, bson.M{"_id": objID})
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error(), "status": false, "message": "Mongo deleting failed"})
			return
		}
		if result.DeletedCount == 0 {
			c.IndentedJSON(http.StatusNotFound, gin.H{"status": false, "message": "Document not found"})
			return
		}
		c.IndentedJSON(http.StatusOK, gin.H{"status": true, "message": "Deleted successfully"})
	}
}
