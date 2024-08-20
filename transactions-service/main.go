package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/nats-io/nats.go"
)

const (
	ncURL                 = "nats://nats:4222"
	ncGetBalanceTopic     = "get-balance"
	kafkaAddr             = "kafka:9092"
	kafkaUserCreatedTopic = "user-created"
	kafkaConsumerGroupID  = "playtronix"
)

var (
	dbConn *sqlx.DB
	ncConn *nats.Conn
)

func main() {
	appPort := flag.String("p", "9102", "the port in which application will be run")
	flag.Parse()

	router := gin.Default()
	router.POST("/add-money", addMoneyHandler)
	router.POST("/transfer-money", transferMoneyHandler)

	dbConn = dbConnect()
	ncConn = natsConnect()

	initDBSchema()

	ncConn.Subscribe(ncGetBalanceTopic, natsGetBalanceHandler)

	go consumeUserCreated()

	serv := &http.Server{
		Addr:        ":" + *appPort,
		Handler:     router,
		ReadTimeout: 3 * time.Second,
	}

	if err := serv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}
}

func dbConnect() *sqlx.DB {
	dbConn, err := sqlx.Connect("postgres", "postgresql://admin:123456@db_txsvc:5432/txsvc?sslmode=disable")
	if err != nil {
		log.Fatalln("could not connect to database", err)
	}
	return dbConn
}

func natsConnect() *nats.Conn {
	ncConn, err := nats.Connect(ncURL)
	if err != nil {
		log.Fatalf("could not connect to nats on: %s\n", ncURL)
	}
	return ncConn
}

func initDBSchema() {
	_, err := dbConn.Exec(`
		create table if not exists "user" (
			"user_id"     bigint not null,
			"balance"     decimal(18, 2) not null,
			"created_at"  timestamp not null default now(),
			unique (user_id)
		)`)
	if err != nil {
		log.Fatalln("initDBSchema failed:", err)
	}
	_, err = dbConn.Exec(`create index if not exists idx_user_user_id on "user" using hash ("user_id")`)
	if err != nil {
		log.Fatalln("initDBSchema failed:", err)
	}
}
