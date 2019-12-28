FROM golang:1.13.1-alpine3.10 as BUILD

WORKDIR /build
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./*.go ./
RUN go build .

FROM alpine:3.10

WORKDIR /opt
COPY --from=BUILD /build/frontend ./frontend
COPY ./views ./views

ENV ADDR 8080
ENV BACKEND "backend:8090"
EXPOSE ${ADDR}
CMD /opt/frontend --listen :$ADDR --backend $BACKEND
