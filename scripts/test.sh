#!/usr/bin/env bash

set -e
echo -n > coverage.txt

# docker rm returns 1 if the image isn't found. This is OK and expected, so we suppress it
# Any return status other than 0 or 1 is unusual and so we exit
remove_docker_container() {
  docker kill PostgresStoreTest >/dev/null 2>&1 || true
  docker rm PostgresStoreTest >/dev/null 2>&1 || true
  docker kill MongoStoreTest >/dev/null 2>&1 || true
  docker rm MongoStoreTest >/dev/null 2>&1 || true
  docker kill CouchDBStoreTest >/dev/null 2>&1 || true
  docker rm CouchDBStoreTest >/dev/null 2>&1 || true
  docker kill RabbitMQTest >/dev/null 2>&1 || true
  docker rm RabbitMQTest >/dev/null 2>&1 || true
}

remove_docker_container

docker run -p 5432:5432 --name PostgresStoreTest -e POSTGRES_PASSWORD=mysecretpassword -d postgres:11.8 >/dev/null || true
docker run -p 27017:27017 --name MongoStoreTest -d mongo:4.2.8 >/dev/null || true
docker run -p 5984:5984 -d --name CouchDBStoreTest couchdb:2.3.1 >/dev/null || true
docker run -d --hostname my-rabbit --name RabbitMQTest rabbitmq:3 > /dev/null || true

export RABBITMQ_HOST=${RABBITMQ_HOST:-localhost}
export MONGODB_HOST=${MONGODB_HOST:-localhost}

PKGS=$(go list github.com/scoir/canis/... 2> /dev/null | grep -v vendor | grep -v mocks | grep -v cmd | grep -v pb.go)
go test $PKGS -count=1 -race -coverprofile=profile.out -covermode=atomic -timeout=10m
if [ -f profile.out ]; then
  cat profile.out >>coverage.txt
  rm profile.out
fi

remove_docker_container
