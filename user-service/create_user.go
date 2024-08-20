package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
	"github.com/gin-gonic/gin"
	"github.com/lib/pq"
)

func createUserHandler(c *gin.Context) {
	var req CreateUserReq
	err := c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "you must provide a json with fields: email (string)"})
		return
	}

	if req.Email == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "empty email"})
		return
	}

	tx, err := dbConn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}
	defer tx.Rollback() // Defer a rollback in case anything fails.

	createdAt := time.Now()
	var userID int64
	err = tx.QueryRowContext(c.Request.Context(), `insert into "user" (email, created_at) values ($1, $2) RETURNING user_id`, req.Email, createdAt).Scan(&userID)
	if err != nil {
		log.Println(err)
		if pqError, ok := err.(*pq.Error); ok && pqError.Code == "23505" {
			c.JSON(http.StatusConflict, gin.H{"error": "user with the same email already exists"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	msg := UserCreatedMsg{
		UserID:    userID,
		Email:     req.Email,
		CreatedAt: createdAt,
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	_, _, err = kafkaProducer.SendMessage(&sarama.ProducerMessage{
		Topic: kafkaUserCreatedTopic,
		Value: sarama.ByteEncoder(data),
	})
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"user_id": fmt.Sprintf("%d", userID),
	})
}
