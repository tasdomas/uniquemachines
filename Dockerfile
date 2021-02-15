FROM ubuntu:20.04 AS build-env

# Install deps
RUN apt-get -qq update && apt-get -qq install -y curl git gcc build-essential ca-certificates

# Install go
RUN curl -sS -O https://storage.googleapis.com/golang/go1.14.2.linux-amd64.tar.gz
RUN tar -C /usr/local -xzf go1.14.2.linux-amd64.tar.gz
ENV PATH /usr/local/go/bin:$PATH
ENV GO111MODULE=on

RUN mkdir /src
WORKDIR /src

COPY . .

RUN go build -o cl ./client

FROM ubuntu:20.04 AS client
RUN apt-get -qq update && apt-get -qq install -y ca-certificates
WORKDIR /root/
COPY --from=build-env /src/cl .
CMD ["./cl"]

FROM client AS clone
# create an existing token status file (imitate cloned machines)
RUN echo '{"machine-id":"this-is-a-clone","token":"some-token"}' > /tmp/token-status
CMD ["./cl"]
