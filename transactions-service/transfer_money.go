package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func transferMoneyHandler(c *gin.Context) {
	var req TransferMoneyReq
	err := c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "you must provide a json with fields: from_user_id (int), to_user_id (int), amount_to_transfer (float)"})
		return
	}

	if req.FromUserID == req.ToUserID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "from_user_id and to_user_id must be different"})
		return
	}
	if req.AmountToTransfer <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount_to_transfer must geater than 0"})
		return
	}

	tx, err := dbConn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}
	defer tx.Rollback() // Defer a rollback in case anything fails.

	rows, err := tx.QueryContext(c.Request.Context(), `select * from "user" where user_id = $1 or user_id = $2 for update`, req.FromUserID, req.ToUserID)

	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	users := make(map[int64]User)
	for rows.Next() {
		var u User
		err := rows.Scan(&u.UserID, &u.Balance, &u.CreatedAt)
		if err != nil {
			log.Println(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
			return
		}
		users[u.UserID] = u
	}

	if len(users) < 2 {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "not found one/both users"})
		return
	}

	if users[req.FromUserID].Balance-req.AmountToTransfer < 0 {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you cannot transfer money from user_id:%d to user_id:%d, because balance of user_id:%d will turn negative", req.FromUserID, req.ToUserID, req.FromUserID)})
		return
	}

	_, err = tx.ExecContext(c.Request.Context(), `update "user" set balance=$1 where user_id=$2`, users[req.FromUserID].Balance-req.AmountToTransfer, req.FromUserID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	_, err = tx.ExecContext(c.Request.Context(), `update "user" set balance=$1 where user_id=$2`, users[req.ToUserID].Balance+req.AmountToTransfer, req.ToUserID)
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

	c.JSON(http.StatusOK, gin.H{
		"message": "success",
	})
}
