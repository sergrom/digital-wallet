package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/nats-io/nats.go"
)

func natsGetBalanceHandler(msg *nats.Msg) {
	userID, err := strconv.ParseInt(string(msg.Data), 10, 64)
	if err != nil {
		log.Println(err)
		ncConn.Publish(msg.Reply, []byte("error: could not parse user_id"))
		return
	}

	users := make([]User, 0, 1)
	err = dbConn.Select(&users, `select * from "user" where user_id=$1 limit 1`, userID)
	if err != nil {
		log.Println(err)
		ncConn.Publish(msg.Reply, []byte("error: something went wrong"))
		return
	}
	if len(users) == 0 {
		log.Println("not found user with user_id =", userID)
		ncConn.Publish(msg.Reply, []byte("error: not found user with user_id ="+string(msg.Data)))
		return
	}

	ncConn.Publish(msg.Reply, []byte(fmt.Sprintf("%.2f", users[0].Balance)))
}
