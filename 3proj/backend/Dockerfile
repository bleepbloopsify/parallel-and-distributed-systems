FROM golang:1.13-alpine3.10 as BUILD

WORKDIR /build

# We don't need this because the backend has no dependencies
# COPY ./go.mod ./go.sum ./
# RUN go mod download
COPY ./ ./
RUN go build .

FROM alpine:3.10

WORKDIR /opt
COPY --from=build /build/backend ./backend

CMD ["/opt/backend"]
