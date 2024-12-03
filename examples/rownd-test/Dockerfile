FROM golang:1.21-alpine

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy the source code
COPY . .

# Build the application
RUN go build -o /rownd-server ./server

# Copy static files
COPY client/static ./client/static

# Expose port 8080
EXPOSE 8080

# Run the application
CMD ["/rownd-server"]