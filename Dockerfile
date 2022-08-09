# syntax=docker/dockerfile:1
# multi-stage build
##
## Build
##
# Alpine is chosen for its small footprint
# compared to Ubuntu
FROM golang:1.18-alpine AS build
WORKDIR /app
COPY ./ ./
# Download necessary Go modules
RUN go mod download
RUN go build -o /service-account ./cmd/service-account/main.go
##
## Deploy
##
FROM golang:1.18-alpine
WORKDIR /
COPY --from=build /service-account /service-account
COPY ./web/template ./web/template
EXPOSE 8080
# USER nonroot:nonroot
ENTRYPOINT [ "/service-account" ]