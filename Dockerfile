FROM golang:trixie AS build

WORKDIR /app

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Get sass
RUN apt-get update && apt-get install -y curl postgresql \
  && curl -L --silent https://github.com/sass/dart-sass/releases/download/1.94.0/dart-sass-1.94.0-linux-x64.tar.gz \
     | tar -xz -C /opt \
  && ln -s /opt/dart-sass/sass /usr/local/bin/sass \
  && rm -rf /var/lib/apt/lists/*

# Copy all files
# TODO: there is no caching after this point, it would be good to run sqlc before this since it's so slow the first time
COPY . .

RUN ./scripts/init_db.sh

# Generate files. Order is arbitrary but runs roughly in order of slowness
RUN sass styles.scss static/gen.css -q
RUN go tool templ generate
RUN go generate ./...
RUN go tool sqlc generate

# Build application
RUN go build -ldflags="-X 'main.BuildKey=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 8)'" -o bino github.com/jonathangjertsen/bino

FROM gcr.io/distroless/base-debian12

WORKDIR /app

COPY --from=build /app/bino .
COPY static ./static

EXPOSE 8080
ENTRYPOINT ["./bino"]
