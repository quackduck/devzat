<div align="center">
  <img src="https://github.com/CaenJones/Devzat-readme-update/blob/main/src/Welcome%20To%20@(4).png?raw=true" alt="Logo"> 
<a href="https://www.producthunt.com/posts/devzat?utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-devzat" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=298678&theme=light&period=daily" alt="Devzat - Chat with other devs over SSH in your Terminal | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
</div>

Devzat is a chatroom right into your terminal! You can join on any device via SSH or our Slack and Discord integrations. We are a community of programmers, hobyists, or people just looking to socialize.

 <h3>Join Devzat via SSH</h3>

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
We also have a Slack bridge! If you're on the [Hack Club](https://hackclub.com) Slack, check out the `#ssh-chat-bridge` channel!

Feel free to make a [new issue](https://github.com/quackduck/devzat/issues) if something doesn't work.

See the [status site](https://stats.uptimerobot.com/kxMQqfYk4y) of the main Devzat server to check if it might be down.

[![asciicast](https://asciinema.org/a/477083.svg)](https://asciinema.org/a/477083?speed=3)

<h3>Some Commands to get you started...</h3>
Devzat has lots of cool fetures, games, and some secrets... Here are some basic commands to get you up and running!

```text
   =<user>   <msg>           DM <user> with <msg>
   color     <color>         Change your name's color
   exit                      Leave the chat
   man       <cmd>           Get help for a specific command
   bell      on|off|all      ANSI bell on pings (on), never (off) or for every message (all)
   clear                     Clear the screen
   cd        #room|user      Join #room, DM user or run cd to see a list
   tz        <zone> [24h]    Set your IANA timezone (like tz Asia/Dubai) and optionally set 24h
   nick      <name>          Change your username
   pwd                     Show your current room
   id       <user>         Get a unique ID for a user (hashed key)
'''
Type CMDS in the chat to get the full list.

<h3>Self-host your own Devzat instance</h3>

If you want to self host your own instance, you can grab basicly any LINUX/UNIX device with an internet connection and golang installed, and paste these commands in:
```shell
git clone https://github.com/quackduck/devzat && cd devzat
go install # or build, if you want to keep things pwd
ssh-keygen -qN '' -f devzat-sshkey # new ssh host key for the server
devzat # run! the default config is used & written automatically
```
These commands download, build, setup and run a Devzat server listening on port 2221, the default port (change by setting `$PORT`).
If you have trouble connecting to your devzat instance, you might be a pubkey issue so try to connect through port 443.

When self-hosting, Devzat can integrate with Slack and/or Discord to bridge messages, and Twitter to post new-user announcements. 
See the [Admin's Manual](Admin's%20Manual.md) for more info and configuration documentation.

Devzat has a plugin API you can use to integrate your own services: [documentation](plugin/README.md). Feel free to add a plugin to the main instance. Just ask for a token on the server.

<h3>We are still growing!</h3>

Devzat has a vibrant community, and it is still growing! Join today to make some friends, or start your own server today!
[![Stargazers over time](https://starchart.cc/quackduck/devzat.svg)](https://starchart.cc/quackduck/devzat)


<h3>People who you might know who have joined</h3>

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

### *Made by [Ishan Goel](https://twitter.com/usrbinishan/) with feature ideas and contributions from friends.
</div>
