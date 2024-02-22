# CLI Mail Viewer
Simple cli mail viewer using IMAP connection

## Setup
- Create .env file in project dir
```
IMAP_SERVER=host:port
IMAP_USERNAME=email
IMAP_PASSWORD=pass
```

- Run
```bash
$ go run main.go
```

## Features
- Simple message preview inside TUI
- HTML message view after clicking enter on message (uses xdg-open to open .html file from temp)

## TODO:
- [ ] Code cleanup
- [ ] Building crossplatform binary
