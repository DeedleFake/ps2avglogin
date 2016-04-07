FROM golang
MAINTAINER DeedleFake

COPY . /go/src/github.com/DeedleFake/ps2avglogin/
RUN go get -v github.com/DeedleFake/ps2avglogin

RUN mkdir -p /data
WORKDIR /data

EXPOSE 8080
CMD ["ps2avglogin"]
