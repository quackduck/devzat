# Devzat Admin's Manual

This document is for those who want to manage a self-hosted Devzat server.

Feel free to make a [new issue](https://github.com/quackduck/devzat/issues) if something doesn't work.

## Installation
```shell
git clone https://github.com/quackduck/devzat
cd devzat
```
To compile Devzat, you will need Go installed with a minimum version of 1.17.

Now run `go install` to install the Devzat binary globally, or run `go build` to build and keep the binary in the working directory.

You may need to generate a new key pair for your server using the `ssh-keygen` command. When prompted, save as `devzat-sshkey` since this is the default location (it can be changed in the config).
While you can use the same key pair that your user account has, it is recommended to use a new key pair.

## Usage

```shell
./devzat # use without "./" for a global binary
```

Devzat listens on port 2221 for new SSH connections by default. Users can now join using `ssh -p 2221 <server-hostname>`.

Set the environment variable `PORT` to a different port number or edit your config to change what port Devzat listens for SSH connections on. Users would then run `ssh -p <port> <server-hostname>` to join.

## Configuration

Devzat writes the default config file if one isn't found, so you do not need to make one before using Devzat. 

The default location Devzat looks for a config file is `devzat.yml` in the current directory. Alternatively, it uses the path set in the `DEVZAT_CONFIG` environment variable.

An example config file:
```yaml
# what port to host a server on ($PORT overrides this)
port: 2221
# an alternate port to avoid firewalls
alt_port: 443
# what port to host profiling on (unimportant)
profile_port: 5555
# where to store data such as bans and logs
data_dir: devzat-data
# where the SSH private key is stored
key_file: devzat-sshkey
# where an integration config is stored (optional)
integration_config: devzat-integrations.yml
# whether to censor messages (optional)
censor: true
# a list of admin IDs and notes about them
admins:
  d6acd2f5c5a8ef95563883032ef0b7c0239129b2d3672f964e5711b5016e05f5: 'Arkaeriit: github.com/Arkaeriit'
  ff7d1586cdecb9fbd9fcd4c9548522493c29172bc3121d746c83b28993bd723e: 'Ishan Goel: quackduck'
```

### Using admin power

As an admin, you can ban, unban and kick users. When logged into the chat, you can run commands like these:
```shell
ban <user>
ban <user> 1h10m
unban <user ID or IP>
kick <user>
```

If running these commands makes Devbot complain about authorization, you need to add your ID under the `admins` key in your config file (`devzat-config.yml` by default).

### Enabling a user allowlist

Devzat can use be used as a private chatroom. Add this to your config:

```yaml
private: true # enable allowlist checking
allowlist: 
  272b326d7d5e9a6b1d98a10b453bdc8cc950fc15cae2c2e858e30645c72ae7c0: 'John Doe'
  ...
```

The `allowlist` has the same format as the `admins` list. Add the IDs of the allowed users and info about that user (this is to make IDs easier to identify when editing the config file, and isn't used by Devzat)

All admins are allowed even if their ID is not in the allowlist. So, if everyone on the private server is an admin, an allowlist isn't necessary, just enable private mode.

Message backlog on `#main` is disabled in private chats. Only those logged in at the same time as you can read your messages.

### Enabling integrations

Devzat includes features that may not be needed by self-hosted instances. These are called integrations.

You can enable these integrations by setting the `integration_config` in your config file to some path:

```yaml
integration_config: devzat-integrations.yml
```
Now make a new file at that path. This is your integration config file.

#### Using the Slack integration

Devzat supports a bridge to Slack. You'll need a Slack bot token so Devzat can post to and receive messages from Slack. Follow the guide [here](https://api.slack.com/authentication/basics) to get your token and add a Slack app to your workspace. Ensure it has read and write scopes.

Add your bot token to your integration config file. The `prefix` key defines what messages from Slack rendered in Devzat will be prefixed with. Find the channel ID of the channel you want to bridge to with a right-click on it in Slack.

```yaml
slack:
    token: xoxb-XXXXXXXXXX-XXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXX
    channel_id: XXXXXXXXXXX # usually starts with a C, but could be a G or D
    prefix: Slack
```

#### Using the Discord integration

Devzat supports a bridge to Discord. You'll need a Discord bot token so Devzat can post to and receive messages from Discord. Follow the guide [here](https://www.writebots.com/discord-bot-token) to set up your bot.

Add your bot token to your integration config file. The `prefix` key defines what messages from Discord rendered in Devzat will be prefixed with. Find the channel ID of the channel you want to bridge to with a right-click on it.

```yaml
discord:
    token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    channel_id: XXXXXXXXXXXXXXXXXXX
    prefix: Discord
```

#### Using the Twitter integration

Devzat supports sending updates about who is online to Twitter. You need to make a new Twitter app through a [Twitter developer account](https://developer.twitter.com/en/apply/user)

Now add in the relevant keys to your integration config file:
```yaml
twitter:
    consumer_key: XXXXXXXXXXXXXXXXXXXXXXXXX
    consumer_secret: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    access_token: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
    access_token_secret: XXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXXX
```

### Using the plugin API integration

Devzat includes a built-in gRPC plugin API. This is useful for building your own integration or using a third-party one.

Documentation for using the gRPC API is available [here](plugin/README.md). This integration stores API tokens inside the data directory.

```yaml
rpc:
    port: 5556 # port to listen on for gRPC clients
```

Use the token issuing commands detailed in the [plugin documentation](plugin/README.md) to allow clients to authenticate.

You may also hard-code a key that can be used as an authentication token, but this is not recommended. This option can be useful for server owners trying to keep some API clients always online, since this key cannot be revoked by admins (unlike tokens).

```yaml
    key: "some string" # a string that gRPC clients authenticate with (optional)
```

You can use any number of integrations together.

There are 4 environment variables you can set to quickly disable integrations on the command line:
* `DEVZAT_OFFLINE_TWITTER=true` will disable Twitter
* `DEVZAT_OFFLINE_SLACK=true` will disable Slack
* `DEVZAT_OFFLINE_RPC=true` will disable the gRPC server
* `DEVZAT_OFFLINE=true` will disable all integrations.

