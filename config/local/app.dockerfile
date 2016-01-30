FROM golang

RUN apt-get update \
  && apt-get install -y --no-install-recommends \
    inotify-tools \
  && rm -rf /var/lib/apt/lists/*

RUN go get \
  golang.org/x/crypto/ssh \
  github.com/stretchr/testify/assert

WORKDIR /app
VOLUME /app
ENV GO_PATH=/app
ENTRYPOINT ["./shell/config/local/up.sh"]