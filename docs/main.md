This .md file focus on describe how the index.html should work

## Config

The `init` socket message receives a property called `interface_config` this interface receives the current JSON string

The `interface_config` data is sourced from the bot state key `"htmlinterface"`, typically set via the `[State]` section in `init.txt`:
```
[State]
State.Data.htmlinterface = {"ignoreKick":false,"ignoreYoutube":false,"ignoreTwitch":false}
```

All property can be null or undefined
```json
{
    ignoreKick: boolean,
    ignoreYoutube: boolean,
    ignoreTwitch: boolean,
}
```

### Ignore's 
Should not list as possibility for login and/or the panel