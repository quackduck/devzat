# Devchat
Chat over SSH

Try it out:

```sh
ssh sshchat.hackclub.com
```

Add this to `~/.ssh/config`:
```json
Host chat
    HostName sshchat.hackclub.com
```

Now you can join with just:
```sh
ssh chat
```
