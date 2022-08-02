# Build stage
FROM golang AS builder

WORKDIR /foreman

COPY . .

RUN go mod download

RUN go build -o foreman *.go

# Run stage
FROM ubuntu

WORKDIR /foreman

COPY --from=builder /foreman/foreman .
COPY --from=builder /foreman/procfile.yml .

CMD ["./foreman"]
