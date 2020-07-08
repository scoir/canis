FROM scoir/libursa:latest

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata

COPY ./bin/steward /usr/local/bin/steward
COPY ./bin/agent /usr/local/bin/agent
COPY ./bin/canisctl /usr/local/bin/canisctl

WORKDIR /usr/local/canis
COPY ./static ./static

