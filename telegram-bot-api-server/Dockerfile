FROM ubuntu:24.04 AS builder

WORKDIR /build

RUN apt-get update && \
    apt-get upgrade -y && \
    apt-get install -y \
    make \
    git \
    zlib1g-dev \
    libssl-dev \
    gperf \
    cmake \
    g++

RUN git clone --recursive https://github.com/tdlib/telegram-bot-api.git && \
    cd telegram-bot-api && \
    mkdir build && \
    cd build && \
    cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=.. .. && \
    cmake --build . --target install

FROM ubuntu:24.04

COPY --from=builder /build/telegram-bot-api/bin/telegram-bot-api /usr/local/bin/

ENTRYPOINT ["telegram-bot-api"]