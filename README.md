# Digital Wallet (testing task for playtronix company)

Digital Wallet is a programm that maintains user wallet and performs money transfer between
users inside the platform

## Run:
```
docker-compose up
```
## User-service
Creating User:
```
curl -H "Content-Type: application/json" -X POST --data '{"email":"user1@example.com"}' http://localhost:9101/create-user

curl -H "Content-Type: application/json" -X POST --data '{"email":"user2@example.com"}' http://localhost:9101/create-user
```

Getting Balance:
```
curl -H "Content-Type: application/json" -X GET http://localhost:9101/balance?email=user1@example.com

curl -H "Content-Type: application/json" -X GET http://localhost:9101/balance?email=user2@example.com
```

## Transactions-service
Adding Money:
```
curl -H "Content-Type: application/json" -X POST --data '{"user_id":1, "amount":200}' http://localhost:9102/add-money

curl -H "Content-Type: application/json" -X POST --data '{"user_id":2, "amount":100}' http://localhost:9102/add-money
```
Transfer Money:
```
curl -H "Content-Type: application/json" -X POST --data '{"from_user_id":1,"to_user_id":2,"amount_to_transfer":50.00}' http://localhost:9102/transfer-money
```

## Also
###### UserService Database:
 - host: localhost
 - port: 6543
 - db: usrsvc
 - user: admin
 - pass: 123456
###### TransactionsService Database:
 - host: localhost
 - port: 6544
 - db: txsvc
 - user: admin
 - pass: 123456
###### Kafka:
 - host: localhost
 - port: 9093