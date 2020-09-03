FROM scoir/indy-libs:latest

RUN apt-get update
RUN apt-get install -y ca-certificates
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y tzdata

COPY ./bin/canis-apiserver /usr/local/bin/canis-apiserver
COPY ./bin/canis-didcomm-issuer /usr/local/bin/canis-didcomm-issuer
COPY ./bin/canis-didcomm-doorman /usr/local/bin/canis-didcomm-doorman
COPY ./bin/canis-didcomm-lb /usr/local/bin/canis-didcomm-lb
COPY ./bin/agent /usr/local/bin/agent
COPY ./bin/sirius /usr/local/bin/sirius

WORKDIR /usr/local/canis
COPY ./static ./static

