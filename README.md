# slack-files

If you don't have budget to pay for a slack account you probably will need starting to delete files once you get the storage out of space. This CLI is just a simple solution to help find out the largest files in your account.

### Download

[Releases](https://github.com/maxcnunes/slack-files/releases)

### Options

* **--query**: Search query. May contain booleans, etc.
* **--token**: Slack Authentication token (Requires scope: search:read).

## Development

```bash
go get -v ./...

go run main.go --token <TOKEN>
```

## Build

Using [goxc](https://github.com/laher/goxc).

```bash
goxc
```
