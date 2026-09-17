# 🚀 Omnia

**Everything at your fingertips** — Omnia is a blazingly fast, minimalist application launcher designed to be your universal gateway to everything on your system. Whether you're launching desktop applications, running CLI commands, or finding that one tool you installed months ago, Omnia puts it all within reach through an elegant terminal interface.

Built with the modern [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework, Omnia combines the power of fuzzy search with the speed of native Go to deliver an incredibly responsive launching experience.

## ✨ Features

- **🔍 Fuzzy Search**: Find what you need with intelligent, typo-tolerant search
- **🖥️ Desktop Applications**: Launch any installed GUI application
- **⚡ Shell Commands**: `>` runs any command in a new terminal, with completion
- **📁 File Finder**: `/`, `~` or `f ` searches your files and completes paths
- **🧮 Calculator**: type `2^10 + 0xff` or `sqrt(2)*pi` and get the answer inline
- **💱 Currency Converter**: `100 usd to sgd`, `€50`, `usd to jpy` with live rates
- **🤖 Quick AI Answers**: `? how do I untar a .tar.xz` streams an answer from Claude
- **⌨️ Keyboard-First**: Navigate entirely with keyboard shortcuts
- **🎨 Clean Interface**: Minimal, distraction-free design
- **🚀 Lightning Fast**: Built in Go for maximum performance

## 🛠 Installation

### Build from Source

```bash
git clone https://github.com/gunbux/omnia.git
cd omnia
go build -o omnia .
```

## 🎯 Usage

Launch Omnia from your terminal:

```bash
./omnia
```

### Modes

Omnia picks a mode from what you type. The chip on the left of the input shows the active mode.

| Type | Mode | What happens |
|------|------|--------------|
| `firefox` | **Apps** | Fuzzy-searches installed applications. Enter launches the selection, or runs the text as a command if nothing matches. |
| `> htop` | **Shell** | Completes command names. Enter runs the command in a new terminal (`kitty`). |
| `/etc/ho`, `~/Doc`, `f report` | **Files** | `/` and `~` complete paths like a shell; `f ` searches file names under your home directory (uses `fd` when installed). |
| `2^10 + 0xff`, `= 5!` | **Calc** | Arithmetic with `+ - * / ^ % !`, functions (`sqrt`, `sin`, `log`, `min`, …), constants (`pi`, `e`), hex/binary literals and `k`/`m` suffixes. `=` forces calc mode. |
| `100 usd to sgd`, `$50`, `€1,250 in gbp` | **Currency** | Converts using live rates from open.er-api.com (cached for 12 hours, works offline with the last rates). Without a target it shows a few major currencies. |
| `? how do I untar a .tar.xz`, `ai …` | **Ask AI** | Streams a short answer from Claude. Enter copies the answer and closes; Esc goes back. |

### Keys

| Key | Action |
|-----|--------|
| **Enter** | Launch / run / open / copy, depending on the mode |
| **↑ ↓** / **Tab** / **Shift+Tab** | Move through results |
| **→** (at end of input) | Complete the selected path or command into the input |
| **Ctrl+Y** | Copy the selected result (file path, number, amount, answer) |
| **Ctrl+T** | Files mode: open a terminal in the selected directory |
| **Esc** | Quit (or go back from an AI answer) |

### AI backend

The AI mode needs one of:

- the [Claude Code](https://claude.com/claude-code) CLI on your `PATH`, which reuses your existing login, or
- `ANTHROPIC_API_KEY` (or `ANTHROPIC_AUTH_TOKEN`) set, which uses the Anthropic API directly.

Optional overrides: `OMNIA_AI_BACKEND=api|claude` to force a backend and `OMNIA_AI_MODEL` to pick a model (default `claude-opus-5`).

## 🔧 Hyprland Integration

For the ultimate workflow integration, add these configurations to your Hyprland setup:

### Keybinding
```ini
# ~/.config/hypr/hyprland.conf
bind = $mainMod, SPACE, exec, kitty -T omnia /home/chun/repo/omnia/omnia
```

### Window Rules
```ini
# ~/.config/hypr/hyprland.conf
windowrulev2 = float, title:^(omnia)$
windowrulev2 = size 50% 50%, title:^(omnia)$
windowrulev2 = center, title:^(omnia)$
```

This setup gives you:
- **Super + Space**: Instantly summon Omnia
- **Floating window**: Appears over your current workspace
- **Perfect sizing**: Takes up 50% of your screen, centered
- **Seamless integration**: Feels like a native launcher

## 🏗 Development

### Build Commands
```bash
# Build the application
go build -o omnia .

# Run in development
go run main.go

# Run tests
go test ./...

# Format code
go fmt ./...

# Lint (requires golangci-lint)
golangci-lint run
```

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

## 📄 License

This project is licensed under the MIT License.
