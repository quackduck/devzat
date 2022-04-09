import Devzat from "../dist";

const plugin = new Devzat("localhost:5556");

// setInterval(() => plugin.sendMessage({
//     from: "Test bot",
//     room: "#main",
//     msg: "Hello world!"
// }), 4000);

plugin.onMessageSend({
    middleware: true,
    once: false
}, message => {
    console.log("new message!", message.msg);

    return message.msg + " TESTING";
})