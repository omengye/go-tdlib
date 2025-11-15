FROM golang:1.25-bookworm AS builder

WORKDIR /
COPY . .
RUN apt-get update && apt-get install -y ca-certificates libssl-dev zlib1g-dev bash cmake gperf && rm -rf /var/lib/apt/lists/* \
    && mkdir /td/build && cd /td/build && cmake -DCMAKE_BUILD_TYPE=Release -DCMAKE_INSTALL_PREFIX:PATH=/td/tdlib .. && cmake --build . --target install

FROM golang:1.25-bookworm AS runner
RUN apt-get update && apt-get install -y ca-certificates libssl-dev zlib1g-dev bash && rm -rf /var/lib/apt/lists/*
COPY --from=builder /td/tdlib /td/tdlib
