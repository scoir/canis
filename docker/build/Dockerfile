FROM scoir/indy-libs:latest

RUN apt-get -y update
RUN apt-get -y upgrade

RUN apt-get install -y build-essential wget git curl protobuf-compiler unzip

RUN wget https://dl.google.com/go/go1.14.6.linux-amd64.tar.gz -q -P /tmp -nH
RUN tar -xzf /tmp/go1.14.6.linux-amd64.tar.gz -C /tmp
RUN mv /tmp/go /usr/local

RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip

ENV PATH /usr/local/go/bin:$PATH
RUN mkdir /canis
RUN mkdir /etc/canis

RUN unzip protoc-3.13.0-linux-x86_64.zip -d /canis/
