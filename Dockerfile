FROM debian:trixie-slim
RUN apt update && apt install -y \
  curl git golang
ENV GOTOOLCHAIN=go1.25.12
ARG VERSION=0.40.0
RUN curl -LO https://github.com/tinygo-org/tinygo/releases/download/v${VERSION}/tinygo_${VERSION}_amd64.deb
RUN dpkg -i tinygo_${VERSION}_amd64.deb
COPY ./go.mod /app/
COPY ./go.sum /app/
WORKDIR /app
RUN go mod download
ENV TAG=drv8311
COPY ./ /app/
# RUN go mod tidy
CMD ["bash", "-c", "tinygo build -target pico -tags ${TAG} -o output.uf2 . && cat output.uf2"]
