FROM scoir/indy-libs:latest

RUN apt-get install -y ca-certificates
RUN apt-get update

COPY ./bin/canis-apiserver /usr/local/bin/canis-apiserver
COPY ./bin/canis-didcomm-issuer /usr/local/bin/canis-didcomm-issuer
COPY ./bin/canis-didcomm-verifier /usr/local/bin/canis-didcomm-verifier
COPY ./bin/canis-didcomm-doorman /usr/local/bin/canis-didcomm-doorman
COPY ./bin/canis-didcomm-lb /usr/local/bin/canis-didcomm-lb
COPY ./bin/http-indy-resolver /usr/local/bin/http-indy-resolver
COPY ./bin/canis-webhook-notifier /usr/local/bin/canis-webhook-notifier
COPY ./bin/sirius /usr/local/bin/sirius

WORKDIR /usr/local/canis
COPY ./static ./static

