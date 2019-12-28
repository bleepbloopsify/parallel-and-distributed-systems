FROM golang:1.13.1-alpine3.10 as BUILD

WORKDIR /build
COPY ./ ./
RUN go build .

FROM alpine:3.10

WORKDIR /opt
COPY --from=BUILD /build/backend ./backend

ENV ADDR 8090
EXPOSE ${ADDR}
CMD /opt/backend --listen :$ADDR