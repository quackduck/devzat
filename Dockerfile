# Use the official Alpine-based Golang image as the build stage
FROM golang:1.18-alpine AS builder

# Set the working directory
WORKDIR /go/src/github.com/quackduck/devzat

# Copy the project source code into the container
COPY . .

# Install project dependencies
RUN go mod download

# Build the project
RUN go build -o devzat .




# Use the latest Alpine Linux image as the base for the final image
FROM alpine:latest

# Install necessary tools
RUN apk add --no-cache openssh-server

# Copy the compiled binary to the final image
COPY --from=builder /go/src/github.com/quackduck/devzat /usr/local/bin/devzat

# Generate the SSH host key
RUN ssh-keygen -qN '' -f /etc/ssh/ssh_host_rsa_key

# Expose the SSH port
EXPOSE 22

# Configure and start the SSH service
RUN sed -i 's/#PermitRootLogin prohibit-password/PermitRootLogin yes/' /etc/ssh/sshd_config
CMD ["/usr/sbin/sshd", "-D"]

# Run the application (This line is commented out because the main service is the SSH server)
# CMD ["devzat"]