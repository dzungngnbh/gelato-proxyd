FROM golang:1.23.3-alpine3.20 as builder

ARG GITCOMMIT=docker
ARG GITDATE=docker
ARG GITVERSION=docker

RUN apk add make jq git gcc musl-dev linux-headers

COPY ./ /app

WORKDIR /app

RUN make proxyd

FROM alpine:3.20

RUN apk add bind-tools jq curl bash git redis

COPY ./entrypoint.sh /bin/entrypoint.sh

RUN apk update && apk add ca-certificates && \
    chmod +x /bin/entrypoint.sh

EXPOSE 8080

VOLUME /etc/proxyd

COPY --from=builder /app/bin/proxyd /bin/proxyd

ENTRYPOINT ["/bin/entrypoint.sh"]
CMD ["/bin/proxyd", "/etc/proxyd/proxyd.toml"]
