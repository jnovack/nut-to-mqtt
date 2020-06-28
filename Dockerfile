# build stage
FROM golang:1.14-alpine AS build
WORKDIR /go/src/app

# Create appuser.
# See https://stackoverflow.com/a/55757473/12429735RUN
ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apk add --no-cache git ca-certificates

COPY go.mod .
COPY go.sum .
RUN go mod download

ARG APPLICATION="go-executable"
ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"
ARG VERSION="dirty"
ARG GO_LDFLAGS="-w -s \
        -X github.com/jnovack/go-version.Application=${APPLICATION} \
        -X github.com/jnovack/go-version.BuildDate=${BUILD_RFC3339} \
        -X github.com/jnovack/go-version.Revision=${COMMIT} \
        -X github.com/jnovack/go-version.Version=${VERSION} \
        -extldflags '-static'"

# Build
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "${GO_LDFLAGS}" -o /go/bin/${APPLICATION} cmd/*/*

###############################################################################
# final stage
FROM scratch
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
USER appuser:appuser

ARG APPLICATION="go-executable"
ARG BUILD_RFC3339="1970-01-01T00:00:00Z"
ARG COMMIT="local"
ARG VERSION="dirty"

LABEL org.opencontainers.image.ref.name="jnovack/${APPLICATION}" \
      org.opencontainers.image.created=$BUILD_RFC3339 \
      org.opencontainers.image.authors="Justin J. Novack <jnovack@gmail.com>" \
      org.opencontainers.image.documentation="https://github.com/jnovack/${APPLICATION}/README.md" \
      org.opencontainers.image.description="" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/jnovack/${APPLICATION}" \
      org.opencontainers.image.revision=$COMMIT \
      org.opencontainers.image.version=$VERSION \
      org.opencontainers.image.url="https://hub.docker.com/r/jnovack/${APPLICATION}/"

COPY --from=build /go/bin/${APPLICATION} /app

ENTRYPOINT ["/app"]
