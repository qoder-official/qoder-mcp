FROM scratch
MAINTAINER  Florian Forster <fforster@gitlab.com>

# The binary is built by goreleaser outside of the Dockerfile.
COPY gitlab-mcp /usr/bin/gitlab-mcp

# Copy CA certificates from Alpine.
COPY --from=alpine:latest /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/usr/bin/gitlab-mcp"]
