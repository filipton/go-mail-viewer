# CLI Mail Viewer
Simple cli mail viewer using IMAP connection

## Demo
> [!NOTE]  
> Ascii cast is glitching while recording golang tview apps
[![Demo](https://asciinema.org/a/XB47gzGUWL9ggD44f8PCwkR7f.svg)](https://asciinema.org/a/XB47gzGUWL9ggD44f8PCwkR7f)

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
