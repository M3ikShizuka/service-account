#!/bin/bash
# Test data: database='postgres://${SERVICE_ACCOUNT_POSTGRES_USER}:${SERVICE_ACCOUNT_POSTGRES_PASSWORD}@127.0.0.1:5433/${SERVICE_ACCOUNT_POSTGRES_DB}?sslmode=disable'
database=$1
until ./migrate -database $database -path db up
  do
    echo "waiting for database";
    sleep 2;
  done;