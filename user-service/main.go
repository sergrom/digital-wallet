package main

import (
	"flag"
	"log"
	"net/http"
	"time"

	"github.com/IBM/sarama"
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
)

var (
	dbConn        *sqlx.DB
	ncConn        *nats.Conn
	kafkaProducer sarama.SyncProducer
)

func main() {
	appPort := flag.String("p", "9101", "the port in which application will be run")
	flag.Parse()

	router := gin.Default()
	router.POST("/create-user", createUserHandler)
	router.GET("/balance", balanceHandler)

	dbConn = dbConnect()
	ncConn = natsConnect()
	kafkaProducer = kafkaInitProducer()

	initDBSchema()

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
	dbConn, err := sqlx.Connect("postgres", "postgresql://admin:123456@db_usrsvc:5432/usrsvc?sslmode=disable")
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

func kafkaInitProducer() sarama.SyncProducer {
	producer, err := sarama.NewSyncProducer([]string{kafkaAddr}, nil)
	if err != nil {
		log.Fatalln("could not create kafka producer:", err)
	}

	return producer
}

func initDBSchema() {
	_, err := dbConn.Exec(`
		create table if not exists "user" (
			"user_id"     bigserial primary key,
			"email"       varchar(255) not null,
			"created_at"  timestamp not null default now(),
			unique (email)
		)`)
	if err != nil {
		log.Fatalln("initDBSchema failed:", err)
	}

	_, err = dbConn.Exec(`create index if not exists idx_user_email on "user" using hash ("email")`)
	if err != nil {
		log.Fatalln("initDBSchema failed:", err)
	}
}
