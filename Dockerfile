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
WORKDIR /app
COPY --from=build /service-account /app/service-account
COPY ./configs ./configs
COPY ./web/template ./web/template
EXPOSE 3000
# USER nonroot:nonroot
ENTRYPOINT [ "./service-account" ]