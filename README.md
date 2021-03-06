# slack-files

**Control over the Slack files clean up**

If you don't have budget to pay for a slack account you probably will need starting to delete files once you get the storage out of space. This CLI is just a simple solution to help find out the largest files in your account.

![](https://raw.githubusercontent.com/maxcnunes/slack-files/master/slack-files.png)

**Important!**

This CLI uses 2 different APIs to fetch the files from Slack.
On using the args:

* `--query`: It will fetch the files from https://api.slack.com/methods/search.files
* `--types` and `--days-to`: Will fetch from https://api.slack.com/methods/files.list

Although on using multiple args from different APIs it will gather the data. It will not apply the criteria from one API to another.
For instance on filtering by query and days-to. The data from search.files API will not be filtered by the days.
This is a limitation I will keep this way at least initially to simplify the implementation of the CLI.

But no worries. The CLI will ask you to confirm before deleting any file. Also is possible to backup the files.

### Download

[Releases](https://github.com/maxcnunes/slack-files/releases)


### Args

* **--query [string]**: Search query. Accept multiple values separated by `,`.
* **--token [string]**: Slack Authentication token.
* **--types [string]**: Filter files by type. Accept multiple values separated by `,`.
* **--days-to [int]**: Filter files created before this timestamp (inclusive).
* **--backup [string]**: Path to backup files before delete.

**--types**

Valid types specified by Slack API https://api.slack.com/methods/files.list

* `all` - All files
* `spaces` - Posts
* `snippets` - Snippets
* `images` - Image files
* `gdocs` - Google docs
* `zips` - Zip files
* `pdfs` - PDF files

#### Examples

**Filtering with query by multiple extensions**

```bash
--query ".rar,.tar,.zip,.mp3,.mp4,.pdf,.ppt,.csv,.jpeg,.json"
```

**Filtering by files of all image types**

```bash
--types images
```

**Filtering by all files older than 30 days**

```bash
--days-to 30
```

**Filtering by files of all image types older than 30 days**

```bash
--types images --days-to 30
```

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
