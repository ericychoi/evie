FROM google/golang

MAINTAINER eric.yongjun.choi@gmail.com

RUN go get -u github.com/ericychoi/evie

EXPOSE 55555
CMD ["--server"]
ENTRYPOINT ["gopath/bin/evie"]
