# MyStreamBot
---

MyStreamBot is a lightweight application that can connect to multiple streaming platforms. Allowing the streamer to monitor and control all their streams in one place.

The current capabilities are: send messages, view and manage (to an extent) the chat and events of different streams in the same window. It also has the module features, which allows custom commands creation. 


## Installation
---

First compile and then build the application by running the following commands:
```bash
go build -ldflags="-X 'main.Version=1.0.0' -X 'main.BuildDate=$(date)' -X 'main.CommitHash=$(git rev-parse HEAD)'" -o build/mystreambot.exe .
go run .
```

## Executing
---
Once you have the `.exe` file inside the `build` folder, run it and then open a tab in your browser with the address `localhost:1699`. Log in the desired streaming platforms and you're all set to enjoy MyStreamBot! 😃

This address can also be accessed through your browser smartphone as long as it is in the same connection of your computer.

**Note:** If you ever encounter changes that aren't being reflected on the browser tab, press `CTRL + F5` to erase all cache and reload the page. Doing this may fix the issue.

## Modules

<<<<<<< HEAD
The modules is written in `.lua` and is how you add more functionalities to the bot

[Detailed module creation](docs/MODULE.md)
=======
[Detailed module creation](docs/Module.md)
>>>>>>> df386cb0b32b9dc36b11a81036a67914f3e1b8b7
