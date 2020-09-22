#!/usr/bin/env bash

set -e
echo "" >coverage.txt

# docker rm returns 1 if the image isn't found. This is OK and expected, so we suppress it
# Any return status other than 0 or 1 is unusual and so we exit
remove_docker_container() {
  docker kill PostgresStoreTest >/dev/null 2>&1 || true
  docker rm PostgresStoreTest >/dev/null 2>&1 || true
  docker kill MongoStoreTest >/dev/null 2>&1 || true
  docker rm MongoStoreTest >/dev/null 2>&1 || true
  docker kill CouchDBStoreTest >/dev/null 2>&1 || true
  docker rm CouchDBStoreTest >/dev/null 2>&1 || true
}

remove_docker_container

docker run -p 5432:5432 --name PostgresStoreTest -e POSTGRES_PASSWORD=mysecretpassword -d postgres:11.8 >/dev/null || true
docker run -p 27017:27017 --name MongoStoreTest -d mongo:4.2.8 >/dev/null || true
docker run -p 5984:5984 -d --name CouchDBStoreTest couchdb:2.3.1 >/dev/null || true

for d in $(go list ./pkg/... | grep -v vendor | grep -v mocks | grep -v cmd); do
  go test -race -coverprofile=profile.out -covermode=atomic $d
  if [ -f profile.out ]; then
    cat profile.out >>coverage.txt
    rm profile.out
  fi
done

remove_docker_container
