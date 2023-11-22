FROM docker.io/library/golang:bookworm
COPY ./ /wgmesh
RUN apt-get update && apt-get install -y \
    wireguard \
    wireguard-tools \
    iproute2 \
    iputils-ping \
    tmux \
    vim
WORKDIR /wgmesh
RUN go mod tidy
RUN go build -o /usr/local/bin ./...
