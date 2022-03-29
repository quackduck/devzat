# Devzat Admin's Manual

This document is for those who want to manage a self-hosted Devzat server.


## Installation
```shell
git clone https://github.com/quackduck/devzat
cd devzat
```
Now run `go install` to install the Devzat binary globally, or run `go build` to build and keep the binary in the working directory.

## Usage

```shell
PORT=4242 ./devchat # use without "./" for a global binary
```

Use a different port number to change what port Devzat listens for SSH connections on.

Users can now join using `ssh -p <port> <server-hostname>`.

You may initially need to generate a key pair for your server, since Devzat looks for an SSH private key at `$HOME/.ssh/id_rsa`. Make a new key pair if needed using the `ssh-keygen` command.

### Using admin power

As an admin, you can ban, unban and kick users. When logged into the chat, you can run commands like these:
```shell
ban <user>
ban <user> 1h10m
unban <user ID or IP>
kick <user>
```

If running these commands makes Devbot complain about authorization, you need to be added to the `admins.json` file.

## Configuration

### Adding admins

Admins are defined in an `admins.json` file in the working directory.
The format is an ID string followed by any notes about the admin.
```json
{
  "ff7d1586cdecb9fbd9fcd4c9548522493c29172bc3121d746c83b28993bd723e": "Ishan Goel - quackduck",
  "d6acd2f5c5a8ef95563883032ef0b7c0239129b2d3672f964e5711b5016e05f5": "Arkaeriit: github.com/Arkaeriit"
}
```

### Disabling integrations

Devzat includes features that may not be needed by self-hosted instances.

Disable Twitter integration by exporting the environment variable `DEVZAT_OFFLINE_TWITTER=true`.

Disable Slack integration by exporting `DEVZAT_OFFLINE_SLACK=true`.

Disable all network usage except for fetching images using `DEVZAT_OFFLINE=true`.
