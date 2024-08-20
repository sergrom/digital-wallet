package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func balanceHandler(c *gin.Context) {
	userEmail := c.Request.URL.Query().Get("email")
	if userEmail == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "please specify an email in GET parameter"})
		return
	}

	var userID int64
	err := dbConn.Get(&userID, `select user_id from "user" where email = $1 limit 1`, userEmail)
	if err != nil {
		log.Println(err)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	msg, err := ncConn.Request(ncGetBalanceTopic, []byte(fmt.Sprintf("%d", userID)), time.Second)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	balanceStr := string(msg.Data)
	if strings.HasPrefix(balanceStr, "error") {
		fmt.Println(balanceStr)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	_, err = strconv.ParseFloat(balanceStr, 64)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"email":   userEmail,
		"balance": balanceStr,
	})
}
