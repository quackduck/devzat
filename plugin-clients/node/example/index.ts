import Devzat from "../dist";

const plugin = new Devzat({
    address: "localhost:5556",
    name: "Demo bot"
});

plugin.onMessageSend({
    middleware: true,
    once: false
}, message => {
    console.log("new message!", message.msg);

    if(!message.msg.startsWith("demo-bot ")) {
        return message.msg + " TESTING";
    }
});

plugin.onMessageSend({
    middleware: false,
    once: true
}, message => {
    console.log("got a message once", message.msg);
});

plugin.command({
    name: "demo-bot",
    argsInfo: "<msg | \"send-test\">",
    info: "Ping the demo bot"
}, invocation => {
    if(invocation.args === "send-test") {
        setInterval(() => plugin.sendMessage({
            room: "#main",
            msg: "Hello world!"
        }), 4000);
        return "Set interval!";
    }
    return `Hello, ${invocation.from}! You said: ${invocation.args}`;
})