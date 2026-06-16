---
title: "Quick start"
description: "Fetch your first record with wandb."
weight: 30
---

Once `wandb` is on your `PATH`, fetch a page. The argument is the path
of the page on wandb.com (everything after the host), or a full URL:

```bash
wandb page <path>
```

By default you get an aligned table. Ask for JSON when you want to pipe it:

```bash
$ wandb page <path> -o json
[
  {
    "id": "<path>",
    "url": "https://wandb.com/<path>",
    "title": "<path>",
    "body": "..."
  }
]
```

## Shape the output

The same flags work on every command:

```bash
wandb page <path> --fields id,url        # keep only these columns
wandb page <path> --template '{{.Body}}' # just the body text
wandb page <path> -o jsonl | jq .url     # one object per line, into jq
```

`-o` takes `table`, `json`, `jsonl`, `csv`, `tsv`, `url`, or `raw`. Left to
`auto`, it prints a table to a terminal and JSONL into a pipe, so the same
command reads well by hand and parses cleanly downstream. See
[output formats](/reference/output/) for the full contract.

## Follow the links

`links` lists the pages a page links to, and each one is a path you can fetch in
turn:

```bash
wandb links <path> -n 10                 # the first ten links
wandb links <path> -o url                # just the URLs
wandb links <path> -o url | head -3 | xargs -n1 wandb page
```

## Serve it instead

The same operations are available over HTTP and to agents over MCP:

```bash
wandb serve --addr :7777 &
curl -s 'localhost:7777/v1/page/<path>'          # NDJSON, one record per line
wandb mcp                                # MCP over stdio: page, links
```

## What to build next

This scaffold ships one example type, `page`, wired end to end so the whole
chain works today. To make it really about wandb, model the records you
care about in `wandb/` and declare their operations in
`wandb/domain.go`. Each one you add shows up as a command here, a route
under `serve`, and a tool under `mcp`, with no extra wiring. The
[guides](/guides/) cover the common jobs.
