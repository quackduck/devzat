import Devzat from "../dist";

const plugin = new Devzat({
    address: "localhost:5556",
    name: "Demo bot",
    token: process.env.DEVZAT_TOKEN
});

plugin.onMessageSend({
    middleware: true,
    once: false
}, message => {
    console.log("new message!", message);

    if(!message.msg.startsWith("demo-bot ")) {
        return message.msg + " TESTING";
    }
});

plugin.onMessageSend({
    middleware: false,
    once: true
}, message => {
    console.log("got a message once", message);
});

plugin.command({
    name: "demo-bot",
    argsInfo: "<msg | \"send-test\">",
    info: "Ping the demo bot"
}, async invocation => {
    console.log("got a command", invocation);
    if(invocation.args === "send-test") {
        setInterval(() => plugin.sendMessage({
            room: "#main",
            msg: "Hello world!"
        }), 4000);
        return "Set interval!";
    } else if(invocation.args === "ephemeral-test") {
        await plugin.sendMessage({
            room: invocation.room,
            msg: "Pong!",
            ephemeralTo: invocation.from
        });
    } else return `Hello, ${invocation.from}! You said: ${invocation.args}`;
})