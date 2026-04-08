# linksaver

A command-line tool for saving and managing links with metadata.

## Features

- Save links with optional metadata (name, description, author, license)
- View saved links in a formatted display
- List links in a compact format
- Open all saved links in your default browser
- Store data in a local JSON configuration file

## Installation

1. Clone or download this repository
2. Ensure you have Go installed (version 1.24.4 or later)
3. Build the application:

```bash
go install github.com/shadowdara/linksaver@latest
```

## Usage

### Initialize Configuration

First, create a configuration file:

```bash
linksaver init
```

This will create a `.samengine/linksaver.json` file in your current directory.

### Add a Link

Add a new link with metadata:

```bash
linksaver add
```

You will be prompted to enter:
- Name (optional)
- Link URL
- Description
- Author (optional)
- License (optional)
- License Link (optional)

### View Links

Display all saved links with their metadata:

```bash
linksaver view
```

### List Links

Show a compact list of all links:

```bash
linksaver list
```

### Open All Links

Open all saved links in your default browser:

```bash
linksaver
```

### Help

Display available commands:

```bash
linksaver help
```

## Configuration

Links are stored in `.samengine/linksaver.json` in your current directory. The configuration includes:

- Project name
- Pretty printing setting
- Array of saved links with metadata

## Requirements

- Go 1.24.4 or later
- Operating system with a default browser (Windows, macOS, or Linux with xdg-open)

## License

This project is open source. See individual links for their respective licenses.

