FROM golang:trixie AS build

WORKDIR /

# Download go dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy remaining files
COPY . .

#RUN ./scripts/init_db.sh

# Build application
RUN go build -ldflags="-X 'main.BuildKey=$(tr -dc A-Za-z0-9 </dev/urandom | head -c 8)'" -o bino github.com/jonathangjertsen/bino/cmd

FROM gcr.io/distroless/base-debian12

WORKDIR /

COPY --from=build /bino .
COPY cmd/static ./cmd/static
COPY config.json ./config.json

EXPOSE 8080
ENTRYPOINT ["./bino"]
