FROM ubuntu:14.04

MAINTAINER eric.yongjun.choi@gmail.com

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get -y --force-yes install wget make gcc libevent-dev
RUN wget --no-check-certificate https://github.com/nicolasff/webdis/archive/0.1.1.tar.gz -O webdis-0.1.1.tar.gz
RUN tar -xvzf webdis-0.1.1.tar.gz
RUN cd webdis-0.1.1 && make && make install && cd ..
RUN rm -rf webdis-0.1.1 webdis-0.1.1.tag.gz
RUN apt-get remove -y wget make gcc

EXPOSE 7379

ADD webdis.json /etc/webdis.json

CMD /usr/local/bin/webdis /etc/webdis.json
