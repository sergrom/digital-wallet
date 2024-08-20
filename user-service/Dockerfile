FROM golang:1.22 AS build-stage

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /usrsvc

# Deploy the application binary into a lean image
FROM alpine:latest as production

WORKDIR /

COPY --from=build-stage /usrsvc /usrsvc

ENTRYPOINT ["/usrsvc"]