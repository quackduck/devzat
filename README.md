# Devzat

Where are the devs at? Devzat!

Devzat is chat over SSH  
![image](https://user-images.githubusercontent.com/38882631/115499526-a4d70280-a280-11eb-8723-817f54eccf3e.png)


## Usage

Try it out:

```sh
ssh devzat.hackclub.com
```

Add this to `~/.ssh/config`:
```ssh
Host chat
    HostName devzat.hackclub.com
```

Now you can join with just:
```sh
ssh chat
```

If you're under a firewall, you can still join on port 443:
```sh
ssh devzat.hackclub.com -p 443
```

We also have a Slack bridge! If you're on the [Hack Club](https://hackclub.com) Slack, check out the `#ssh-chat-bridge` channel!

```text
Available commands
   /users   list users
   /nick    change your name
   /color   change your name color
   /exit    leave the chat
   /hide    hide messages from HC Slack
   /bell    toggle the ansi bell
   /id      get a unique identifier for a user
   /all     get a list of all unique users ever
   /people  see info about nice people who joined
   /ban     ban a user, requires an admin pass
   /kick    kick a user, requires an admin pass
   /help    show this help message
Made by Ishan Goel with feature ideas from Hack Club members.
Thanks to Caleb Denio for lending his server!
```

## People

People who you might know who have joined:

Zach Latta - Founder of Hack Club: _"omg amazing! this is so awesome"_  
Bereket [@heybereket](https://twitter.com/heybereket): _"this is pretty cool"_  
Ayush [@ayshptk](https://twitter.com/ayshptk): _"Can I double star the repo somehow :pleading_face:"_  
Srushti [@srushtiuniverse](https://twitter.com/srushtiuniverse):  _"Yess It's awesome. I tried it."_   
Arav [@tregsthedev](https://twitter.com/tregsthedev):  https://twitter.com/tregsthedev/status/1384180393893498880  
Krish [@krishnerkar_](https://twitter.com/krishnerkar_):  https://twitter.com/krishnerkar_/status/1384173042616573960

From Hack Club:
Sam Poder, Caleb Denio, Safin Singh, Eleeza A, Jubril
Sarthak Mohanty, Anghe, Tommy Pujol
