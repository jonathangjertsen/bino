FROM golang:trixie AS build

WORKDIR /

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Get sass
RUN apt-get update && apt-get install -y curl postgresql \
  && curl -L --silent https://github.com/sass/dart-sass/releases/download/1.94.0/dart-sass-1.94.0-linux-x64.tar.gz \
     | tar -xz -C /opt \
  && ln -s /opt/dart-sass/sass /usr/local/bin/sass \
  && rm -rf /var/lib/apt/lists/*

# Copy the stuff needed for SQLC to run with caching
COPY sqlc.yaml ./
COPY cmd ./cmd
RUN go tool sqlc generate

# Copy remaining files
COPY . .

RUN ./scripts/init_db.sh

# Generate files. Order is arbitrary but runs roughly in order of slowness
RUN sass styles.scss static/gen.css -q
RUN go tool templ generate
RUN go generate ./...

# Build application
RUN go build -ldflags="-X 'main.BuildKey=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 8)'" -o bino github.com/jonathangjertsen/bino/cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /bino .
COPY cmd/static ./cmd/static
COPY config.json ./config.json

EXPOSE 8080
ENTRYPOINT ["./bino"]
