FROM ubuntu:18.04

LABEL AUTHOR="bleepbloopsify"

RUN apt -y update && apt -y upgrade && apt -y install golang

WORKDIR /opt/kittenstorage
COPY . .

RUN go build .

EXPOSE 8080
CMD ["/opt/kittenstorage/kittenstorage"]
