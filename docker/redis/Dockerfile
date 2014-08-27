FROM ubuntu:14.04

MAINTAINER eric.yongjun.choi@gmail.com

RUN DEBIAN_FRONTEND=noninteractive apt-get update
RUN DEBIAN_FRONTEND=noninteractive apt-get install -y redis-server

EXPOSE 6379
CMD ["--save 1 1"]
ENTRYPOINT  ["/usr/bin/redis-server"]
