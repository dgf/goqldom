# goqldom service

GraphQL based HTTP service for DOM selections.

Start the extracted program, the dynamic bounded URL should be opened immediately in your default browser.

Just paste one of the included `examples` into the playground
or use the type completion system and schema documentation to explore it by your own.

## Configure interface and port

To start the service on a specific interface and port, call it from the command line.

Example: Linux on default public interface

```shell
./goqldom -addr :8080
```

Example Windows on localhost

```shell
goqldom.exe -addr localhost:8080
```
