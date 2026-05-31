# point-tui

A feature-rich TUI client for the Point microblogging platform, built with Go and Bubble Tea.

## Example
![Example](docs/example.png)

## Features

- **Three-pane Miller column layout:** Navigate through tags, posts, and post previews seamlessly.
- **Timeline View:** Browse posts by year.
- **Tag-based Navigation:** Filter posts by tags and explore tag hierarchies.
- **Rich Media Support:** Image and video previewing (via ANSI art/terminal rendering).
- **Interactive TUI:** Built with Bubble Tea, featuring mouse support, keyboard shortcuts, and responsive layout.
- **Secure Access:** Optional login support for private content.
- **Post Details:** View full content, excerpts, and metadata of posts.
- **Cached Media:** Local caching for images and media for faster loading.

## Installation

### Prerequisites

- Go 1.21 or later

### Building from Source

```bash
git clone https://github.com/dariy/point-tui.git
cd point-tui
make build
```

The binary will be available as `point-tui` in the project root.

## Usage

Run the application:

```bash
./point-tui [url]
```

### Options

- `-l`, `--login`: Prompt for password to access private content.

### Examples

```bash
./point-tui                  # Use default instance (https://darii.net)
./point-tui darii.net        # Use specific instance
./point-tui -l darii.net     # Use specific instance and login
```
## Keybindings

| Key | Action |
| --- | --- |
| `k`/`↑` | Move up |
| `j`/`↓` | Move down |
| `h`/`←` | Previous pane |
| `l`/`→` | Next pane |
| `tab` | Cycle pane forward |
| `shift+tab` | Cycle pane backward |
| `1`-`5` | Jump to specific pane (Timeline, Tags, Posts, Preview, Messages) |
| `enter` | Open/Select |
| `/` | Search |
| `r` | Reload |
| `f` | Toggle fullscreen |
| `p` | Play/Pause video |
| `o`/`u` | Open in browser |
| `q`/`ctrl+c` | Quit |

