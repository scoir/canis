FROM scoir/indy-libs:latest

RUN apt-get update
RUN apt-get -y upgrade

RUN apt-get install -y build-essential
RUN apt-get install -y wget

RUN wget https://dl.google.com/go/go1.14.6.linux-amd64.tar.gz -q -P /tmp -nH

RUN tar -xzf /tmp/go1.14.6.linux-amd64.tar.gz -C /tmp
RUN mv /tmp/go /usr/local

ENV PATH /usr/local/go/bin:$PATH
RUN mkdir /canis
RUN mkdir /etc/canis