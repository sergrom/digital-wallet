package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func addMoneyHandler(c *gin.Context) {
	var req AddMoneyReq
	err := c.BindJSON(&req)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "you must provide a json with fields: user_id (int), amount (float)"})
		return
	}

	tx, err := dbConn.BeginTx(c.Request.Context(), nil)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}
	defer tx.Rollback() // Defer a rollback in case anything fails.

	row := tx.QueryRowContext(c.Request.Context(), `select * from "user" where user_id=$1 limit 1`, req.UserID)
	if row.Err() != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	var user User
	if err = row.Scan(&user.UserID, &user.Balance, &user.CreatedAt); err != nil {
		log.Println(err)
		if errors.Is(err, sql.ErrNoRows) {
			c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	if user.Balance+req.Amount < 0 {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("you cannot subtract %.2f amount because balance will turn negative", req.Amount)})
		return
	}

	_, err = tx.ExecContext(c.Request.Context(), `update "user" set balance=balance+$1 where user_id = $2`, req.Amount, req.UserID)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusOK, gin.H{"error": "something went wrong, please try later"})
		return
	}

	// Commit the transaction.
	if err = tx.Commit(); err != nil {
		log.Println(err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "something went wrong, please try later"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"updated_balance": fmt.Sprintf("%.2f", user.Balance+req.Amount),
	})
}
