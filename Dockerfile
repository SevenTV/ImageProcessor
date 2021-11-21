FROM ghcr.io/seventv/libwebp:latest as libwebp

FROM ghcr.io/seventv/libavif:latest as libavif

FROM ghcr.io/seventv/gifsicle:latest as gifsicle

FROM ghcr.io/seventv/gifski:latest as gifski

FROM golang:1.17.3-alpine as builder

WORKDIR /tmp/images

COPY . .

ARG BUILDER
ARG VERSION

ENV IMAGES_BUILDER=${BUILDER}
ENV IMAGES_VERSION=${VERSION}

RUN apk add --no-cache make git && \
    make linux

FROM ghcr.io/seventv/ffmpeg

RUN apk add --no-cache optipng vips-tools

COPY --from=libwebp /libwebp/cwebp /usr/bin
COPY --from=libwebp /libwebp/dwebp /usr/bin
COPY --from=libwebp /libwebp/webpmux /usr/bin
COPY --from=libwebp /libwebp/img2webp /usr/bin
COPY --from=libwebp /libwebp/anim_dump /usr/bin

COPY --from=libavif /libavif/avifdump /usr/bin
COPY --from=libavif /libavif/avifdec /usr/bin
COPY --from=libavif /libavif/avifenc /usr/bin

COPY --from=gifsicle /gifsicle/gifsicle /usr/bin
COPY --from=gifski /gifski/target/release/gifski /usr/bin

WORKDIR /app

COPY --from=builder /tmp/images/bin/images .

ENTRYPOINT ["./images"]
