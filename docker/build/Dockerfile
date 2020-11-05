FROM canislabs/indy-libs:latest

RUN apt-get -y update --fix-missing
RUN apt-get -y upgrade

RUN apt-get install -y build-essential wget git curl protobuf-compiler unzip apt-transport-https ca-certificates gnupg-agent software-properties-common rsync

RUN curl -fsSL https://download.docker.com/linux/ubuntu/gpg | apt-key add -
RUN add-apt-repository \
       "deb [arch=amd64] https://download.docker.com/linux/ubuntu \
       $(lsb_release -cs) \
       stable"

RUN apt-get update
RUN apt-get install -y docker-ce docker-ce-cli containerd.io

RUN wget https://dl.google.com/go/go1.14.6.linux-amd64.tar.gz -q -P /tmp -nH
RUN tar -xzf /tmp/go1.14.6.linux-amd64.tar.gz -C /tmp
RUN mv /tmp/go /usr/local

RUN wget https://github.com/protocolbuffers/protobuf/releases/download/v3.13.0/protoc-3.13.0-linux-x86_64.zip

RUN curl https://raw.githubusercontent.com/helm/helm/master/scripts/get-helm-3 | bash

ENV PATH /usr/local/go/bin:$PATH
RUN mkdir /canis
RUN mkdir /etc/canis

RUN unzip protoc-3.13.0-linux-x86_64.zip -d /canis/
