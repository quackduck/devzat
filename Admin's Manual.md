# Devzat Admin's Manual

This document is for those who want to manage a self-hosted Devzat server.


## Installation
```shell
git clone https://github.com/quackduck/devzat
cd devzat
```
To compile Devzat, you will need to have Go installed with a minimum version of 1.17.

Now run `go install` to install the Devzat binary globally, or run `go build` to build and keep the binary in the working directory.

You may need to generate a new key pair for your server using the `ssh-keygen` command. When prompted, save as `devzat-sshkey` since this is the default location (it can be changed in the config).
While you can use the same key pair that your user account has, it is recommended to use a new key pair.

## Usage

```shell
./devchat # use without "./" for a global binary
```

Devzat listens on port 2221 for new SSH connections by default. Users can now join using `ssh -p 2221 <server-hostname>`.

Set the environment variable `PORT` to a different port number or edit your config to change what port Devzat listens for SSH connections on.

## Configuration

Devzat writes the default config file if one isn't found, so you do not need to make one before using Devzat. 

The default location Devzat looks for a config file is `devzat-config.yml` in the current directory. Alternatively, it uses the path set in the `DEVZAT_CONFIG` environment variable.

An example config file:
```yaml
port: 2221 # what port to host a server on ($PORT overrides this)
alt_port: 443 # an alternate port to avoid firewalls
profile_port: 5555 # what port to host profiling on (unimportant)
data_dir: devzat-data # where to store data such as bans and logs
key_file: devzat-sshkey # where the SSH private key is stored
integration_config: devzat-integrations.yml # where an integration config is stored (optional)
admins: # a list of admin IDs and notes about them
  1eab2de20e41abed903ab2f22e7ff56dc059666dbe2ebbce07a8afeece8d0424: 'Shok: school'
  7f0ee4cba8c8d886d654820c4ea09090dc12be00746b9a64b73faab3f83a85c6: 'Benjamin Smith: hackclub, github.com/Merlin04'
  09db0042df8c48488e034cd03dfd30dc61a7db35cd7bb8964cd246b958adc1b9: 'Arcade Wise: hackclub, github.com/l3gacyb3ta'
  12a9f108e7420460864de3d46610f722e69c80b2ac2fb1e2ada34aa952bbd73e: 'jmw: github.com/ciearius'
  15a0a99e4ece5e7a169a0b7df8df87c7a7805207681df908f92052d3d4103287: 'Emma Trzupek: t0rchedf3rn (RHS)'
  41fb7ecb8c216491cbe682f2a4c2964db6a9297f74ae48f0ec8bece1089ffec3: 'Leo, gh: GrandWasTaken, added cause idk'
  111d1c193354309b040854a74aeb15c985d1fbe4390128dd035fe5407c71f2fd: 'elitedino: github.com/elitedino'
  2433e7c03997d13f9117ded9e36cd2d23bddc4d588b8717c4619bedeb3b7e9ad: '@epic: github.com/TAG-Epic'
  ac416b02c106e7407e8e53e74b40d96d6f7c11e365c285a8ab825c219f443dcd: 'Tommy Pujol: hackclub'
  c90cf5ea8a9bf831a9024ecfd9876a7116a2382653a9ce84a6d80b4dcfa2f979: 'cole: github.com/ColeDrain'
  d6acd2f5c5a8ef95563883032ef0b7c0239129b2d3672f964e5711b5016e05f5: 'Arkaeriit: github.com/Arkaeriit'
  e9d47bb4522345d019086d0ed48da8ce491a491923a44c59fd6bfffe6ea73317: 'Arav Narula: twitter'
  f5c7f9826b6e143f6e9c3920767680f503f259570f121138b2465bb2b052a85d: 'Ella Xu: hackclub'
  f466ac6b6be43ba8efbac8406a34cf68f6843f2b79119a82726e4ad6e770ec7d: 'electronoob: electronoob.com'
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


### Enabling integrations

Devzat includes features that may not be needed by self-hosted instances. These are called integrations.

You can enable these integrations by setting the `integration_config` in your config file to some path:

```yaml
integration_config: devzat-integrations.yml
```
Now make a new file at that path. This is your integration config file.

#### Using the Slack integration

Devzat supports a bridge to Slack. You'll need Slack bot token so Devzat can post to and receive messages from Slack. Follow the guide [here](https://api.slack.com/authentication/basics) to get your token and add a Slack app to your workspace. Ensure it has read and write scopes. (feel free to make a [new issue](/issues) if something doesn't work).

Add your bot token to your integration config file. The `prefix` key defines what messages from Slack rendered in Devzat will be prefixed with. Find the channel ID of the channel you want to bridge to with a right-click on it in Slack.

```yaml
slack:
    token: xoxb-XXXXXXXXXX-XXXXXXXXXXXX-XXXXXXXXXXXXXXXXXXXXXXXX
    channel_id: XXXXXXXXXXX # usually starts with a C, but could be a G or D
    prefix: Slack
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

You can use both integrations together.

There are 3 environment variables you can set to quickly disable integrations on the command line:
* `DEVZAT_OFFLINE_TWITTER=true` will disable Twitter
* `DEVZAT_OFFLINE_SLACK=true` will disable Slack
* `DEVZAT_OFFLINE=true` will disable both integrations.
