FROM golang:1.22-alpine
WORKDIR /src

RUN apk add --no-cache make python3 pipx
RUN pipx install datamodel-code-generator
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