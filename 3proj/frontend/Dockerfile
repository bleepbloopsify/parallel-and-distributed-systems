FROM golang:1.13-alpine3.10 as BUILD

WORKDIR /build

COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./*.go ./
RUN go build .

FROM alpine:3.10

WORKDIR /opt
COPY --from=build /build/frontend ./frontend
COPY ./views ./views

CMD ["/opt/frontend"]
