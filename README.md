# Devchat
Chat over SSH

Try it out:

```sh
ssh sshchat.hackclub.com -p 2222
```

Add this to `~/.ssh/config`:
```json
Host chat
    HostName sshchat.hackclub.com
    Port 2222

```

Now you can join with just:
```sh
ssh chat
```
