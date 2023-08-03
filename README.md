<div align="center">
  <img src="https://github.com/CaenJones/Devzat-readme-update/blob/main/src/Welcome%20To%20@(4).png?raw=true" alt="Logo">
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

<h3>Self-host your own Devzat instance</h3>

If you want to self host your own instance, you can grab basicly any LINUX/UNIX device with an internet connection and golang installed, and paste these commands in:
```shell
git clone https://github.com/quackduck/devzat && cd devzat
go install # or build, if you want to keep things pwd
ssh-keygen -qN '' -f devzat-sshkey # new ssh host key for the server
devzat # run! the default config is used & written automatically
```
These commands download, build, setup and run a Devzat server listening on port 2221, the default port (change by setting `$PORT`).
If you have trouble connecting to your devzat instance, you might be a pubkey issue so try to connect through port 443

When self-hosting, Devzat can integrate with Slack and/or Discord to bridge messages, and Twitter to post new-user announcements. 
See the [Admin's Manual](Admin's%20Manual.md) for more info and configuration documentation.

Devzat has a plugin API you can use to integrate your own services: [documentation](plugin/README.md). Feel free to add a plugin to the main instance. Just ask for a token on the server.

<h3>We are still growing!</h3>

Devzat has a vibrant community, and it is still growing! Join today to make some friends, or start your own server today!
[![Stargazers over time](https://starchart.cc/quackduck/devzat.svg)](https://starchart.cc/quackduck/devzat)


<h3>People who you might know who have joined</h3>

Zach Latta - Founder of Hack Club: "omg amazing! this is so awesome"
Ant Wilson - Co founder, Supabase: "brilliant!" [Twitter](https://twitter.com/AntWilson/status/1396444302721445889)
Bereket [@heybereket](https://twitter.com/heybereket): "this is pretty cool"
Ayush [@ayshptk](https://twitter.com/ayshptk): "Can I double star the repo somehow :pleading_face:"
Sanketh [@SankethYS](https://twitter.com/SankethYS): "Heck! How does this work. So cool."
Tony Dinh [@tdinh_me](https://twitter.com/tdinh_me): "supeer cool, oh, open source as well? yeah"
Srushti [@srushtiuniverse](https://twitter.com/srushtiuniverse): "Yess it's awesome. I tried it."
Surjith [@surjithctly](https://twitter.com/surjithctly): "Whoa, who made this?"
Arav [@HeyArav](https://twitter.com/HeyArav): "Okay, this is actually super awesome." [Tweet](https://twitter.com/tregsthedev/status/1384180393893498880)
Harsh [@harshb__](https://twitter.com/harshb__): "im gonna come here everyday to chill when i get bored of studying lol, this is so cool"
Krish [@krishnerkar_](https://twitter.com/krishnerkar_): "SHIT! THIS IS SO DOPE" [Tweet](https://twitter.com/krishnerkar_/status/1384173042616573960)
Amrit [@astro_shenava](https://twitter.com/astro_shenava): "Super cool man"
Mudrank [@mudrankgupta](https://twitter.com/mudrankgupta): "🔥🚀🚀"

<h3>From Hack Club:</h3>

- [Caleb Denio](https://calebden.io)
- [Safin Singh](https://safin.dev)
- [Eleeza](https://github.com/E-Lee-Za)
- [Jubril](https://github.com/s1ntaxe770r)
- [Sarthak Mohanty](https://sarthakmohanty.me)
- [Sam Poder](http://sampoder.com)
- [Rishi Kothari](http://rishi.cx)
- [Amogh Chaubey](https://amogh.sh)
- [Ella](https://ella.cx/)
- [Hugo Hu](https://github.com/Hugoyhu)
- [Matthew Stanciu](https://matthewstanciu.me/)
- [Tanishq Soni](https://tanishqsoni.me)


<div align="center">
<a href="https://www.producthunt.com/posts/devzat?utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-devzat" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=298678&theme=light&period=daily" alt="Devzat - Chat with other devs over SSH in your Terminal | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>
  
### *Made by [Ishan Goel](https://twitter.com/usrbinishan/) with feature ideas and contributions from friends.
</div>
