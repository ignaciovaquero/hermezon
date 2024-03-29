FROM --platform=$BUILDPLATFORM golang:1.15.6-alpine3.12 AS builder

ARG BUILDPLATFORM
ARG TARGETOS
ARG TARGETARCH
ARG HERMEZON_VERSION=${HERMEZON_VERSION:-0.1.0-SNAPSHOT}
ARG HERMEZON_LISTEN_PORT=${HERMEZON_LISTEN_PORT:-8080}

WORKDIR /go/src/github.com/igvaquero18
COPY . .
RUN GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -o /go/bin/hermezon


FROM alpine:3.12.3

LABEL org.opencontainers.image.authors "Ignacio Vaquero Guisasola <ivaqueroguisasola@gmail.com>" \
      org.opencontainers.image.version "${HERMEZON_VERSION}"

RUN addgroup -S hermezon && \
    adduser -S hermezon -G hermezon

USER hermezon:hermezon

COPY --chown=hermezon:hermezon --from=builder /go/bin/hermezon /go/bin/hermezon

EXPOSE ${HERMEZON_LISTEN_PORT}

ENV HERMEZON_TWILIO_ACCOUNT_SID= \
    HERMEZON_TWILIO_ACCOUNT_TOKEN= \
    HERMEZON_EXPECTED_STATUS_CODE= \
    HERMEZON_MAX_RETRIES= \
    HERMEZON_RETRY_SECONDS= \
    HERMEZON_VERBOSE= \
    HERMEZON_LISTEN_PORT=${HERMEZON_LISTEN_PORT} \
    HERMEZON_PRICE_SCHEDULE_FREQUENCY= \
    HERMEZON_AVAILABILITY_SCHEDULE_FREQUENCY= \
    HERMEZON_JWT_SECRET= \
    HERMEZON_DB_FILE_PATH= \
    HERMEZON_TWILIO_PHONE=

ENTRYPOINT [ "/go/bin/hermezon" ]
