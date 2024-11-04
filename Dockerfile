FROM golang:1.23-alpine
WORKDIR /src

RUN apk add --no-cache make python3 py3-pip
RUN pip install datamodel-code-generator  --break-system-packages
ENV GOMODCACHE=/root/.cache/go-mod
ENV GOCACHE=/root/.cache/go-build
RUN apk add --no-cache make
COPY ./go.* ./
RUN --mount=type=cache,target=/root/.cache/go-mod \
  go mod download
COPY . .
RUN --mount=type=cache,target=/root/.cache/go-mod --mount=type=cache,target=/root/.cache/go-build \
  go build -o /src/app ./cmd/app

EXPOSE 8000

CMD ["/src/app"]