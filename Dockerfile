# Use Ubuntu 22.04 as the base image
FROM ubuntu:22.04

# Set non-interactive environment variables to suppress any interactive prompts during package installation
ENV DEBIAN_FRONTEND=noninteractive

# Update package lists and install required packages
RUN apt-get update && apt-get install -y --no-install-recommends \
    iproute2 \
    openssh-server \
    git \
    golang \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

# Set up an independent IP address (assuming you want to use a fixed IP address for the container)
# Replace the placeholder IP address (X.X.X.X) with your desired IP address
RUN ip addr add X.X.X.X/24 dev eth0

# Open port 2221
EXPOSE 2221

# Clone the devzat repository from GitHub
RUN git clone https://github.com/quackduck/devzat /root/devzat

# Change directory to devzat
WORKDIR /root/devzat

# Check if Golang is at the latest version
RUN go version

# Install the devzat application
RUN go install

# Generate an SSH key for devzat
RUN ssh-keygen -qN '' -f devzat-sshkey

# Start the devzat application
CMD ["./devzat"]

# Print the message with instructions for SSH access
RUN echo "Welcome to devzat! SSH to our server by typing \"ssh <new_IP> -p 2221\""
