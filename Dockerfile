FROM scratch
MAINTAINER DeedleFake

EXPOSE 8080

COPY ps2avglogin /ps2avglogin
COPY ca-certificates.crt /etc/ssl/certs/

WORKDIR /data
ENTRYPOINT ["/ps2avglogin"]
