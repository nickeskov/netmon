FROM golang:1.18.3-alpine3.16 AS builder
ARG APP=/app
RUN apk add make
# disable cgo for go build
ENV CGO_ENABLED=0
WORKDIR ${APP}
COPY go.mod .
COPY go.sum .
RUN go mod download
COPY cmd cmd
COPY pkg pkg
COPY Makefile .
RUN make vendor build
ENV TZ=Etc/UTC \
    APP_USER=netmon
RUN addgroup -S $APP_USER \
    && adduser -S $APP_USER -G $APP_USER

FROM scratch AS main
ARG APP=/app
ENV TZ=Etc/UTC \
    APP_USER=netmon

COPY --from=builder /etc/ssl/certs /etc/ssl/certs/
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
USER $APP_USER
EXPOSE 2048
WORKDIR ${APP}
COPY --from=builder /app/build/netmon netmon
ENTRYPOINT ["./netmon"]
