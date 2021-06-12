# Devzat

<a href="https://www.producthunt.com/posts/devzat?utm_source=badge-top-post-badge&utm_medium=badge&utm_souce=badge-devzat" target="_blank"><img src="https://api.producthunt.com/widgets/embed-image/v1/top-post-badge.svg?post_id=298678&theme=light&period=daily" alt="Devzat - Chat with other devs over SSH in your Terminal | Product Hunt" style="width: 250px; height: 54px;" width="250" height="54" /></a>

Where are the devs at? Devzat!

Devzat is chat over SSH. Because there's SSH apps literally on all platforms, even your phone, you can connect to Devzat on any device!

![image](https://user-images.githubusercontent.com/38882631/115499526-a4d70280-a280-11eb-8723-817f54eccf3e.png)


## Usage

Try it out:

```sh
ssh devzat.hackclub.com
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

### Help

```text
Welcome to Devzat! Devzat is chat over SSH: github.com/quackduck/devzat
Because there's SSH apps on all platforms, even on mobile, you can join from anywhere.

Interesting features:
• Many, many commands. Run /commands.
• Rooms! Run /room to see all rooms and use /room #foo to join a new room.
• Markdown support! Tables, headers, italics and everything. Just use \n in place of newlines.
• Code syntax highlighting. Use Markdown fences to send code. Run /example-code to see an example.
• Direct messages! Send a DM using =user <msg>.
• Timezone support, use /tz Continent/City to set your timezone.
• Built in Tic Tac Toe and Hangman! Run /tic or /hang <word> to start new games.
• Emoji replacements! :rocket: => 🚀  (like on Slack and Discord)

For replacing newlines, I often use bulkseotools.com/add-remove-line-breaks.php.

Made by Ishan Goel with feature ideas from friends.
Thanks to Caleb Denio for lending his server!
```
### Commands
```text
Available commands
   =<user> <msg>           DM <user> with <msg>
   /users                  List users
   /nick   <name>          Change your name
   /room   #<room>         Join a room or use /room to see all rooms
   /tic    <cell num>      Play Tic Tac Toe!
   /hang   <char/word>     Play Hangman!
   /people                 See info about nice people who joined
   /tz     <zone>          Change IANA timezone (eg: /tz Asia/Dubai)
   /color  <color>         Change your name's color
   /all                    Get a list of all users ever
   /emojis                 See a list of emojis
   /exit                   Leave the chat
   /help                   Show help
   /commands               Show this message
   /commands-rest          Uncommon commands list
```
```
/commands-rest:
   /hide                   Hide messages from HC Slack
   /bell                   Toggle the ANSI bell used in pings
   /id     <user>          Get a unique ID for a user (hashed IP)
   /ban    <user>          Ban <user> (admin)
   /kick   <user>          Kick <user> (admin)
   /ascii-art              Show some panda art
   /shrug                  ¯\_(ツ)/¯_
   /example-code           Example syntax-highlighted code
   /banIP  <IP/ID>         Ban by IP or ID (admin)
```

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
[Tommy Pujol](https://itstommy.xyz/), [Sam Poder](http://sampoder.com), [Rishi Kothari](http://rishi.cx)    
[Amogh Chaubey](https://amogh.sh), [Ella Xu](https://timeline.ella.cx/), [Hugo Hu](https://github.com/Hugoyhu)
[Matthew Stanciu](https://matthewstanciu.me/), [Tanishq Soni](https://tanishqsoni.me)**

Huge thanks to the amazing [Caleb Denio](https://github.com/cjdenio) for lending me the original Devzat server 💖

### *Made by [Ishan Goel](https://twitter.com/IshanTheIshan) with feature ideas from friends. Thanks to [Caleb Denio](https://twitter.com/CalebDenio) for lending his server!*
