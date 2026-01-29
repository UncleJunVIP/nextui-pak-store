FROM golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libsdl2-dev \
    libsdl2-ttf-dev \
    libsdl2-image-dev \
    libsdl2-gfx-dev

WORKDIR /build

COPY go.mod go.sum* ./

RUN GOWORK=off go mod download

COPY . .

ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN GOWORK=off go build -v \
    -ldflags "-X github.com/UncleJunVIP/nextui-pak-store/version.Version=${VERSION} \
              -X github.com/UncleJunVIP/nextui-pak-store/version.GitCommit=${GIT_COMMIT} \
              -X github.com/UncleJunVIP/nextui-pak-store/version.BuildDate=${BUILD_DATE}" \
    -o pak-store app/pak_store.go

CMD ["/bin/bash"]
