FROM ubuntu:14.04

MAINTAINER eric.yongjun.choi@gmail.com

ENV DEBIAN_FRONTEND noninteractive

RUN apt-get update
RUN apt-get install -y redis-server

EXPOSE 6379

CMD ["--save 1 1", "--dir '/var/lib/redis'"]

ENTRYPOINT  ["/usr/bin/redis-server"]
