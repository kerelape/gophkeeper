# Gophkeeper

## Server

```bash
$ go build -o gophserver ./cmd/server
$ ./gophserver
```
Will output help information about server run configuration:
```
Environment variables:
  DATABASE_DSN string
        Database connection URL
  PASSWORD_MIN_LENGTH uint
        Password minimum length (default "0")
  REST_ADDRESS string
        Address that REST api listens on. (default ":16355")
  REST_HOST_WHITELIST slice
         (default "")
  REST_USE_TLS bool
        Use TLS or not (default "true")
  TOKEN_LIFESPAN int64
        JWT Token lifespan in milliseconds (default "15m")
  TOKEN_SECRET string
        Base64 encoded JWT Token secret
  USERNAME_MIN_LENGTH uint
        Username minimum length (default "0")
exit status 1
```

## CLI

```bash
$ go build -o gophkeeper ./cmd/cli
$ ./gophkeeper -s "https://localhost:16355" help
```

__Where `https://localhost:16355` is the address that the server listens on__

Note that, even though the client does not need to connect to the
server for a `help`, it will still require a value set to the `-s` flag,
thus the value can be any valid string if you only want to see the help.
```bash
$ ./gophkeeper -s none help
```
