const example = w.subscribe("customevent_example.lua");
if (example) {
    example.on("tick", (data) => console.log("Tick:", data));
    example.send("ping", {}, (data) => console.log("Pong:", data));
}
