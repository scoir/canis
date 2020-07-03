FROM scoir/libursa:latest

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata

COPY ./bin/agent /usr/local/bin/agent