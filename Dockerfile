FROM golang:1.25.3-alpine as builder


RUN apk add --no-cache git openssh-client

WORKDIR /app


COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN go build -o /app/govpn

# --- Final Stage ---
# Using the original alpine:latest image
FROM alpine:latest

# We already added this in your original file, which is correct.
RUN apk add --no-cache ca-certificates

WORKDIR /app

ENV SUB_PORT=8443
EXPOSE ${SUB_PORT}

# Copy the built binary from the builder stage
COPY --from=builder /app /app

# Optional but recommended: Copy certificates from the builder stage
# This ensures your final minimal image has the same root CAs as the build env.
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT [ "/app/govpn", "start" ]
