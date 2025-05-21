FROM --platform=linux/arm64 golang:1.24-bullseye

RUN apt-get update && apt-get install -y \
    libsdl2-dev \
    libsdl2-ttf-dev \
    libsdl2-image-dev \
    libsdl2-gfx-dev

WORKDIR /build

COPY go.mod go.sum* ./

RUN go mod download

COPY . .
RUN go build -v -gcflags="all=-N -l" -o pak-store app/pak_store.go

CMD ["/bin/bash"]