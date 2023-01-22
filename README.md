# Devzat

<a href="https://www.producthunt.com/posts/devzat?utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-devzat" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=298678&theme=light&period=daily" alt="Devzat - Chat with other devs over SSH in your Terminal | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

Where are the devs at? Devzat!

Devzat is a custom SSH server that takes you to a chat instead of a shell prompt. Because there's SSH apps on all platforms (even on phones) you can connect to Devzat on any device!

<!-- <img src="https://user-images.githubusercontent.com/38882631/115499526-a4d70280-a280-11eb-8723-817f54eccf3e.png" height=400px /> -->

A recording I took one day:
[![asciicast](https://asciinema.org/a/477083.svg)](https://asciinema.org/a/477083?speed=3)
## Usage

Try it out:

```sh
ssh devzat.hackclub.com
```

You can log in with a nickname:
```sh
ssh nickname@devzat.hackclub.com
```

If you're under a firewall, you can still join on port 443:
```sh
ssh devzat.hackclub.com -p 443
```

If you add this to `~/.ssh/config`:
```ssh
Host chat
    HostName devzat.hackclub.com
```

You'll be able to join with just:
```sh
ssh chat
```

We also have a Slack bridge! If you're on the [Hack Club](https://hackclub.com) Slack, check out the `#ssh-chat-bridge` channel!

Feel free to make a [new issue](https://github.com/quackduck/devzat/issues) if something doesn't work.

### Want to host your own instance?

Quick start:
```shell
git clone https://github.com/quackduck/devzat && cd devzat
go install # or build, if you want to keep things pwd
ssh-keygen -qN '' -f devzat-sshkey # new ssh host key for the server
devzat # run! the default config is used & written automatically
```
These commands download, build, setup and run a Devzat server listening on port 2221, the default port (change by setting `$PORT`).

Check out the [Admin's Manual](Admin's%20Manual.md) for complete self-host documentation!

### Permission denied?

Devzat uses public keys to identify users. If you are denied access: `foo@devzat.hackclub.com: Permission denied (publickey)` try logging in on port 443, which does not require a key, using `ssh devzat.hackclub.com -p 443`.

This error may happen because you do not have an SSH key pair. Generate one with the command `ssh-keygen` if this is the case. (you can usually check if you have a key pair by making sure a file of this form: `~/.ssh/id_*` exists)

### Help

```text
Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Run `cmds` to see a list of commands.

Interesting features:
• Rooms! Run cd to see all rooms and use cd #foo to join a new room.
• Markdown support! Tables, headers, italics and everything. Just use \n in place of newlines.
• Code syntax highlighting. Use Markdown fences to send code. Run eg-code to see an example.
• Direct messages! Send a quick DM using =user <msg> or stay in DMs by running cd @user.
• Timezone support, use tz Continent/City to set your timezone.
• Built in Tic Tac Toe and Hangman! Run tic or hang <word> to start new games.
• Emoji replacements! :rocket: => 🚀  (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Made by Ishan Goel with feature ideas from friends.
Thanks to Caleb Denio for lending his server!
```
### Commands
```text
Commands
   =<user>   <msg>           DM <user> with <msg>
   users                     List users
   color     <color>         Change your name's color
   exit                      Leave the chat
   help                      Show help
   man       <cmd>           Get help for a specific command
   emojis                    See a list of emojis
   bell      on|off|all      ANSI bell on pings (on), never (off) or for every message (all)
   clear                     Clear the screen
   hang      <char|word>     Play hangman
   tic       <cell num>      Play tic tac toe!
   devmonk                   Test your typing speed
   cd        #room|user      Join #room, DM user or run cd to see a list
   tz        <zone> [24h]    Set your IANA timezone (like tz Asia/Dubai) and optionally set 24h
   nick      <name>          Change your username
   pronouns  @user|pronouns  Set your pronouns or get another user's
   theme     <theme>|list    Change the syntax highlighting theme
   rest                      Uncommon commands list
   cmds                      Show this message
```
```
The rest
   people                  See info about nice people who joined
   id       <user>         Get a unique ID for a user (hashed key)
   admins                  Print the ID (hashed key) for all admins
   eg-code  [big]          Example syntax-highlighted code
   lsbans                  List banned IDs
   ban      <user>         Ban <user> (admin)
   unban    <IP|ID> [dur]  Unban a person and optionally, for a duration (admin)
   kick     <user>         Kick <user> (admin)
   art                     Show some panda art
   pwd                     Show your current room
   shrug                   ¯\_(ツ)_/¯
```

## Integrations

When self-hosting an instance, Devzat can integrate with Slack and/or Discord to bridge messages, and Twitter to post new-user announcements. 
See the [Admin's Manual](Admin's%20Manual.md) for more info.

Devzat has a plugin API you can use to integrate your own services: [documentation](plugin/README.md). Feel free to add a plugin to the main instance. Just ask for a token on the server.


## Stargazers over time

[![Stargazers over time](https://starchart.cc/quackduck/devzat.svg)](https://starchart.cc/quackduck/devzat)


## People

People who you might know who have joined:

Zach Latta - Founder of Hack Club: _"omg amazing! this is so awesome"_  
Ant Wilson - Co founder, Supabase: [_"brilliant!"_](https://twitter.com/AntWilson/status/1396444302721445889)  
Bereket [@heybereket](https://twitter.com/heybereket): _"this is pretty cool"_  
Ayush [@ayshptk](https://twitter.com/ayshptk): _"Can I double star the repo somehow :pleading_face:"_  
Sanketh [@SankethYS](https://twitter.com/SankethYS): _"Heck! How does this work. So cool."_  
Tony Dinh [@tdinh_me](https://twitter.com/tdinh_me): _"supeer cool, oh, open source as well? yeah"_  
Srushti [@srushtiuniverse](https://twitter.com/srushtiuniverse): _"Yess it's awesome. I tried it."_  
Surjith [@surjithctly](https://twitter.com/surjithctly): _"Whoa, who made this?"_  
Arav [@HeyArav](https://twitter.com/HeyArav): [_"Okay, this is actually super awesome."_](https://twitter.com/tregsthedev/status/1384180393893498880)  
Harsh [@harshb__](https://twitter.com/harshb__): _"im gonna come here everyday to chill when i get bored of studying lol, this is so cool"_
Krish [@krishnerkar_](https://twitter.com/krishnerkar_):  [_"SHIT! THIS IS SO DOPE"_](https://twitter.com/krishnerkar_/status/1384173042616573960)  
Amrit [@astro_shenava](https://twitter.com/astro_shenava): _"Super cool man"_  
Mudrank [@mudrankgupta](https://twitter.com/mudrankgupta): "🔥🚀🚀"

From Hack Club:  
**[Caleb Denio](https://calebden.io), [Safin Singh](https://safin.dev), [Eleeza](https://github.com/E-Lee-Za)   
[Jubril](https://github.com/s1ntaxe770r), [Sarthak Mohanty](https://sarthakmohanty.me)    
[Sam Poder](http://sampoder.com), [Rishi Kothari](http://rishi.cx)    
[Amogh Chaubey](https://amogh.sh), [Ella](https://ella.cx/), [Hugo Hu](https://github.com/Hugoyhu)
[Matthew Stanciu](https://matthewstanciu.me/), [Tanishq Soni](https://tanishqsoni.me)**

Huge thanks to the amazing [Caleb Denio](https://github.com/cjdenio) for lending me the original Devzat server 💖

### *Made by [Ishan Goel](https://twitter.com/usrbinishan/) with feature ideas from friends. Thanks to [Caleb Denio](https://twitter.com/CalebDenio) for lending his server!*
