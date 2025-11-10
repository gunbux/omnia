# 🚀 Omnia

**Everything at your fingertips** — Omnia is a blazingly fast, minimalist application launcher designed to be your universal gateway to everything on your system. Whether you're launching desktop applications, running CLI commands, or finding that one tool you installed months ago, Omnia puts it all within reach through an elegant terminal interface.

Built with the modern [Bubble Tea](https://github.com/charmbracelet/bubbletea) TUI framework, Omnia combines the power of fuzzy search with the speed of native Go to deliver an incredibly responsive launching experience.

## ✨ Features

- **🔍 Fuzzy Search**: Find what you need with intelligent, typo-tolerant search
- **🖥️ Desktop Applications**: Launch any installed GUI application
- **⚡ CLI Commands**: Execute terminal commands and utilities
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

### Navigation
- **Type**: Start typing to search and filter results
- **Tab**: Switch between input and results
- **Arrow Keys**: Navigate through completions
- **Enter**: Launch the selected item
- **Esc**: Exit Omnia

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
