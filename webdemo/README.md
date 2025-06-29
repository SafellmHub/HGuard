# HallucinationGuard Web Demo

This directory contains a web server and UI demo for HallucinationGuard agents. Use this to interactively test tool call validation and execution via a browser or API.

## Running the Web Server

```sh
go run webdemo/main.go
```

The server will start at [http://localhost:8080](http://localhost:8080).

## Testing the API

Send a POST request to `/prompt`:

```sh
curl -X POST -H "Content-Type: application/json" -d '{"prompt":"What is the weather in London?"}' http://localhost:8080/prompt
```

## More Examples

See [`../scaffold/`](../scaffold/README.md) for agent logic, tool schemas, and configuration details.

## License

MIT
