This .md file focus on describe how the index.html should work

## Config

The `init` socket message receives a property called `interface_config` this interface receives the current JSON string

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