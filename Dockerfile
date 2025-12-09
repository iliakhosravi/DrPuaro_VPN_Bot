FROM golang:1.25.3-alpine as builder


RUN apk add --no-cache git openssh-client

WORKDIR /app


COPY go.mod go.sum ./

# Copy the passed secrets into Docker's fs to be used by ssh
RUN mkdir ~/.ssh
RUN echo "Host github.com\n\tStrictHostKeyChecking no\n" >> /root/.ssh/config
RUN --mount=type=secret,id=idrsa cp /run/secrets/idrsa ~/.ssh/id_rsa
RUN --mount=type=secret,id=idrsapub cp /run/secrets/idrsapub ~/.ssh/id_rsa.pub

# Configure git to use the ssh-keys
RUN git config --global --add url."ssh://git@github.com".insteadOf "https://github.com"
ENV GIT_SSH_COMMAND="ssh -i ~/.ssh/id_rsa -o UserKnownHostsFile=/dev/null -o StrictHostKeyChecking=no"

RUN go mod download

COPY . .

RUN go build -o /app/techybat-bot

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

ENTRYPOINT [ "/app/techybat-bot" ]
