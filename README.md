# Gator RSS Aggregator

Gator is a command-line RSS feed aggregator built in Go. It lets you follow RSS feeds, aggregate posts from them into a PostgreSQL database, browse recent posts, and more — all from your terminal.

This project was built as part of the Boot.dev RSS.Aggregator in Go course. (https://www.boot.dev/courses/build-blog-aggregator-golang) 

## Features

- Register and manage users
- Add and follow RSS feeds
- Automatically aggregate posts in the background
- Browse recent posts from followed feeds
- Follow/unfollow feeds easily
- Simple configuration via a JSON file

## Prerequisites

To run Gator, you'll need:

- **Go** — version 1.22 or later (tested with 1.23+)
  - Download & install from https://go.dev/dl/
- **PostgreSQL** — version 13 or later
  - Install locally (e.g., via Homebrew on macOS: `brew install postgresql`, or use Docker, or any other method)
  - Make sure you have a running PostgreSQL server and know your connection details

## Installation

1. Make sure Go is installed and in your PATH (`go version` should work)

2. Install the `gator` CLI globally using `go install`:

   ```bash
   go install github.com/jonasyke/gator@latest