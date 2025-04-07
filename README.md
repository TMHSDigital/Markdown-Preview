# Markdown-Preview

[![Go Version](https://img.shields.io/github/go-mod/go-version/TMHSDigital/Markdown-Preview)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Code Style](https://img.shields.io/badge/code%20style-gofmt-blue.svg)](https://pkg.go.dev/cmd/gofmt)
[![Made with Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](https://go.dev)
[![Build Size](https://img.shields.io/badge/build%20size-2%20MB-success)](https://github.com/TMHSDigital/Markdown-Preview)
[![Go Report Card](https://goreportcard.com/badge/github.com/TMHSDigital/Markdown-Preview)](https://goreportcard.com/report/github.com/TMHSDigital/Markdown-Preview)

<div align="center">
  <img src="static/images/logo.png" alt="Markdown-Preview Logo" width="200"/>
  <br/>
  <strong>A modern, feature-rich markdown editor and previewer built with Go and modern web technologies.</strong>
  <br/><br/>
  <a href="#features">Features</a> •
  <a href="#quick-start">Quick Start</a> •
  <a href="#keyboard-shortcuts">Shortcuts</a> •
  <a href="#architecture">Architecture</a>
</div>

<hr/>

## Overview

Markdown-Preview is a lightweight yet powerful markdown editor that combines real-time preview with modern features. Built with performance and usability in mind, it offers a seamless writing experience for developers, technical writers, and markdown enthusiasts.

<div align="center">
  <img src="static/images/screenshot.png" alt="Markdown-Preview Screenshot" width="800"/>
</div>

## Features

| Category | Features | Description |
|----------|----------|-------------|
| Editor | - Real-time preview<br>- Split view with resizable panes<br>- Syntax highlighting<br>- Auto-save functionality | Advanced editor with instant preview and modern IDE features |
| Appearance | - Dark/Light mode<br>- Clean, modern UI<br>- Mobile responsive design | Polished interface that adapts to any device or preference |
| Functionality | - Offline support<br>- Image upload & drag-n-drop<br>- Keyboard shortcuts | Works without internet and supports rich media content |
| Documentation | - Interactive markdown guide<br>- Tooltips and hints | Built-in learning resources and contextual help |

## Technical Details

### Performance Metrics

| Metric | Value | Description |
|--------|-------|-------------|
| Initial Load | ~2MB | Optimized bundle size |
| Preview Latency | ~50ms | Real-time markdown rendering |
| Memory Usage | <100MB | Efficient resource utilization |
| Browser Support | Modern Browsers | Chrome 80+, Firefox 75+, Safari 13+ |

### Dependencies

```go
require (
	github.com/gomarkdown/markdown v0.0.0-latest
	github.com/gorilla/mux v1.8.0
	github.com/tdewolff/minify/v2 v2.12.0
)
```

## Quick Start

### Prerequisites

```bash
go version >= 1.16
```

### One-Line Installation

```bash
git clone https://github.com/TMHSDigital/Markdown-Preview.git && cd Markdown-Preview && go run main.go
```

### Manual Installation

1. Clone the repository:
```bash
git clone https://github.com/TMHSDigital/Markdown-Preview.git
cd Markdown-Preview
```

2. Install dependencies:
```bash
go mod download
```

3. Run the application:
```bash
go run main.go
```

4. Open your browser:
```bash
http://localhost:8080
```

## Keyboard Shortcuts

| Category | Shortcut | Action | Context |
|----------|----------|--------|---------|
| **File Operations** | `Ctrl+S` | Save | Editor |
| **Text Formatting** | `Ctrl+B`<br>`Ctrl+I`<br>`Ctrl+K` | Bold<br>Italic<br>Code Block | Editor |
| **Insert Elements** | `Ctrl+L` | Link | Editor |
| **UI Controls** | `Ctrl+M`<br>`Ctrl+/` | Toggle Preview<br>Toggle Guide | Global |

## Architecture

```plaintext
Markdown-Preview/
├── main.go           # Application entry point and server setup
├── static/          # Static assets and client-side resources
│   ├── styles.css   # CSS styling and theme definitions
│   ├── scripts/     # JavaScript modules
│   └── images/      # Static images and icons
├── templates/       # Go HTML templates
│   └── index.html   # Main application template
├── internal/        # Internal packages
│   ├── markdown/    # Markdown processing
│   └── server/      # HTTP server logic
└── uploads/         # User uploaded content (gitignored)
```

## Security Features

| Feature | Implementation |
|---------|---------------|
| Input Sanitization | HTML sanitization and markdown safe rendering |
| Upload Security | File type validation and size limits |
| XSS Protection | Content Security Policy (CSP) headers |
| CSRF Protection | Token-based request validation |

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Tech Stack

| Category | Technology | Usage |
|----------|------------|--------|
| Backend | Go 1.16+ | Server runtime |
| Frontend | Vanilla JS | Client-side logic |
| Markdown | gomarkdown | Markdown processing |
| Syntax Highlighting | [Prism.js](https://prismjs.com/) | Code block highlighting |
| UI Components | [Material Icons](https://material.io/resources/icons/) | Interface icons |
| Templates | Go html/template | HTML rendering |

---

<div align="center">
  <sub>Powered by Go. Built by TMHSDigital.</sub>
</div>