FROM golang
MAINTAINER DeedleFake

RUN go get -u -v github.com/DeedleFake/ps2avglogin

RUN mkdir -p /data
WORKDIR /data

EXPOSE 8080
CMD ps2avglogin
