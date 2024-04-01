FROM golang:1.21.1-bullseye AS build

WORKDIR /home/app
RUN echo "hello"

# FROM test.io/ignore:latest AS ignore

FROM gcr.io/distroless/base-nossl-debian11

LABEL test.test.test="test"

COPY --from=from-path /tmp /tmp

CMD ["run"]
