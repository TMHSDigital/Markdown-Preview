package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/gomarkdown/markdown"
	"github.com/gorilla/websocket"
)

var (
	port         = flag.String("port", "8080", "HTTP server port")
	host         = flag.String("host", "localhost", "Host to bind to")
	maxPort      = flag.Int("max-port", 8180, "Maximum port to try")
	noOpen       = flag.Bool("no-open", false, "Don't open browser automatically")
	markdownFile = flag.String("file", "content.md", "Markdown file to preview")
	uploadDir    = flag.String("upload-dir", "uploads", "Directory for uploaded images")
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

func init() {
	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(*uploadDir, 0755); err != nil {
		log.Fatal(err)
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}

func findAvailablePort(startPort string, hostname string) (string, string) {
	// Remove ":" prefix if present
	cleanPort := strings.TrimPrefix(startPort, ":")
	portNum, err := strconv.Atoi(cleanPort)
	if err != nil {
		portNum = 8080 // Default if port is invalid
	}

	for portNum <= *maxPort {
		addr := fmt.Sprintf("%s:%d", hostname, portNum)
		listener, err := net.Listen("tcp", addr)
		if err == nil {
			listener.Close()
			return addr, fmt.Sprintf("http://%s:%d", hostname, portNum)
		}
		portNum++
	}
	log.Fatalf("No available ports found between %s and %d", startPort, *maxPort)
	return "", ""
}

func handleImageUpload(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Handle file upload
	file, header, err := r.FormFile("image")
	if err != nil {
		http.Error(w, "Invalid file upload", http.StatusBadRequest)
		return
	}
	defer file.Close()

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	filename := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)
	filepath := filepath.Join(*uploadDir, filename)

	// Create the file
	dst, err := os.Create(filepath)
	if err != nil {
		http.Error(w, "Failed to create file", http.StatusInternalServerError)
		return
	}
	defer dst.Close()

	// Copy the uploaded file
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return
	}

	// Return the markdown image syntax
	imageURL := fmt.Sprintf("/uploads/%s", filename)
	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"markdown": "![%s](%s)"}`, header.Filename, imageURL)
}

func main() {
	flag.Parse()

	// Find available port
	addr, url := findAvailablePort(*port, *host)
	log.Printf("Starting server at \u001b[36m%s\u001b[0m", url)
	log.Printf("Watching file: %s", *markdownFile)

	// Create HTTP server
	http.HandleFunc("/", handlePreview)
	http.HandleFunc("/ws", handleWebSocket)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	http.Handle("/uploads/", http.StripPrefix("/uploads/", http.FileServer(http.Dir(*uploadDir))))
	http.HandleFunc("/convert", handleMarkdownConvert)
	http.HandleFunc("/upload", handleImageUpload)
	http.HandleFunc("/guide", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/guide.html")
	})

	// Open browser automatically unless disabled
	if !*noOpen {
		go func() {
			if err := openBrowser(url); err != nil {
				log.Printf("Failed to open browser: %v", err)
			}
		}()
	}

	if err := http.ListenAndServe(addr, nil); err != nil {
		log.Fatal(err)
	}
}

func handleMarkdownConvert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	content := []byte(r.FormValue("markdown"))
	log.Printf("Converting markdown content (length: %d)", len(content))

	// Convert markdown to HTML
	html := markdown.ToHTML(content, nil, nil)
	log.Printf("Generated HTML (length: %d)", len(html))

	// Wrap the HTML in a div for proper rendering
	wrappedHTML := fmt.Sprintf(`<div class="markdown-body">%s</div>`, html)

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(wrappedHTML))
	log.Printf("Sent response with wrapped HTML (length: %d)", len(wrappedHTML))
}

func handlePreview(w http.ResponseWriter, r *http.Request) {
	const htmlStart = `<!DOCTYPE html>
<html>
<head>
	<title>Markdown Preview</title>
	<link rel="stylesheet" href="/static/styles.css">
	<link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/themes/prism-tomorrow.min.css">
	<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/bootstrap-icons@1.7.2/font/bootstrap-icons.css">
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/prism.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/components/prism-go.min.js"></script>
	<script src="https://cdnjs.cloudflare.com/ajax/libs/prism/1.24.1/components/prism-markdown.min.js"></script>
	<script>`

	const jsCode = `
		let lastSavedContent = '';
		let isOnline = true;
		let searchVisible = false;
		let guideVisible = false;

		function updatePreview() {
			const markdown = document.getElementById('editor').value;
			const preview = document.getElementById('preview');
			const formData = new FormData();
			formData.append('markdown', markdown);
			
			preview.innerHTML = '<div class="loading">Converting...</div>';
			
			// Add a small delay for better UX
			setTimeout(() => {
				fetch('/convert', {
					method: 'POST',
					body: formData
				})
				.then(response => {
					if (!response.ok) {
						throw new Error('Network response was not ok');
					}
					return response.text();
				})
				.then(html => {
					// Fade out current content
					preview.style.opacity = '0';
					setTimeout(() => {
						preview.innerHTML = html;
						preview.style.opacity = '1';
						document.querySelectorAll('pre code').forEach((block) => {
							Prism.highlightElement(block);
						});
						updateStatus('success', 'Preview updated');
						updateWordCount();
						
						if (markdown !== lastSavedContent) {
							saveToLocalStorage(markdown);
							lastSavedContent = markdown;
						}
					}, 150);
				})
				.catch(error => {
					preview.innerHTML = '<div class="error">Failed to update preview</div>';
					updateStatus('error', 'Failed to update preview: ' + error.message);
					console.error('Preview error:', error);
				});
			}, 150);
		}

		function updateWordCount() {
			const text = document.getElementById('editor').value;
			const words = text.trim().split(/\s+/).filter(word => word.length > 0).length;
			const chars = text.length;
			const lines = text.split('\n').length;
			document.getElementById('word-count').textContent = 
				words + ' words, ' + chars + ' characters, ' + lines + ' lines';
		}

		function insertMarkdown(type) {
			const editor = document.getElementById('editor');
			const start = editor.selectionStart;
			const end = editor.selectionEnd;
			const text = editor.value;
			let insertion = '';
			let cursorOffset = 0;

			switch(type) {
				case 'bold':
					insertion = '**' + (start === end ? 'bold text' : text.substring(start, end)) + '**';
					cursorOffset = start === end ? 2 : 0;
					break;
				case 'italic':
					insertion = '_' + (start === end ? 'italic text' : text.substring(start, end)) + '_';
					cursorOffset = start === end ? 1 : 0;
					break;
				case 'code':
					insertion = '\x60' + (start === end ? 'code' : text.substring(start, end)) + '\x60';
					cursorOffset = start === end ? 1 : 0;
					break;
				case 'link':
					insertion = '[' + (start === end ? 'link text' : text.substring(start, end)) + '](url)';
					cursorOffset = start === end ? 1 : text.substring(start, end).length + 3;
					break;
				case 'list':
					insertion = '- ' + (start === end ? 'list item' : text.substring(start, end));
					break;
				case 'quote':
					insertion = '> ' + (start === end ? 'quote' : text.substring(start, end));
					break;
				case 'h1':
					insertion = '# ' + (start === end ? 'heading' : text.substring(start, end));
					break;
				case 'h2':
					insertion = '## ' + (start === end ? 'heading' : text.substring(start, end));
					break;
				case 'h3':
					insertion = '### ' + (start === end ? 'heading' : text.substring(start, end));
					break;
			}

			editor.value = text.substring(0, start) + insertion + text.substring(end);
			editor.focus();
			if (start === end) {
				editor.selectionStart = start + cursorOffset;
				editor.selectionEnd = editor.selectionStart + 
					(insertion.length - (type === 'link' ? 7 : 2));
			} else {
				editor.selectionStart = start;
				editor.selectionEnd = start + insertion.length;
			}
			
			updatePreview();
		}

		function toggleSearch() {
			searchVisible = !searchVisible;
			const searchBar = document.getElementById('search-bar');
			searchBar.style.display = searchVisible ? 'flex' : 'none';
			if (searchVisible) {
				document.getElementById('search-input').focus();
			}
		}

		function searchText() {
			const searchTerm = document.getElementById('search-input').value;
			const editor = document.getElementById('editor');
			const text = editor.value;
			const regex = new RegExp(searchTerm, 'gi');
			
			// Clear previous highlights
			const highlights = document.querySelectorAll('.search-highlight');
			highlights.forEach(h => h.classList.remove('search-highlight'));
			
			if (searchTerm) {
				let match;
				let lastIndex = 0;
				while ((match = regex.exec(text)) !== null) {
					// Highlight the match
					editor.setSelectionRange(match.index, match.index + match[0].length);
					lastIndex = match.index;
				}
				if (lastIndex) {
					editor.focus();
				}
			}
		}

		function replaceText() {
			const searchTerm = document.getElementById('search-input').value;
			const replaceWith = document.getElementById('replace-input').value;
			const editor = document.getElementById('editor');
			
			if (searchTerm) {
				const regex = new RegExp(searchTerm, 'g');
				editor.value = editor.value.replace(regex, replaceWith);
				updatePreview();
				updateWordCount();
			}
		}

		function saveToLocalStorage(content) {
			try {
				localStorage.setItem('markdown-content', content);
				updateStatus('success', 'Content saved');
			} catch (e) {
				console.error('Save error:', e);
				updateStatus('error', 'Failed to save content');
			}
		}

		function updateStatus(type, message) {
			const status = document.getElementById('status');
			status.textContent = message;
			status.className = 'status ' + type + ' visible';
			setTimeout(() => {
				status.className = 'status';
				status.textContent = isOnline ? 'Connected' : 'Offline';
			}, 3000);
		}

		// Handle WebSocket connection
		function connectWebSocket() {
			const ws = new WebSocket('ws://' + window.location.host + '/ws');
			
			ws.onopen = () => {
				isOnline = true;
				updateStatus('success', 'Connected');
			};

			ws.onclose = () => {
				isOnline = false;
				updateStatus('error', 'Disconnected');
				// Try to reconnect after 5 seconds
				setTimeout(connectWebSocket, 5000);
			};

			ws.onerror = () => {
				isOnline = false;
				updateStatus('error', 'Connection error');
			};

			ws.onmessage = (event) => {
				if (event.data === 'reload') {
					updatePreview();
				}
			};
		}

		// Keyboard shortcuts
		function handleShortcuts(e) {
			if (e.ctrlKey || e.metaKey) {
				switch(e.key) {
					case 's':
						e.preventDefault();
						saveToLocalStorage(document.getElementById('editor').value);
						break;
					case 'b':
						e.preventDefault();
						document.body.classList.toggle('preview-only');
						break;
				}
			}
		}

		function handleImageUpload(file) {
			const formData = new FormData();
			formData.append('image', file);

			fetch('/upload', {
				method: 'POST',
				body: formData
			})
			.then(response => response.json())
			.then(data => {
				const editor = document.getElementById('editor');
				const cursorPos = editor.selectionStart;
				const textBefore = editor.value.substring(0, cursorPos);
				const textAfter = editor.value.substring(cursorPos);
				
				editor.value = textBefore + data.markdown + textAfter;
				editor.selectionStart = cursorPos + data.markdown.length;
				editor.selectionEnd = editor.selectionStart;
				editor.focus();
				updatePreview();
			})
			.catch(error => {
				updateStatus('error', 'Failed to upload image');
				console.error('Upload error:', error);
			});
		}

		function handlePaste(e) {
			const items = (e.clipboardData || e.originalEvent.clipboardData).items;
			
			for (const item of items) {
				if (item.type.indexOf('image') === 0) {
					e.preventDefault();
					const file = item.getAsFile();
					handleImageUpload(file);
					return;
				}
			}
		}

		function handleDrop(e) {
			e.preventDefault();
			e.stopPropagation();
			
			const files = e.dataTransfer.files;
			for (const file of files) {
				if (file.type.indexOf('image') === 0) {
					handleImageUpload(file);
					return;
				}
			}
		}

		function handleDragOver(e) {
			e.preventDefault();
			e.stopPropagation();
		}

		function toggleGuide() {
			const guide = document.getElementById('guide');
			if (!guide) {
				const guideFrame = document.createElement('iframe');
				guideFrame.id = 'guide';
				guideFrame.src = '/guide';
				guideFrame.className = 'guide-pane';
				document.body.appendChild(guideFrame);
				document.body.classList.add('guide-open');
				guideVisible = true;
			} else {
				guide.remove();
				document.body.classList.remove('guide-open');
				guideVisible = false;
			}
		}

		// Initialize
		document.addEventListener('DOMContentLoaded', function() {
			const editor = document.getElementById('editor');
			const resizer = document.getElementById('resizer');
			let isResizing = false;
			let lastX;

			// Editor changes
			let timeout;
			editor.addEventListener('input', function() {
				clearTimeout(timeout);
				timeout = setTimeout(updatePreview, 150);
				updateWordCount();
			});

			// Update line numbers on scroll
			editor.addEventListener('scroll', function() {
				document.getElementById('line-numbers').scrollTop = editor.scrollTop;
			});

			// Initialize line numbers
			function updateLineNumbers() {
				const lineCount = editor.value.split('\n').length;
				const lineNumbers = document.getElementById('line-numbers');
				lineNumbers.innerHTML = Array(lineCount).fill(0).map((_, i) => 
					'<div class="line-number">' + (i + 1) + '</div>').join('');
			}
			
			editor.addEventListener('input', updateLineNumbers);
			updateLineNumbers();

			// Resizer handling
			resizer.addEventListener('mousedown', (e) => {
				isResizing = true;
				lastX = e.clientX;
			});

			document.addEventListener('mousemove', (e) => {
				if (!isResizing) return;

				const container = document.querySelector('.container');
				const editorPane = document.querySelector('.editor-pane');
				const delta = e.clientX - lastX;
				const newWidth = editorPane.offsetWidth + delta;
				
				if (newWidth > 100 && newWidth < container.offsetWidth - 100) {
					editorPane.style.width = newWidth + 'px';
					lastX = e.clientX;
				}
			});

			document.addEventListener('mouseup', () => {
				isResizing = false;
			});

			// Keyboard shortcuts
			document.addEventListener('keydown', handleShortcuts);

			// Image handling
			editor.addEventListener('paste', handlePaste);
			editor.addEventListener('drop', handleDrop);
			editor.addEventListener('dragover', handleDragOver);

			// File input handling
			document.getElementById('image-input').addEventListener('change', function(e) {
				if (e.target.files.length > 0) {
					handleImageUpload(e.target.files[0]);
				}
			});

			// Load saved content or trigger initial preview
			const saved = localStorage.getItem('markdown-content');
			if (saved) {
				editor.value = saved;
				updateLineNumbers();
			}
			updatePreview();
			updateWordCount();
			connectWebSocket();

			// Theme handling
			const savedTheme = localStorage.getItem('theme') || 'light';
			document.documentElement.setAttribute('data-theme', savedTheme);
			updateThemeIcon();

			// Initialize tooltips with animation
			const tooltips = document.querySelectorAll('.tooltip');
			tooltips.forEach(tooltip => {
				tooltip.addEventListener('mouseenter', () => {
					const tooltipText = tooltip.querySelector('.tooltip-text');
					tooltipText.style.animation = 'fadeIn 0.2s ease-in-out';
				});
			});

			// Add subtle hover effect to guide examples
			const guideExamples = document.querySelectorAll('.guide-example');
			guideExamples.forEach(example => {
				example.addEventListener('mouseenter', () => {
					example.style.transform = 'translateY(-2px)';
					example.style.transition = 'transform 0.2s ease-in-out';
				});
				example.addEventListener('mouseleave', () => {
					example.style.transform = 'translateY(0)';
				});
			});

			// Add ripple effect to buttons
			document.querySelectorAll('button').forEach(button => {
				button.addEventListener('click', function(e) {
					const rect = button.getBoundingClientRect();
					const ripple = document.createElement('div');
					ripple.className = 'ripple';
					ripple.style.left = (e.clientX - rect.left) + 'px';
					ripple.style.top = (e.clientY - rect.top) + 'px';
					button.appendChild(ripple);
					
					setTimeout(() => ripple.remove(), 600);
				});
			});

			// Enhance toolbar button interactions
			document.querySelectorAll('.toolbar button').forEach(button => {
				button.addEventListener('mouseenter', () => {
					button.style.transform = 'translateY(-1px)';
				});
				button.addEventListener('mouseleave', () => {
					button.style.transform = 'translateY(0)';
				});
			});

			// Add smooth scrolling to preview
			document.querySelector('.preview-pane').style.scrollBehavior = 'smooth';
		});

		function toggleTheme() {
			const currentTheme = document.documentElement.getAttribute('data-theme');
			const newTheme = currentTheme === 'light' ? 'dark' : 'light';
			document.documentElement.setAttribute('data-theme', newTheme);
			localStorage.setItem('theme', newTheme);
			updateThemeIcon();
		}

		function updateThemeIcon() {
			const themeIcon = document.querySelector('.bi-moon-stars');
			const currentTheme = document.documentElement.getAttribute('data-theme');
			if (currentTheme === 'dark') {
				themeIcon.classList.remove('bi-moon-stars');
				themeIcon.classList.add('bi-sun');
			} else {
				themeIcon.classList.remove('bi-sun');
				themeIcon.classList.add('bi-moon-stars');
			}
		}
	</script>`

	const htmlEnd = `</script>
</head>
<body>
	<div class="container">
		<div class="editor-pane">
			<div class="toolbar">
				<div class="toolbar-group">
					<button onclick="insertMarkdown('h1')" title="Heading 1"><i class="bi bi-type-h1"></i></button>
					<button onclick="insertMarkdown('h2')" title="Heading 2"><i class="bi bi-type-h2"></i></button>
					<button onclick="insertMarkdown('h3')" title="Heading 3"><i class="bi bi-type-h3"></i></button>
				</div>
				<div class="toolbar-group">
					<button onclick="insertMarkdown('bold')" title="Bold"><i class="bi bi-type-bold"></i></button>
					<button onclick="insertMarkdown('italic')" title="Italic"><i class="bi bi-type-italic"></i></button>
					<button onclick="insertMarkdown('code')" title="Code"><i class="bi bi-code"></i></button>
				</div>
				<div class="toolbar-group">
					<button onclick="insertMarkdown('link')" title="Link"><i class="bi bi-link"></i></button>
					<button onclick="insertMarkdown('list')" title="List"><i class="bi bi-list-ul"></i></button>
					<button onclick="insertMarkdown('quote')" title="Quote"><i class="bi bi-chat-quote"></i></button>
				</div>
				<div class="toolbar-group">
					<button onclick="document.getElementById('image-input').click()" title="Upload Image">
						<i class="bi bi-image"></i>
					</button>
					<input type="file" id="image-input" accept="image/*" style="display: none">
					<button onclick="toggleSearch()" title="Search"><i class="bi bi-search"></i></button>
					<button onclick="toggleGuide()" title="Markdown Guide"><i class="bi bi-question-circle"></i></button>
				</div>
				<div class="toolbar-group">
					<button onclick="toggleTheme()" title="Toggle Dark Mode">
						<i class="bi bi-moon-stars"></i>
					</button>
				</div>
			</div>
			<div id="search-bar" style="display: none;">
				<input type="text" id="search-input" placeholder="Search..." onkeyup="searchText()">
				<input type="text" id="replace-input" placeholder="Replace with...">
				<button onclick="replaceText()">Replace</button>
				<button onclick="toggleSearch()">Close</button>
			</div>
			<div class="editor-container">
				<div id="line-numbers"></div>
				<textarea id="editor" spellcheck="false" placeholder="Type or paste your Markdown here..."># Sample Markdown

This is a sample markdown file to test the previewer.

## Features

- Real-time preview
- Split view with resizable panes
- Clean, modern styling
- Syntax highlighting
- Auto-save
- Offline support
- Keyboard shortcuts (Ctrl+S, Ctrl+B)
- Line numbers
- Word count
- Markdown toolbar
- Search and replace

### Code Example

` + "```" + `go
package main

import "fmt"

func main() {
	fmt.Println("Hello, Markdown Previewer!")
}
` + "```" + `</textarea>
				</div>
				<div id="word-count" class="word-count"></div>
			</div>
			<div id="resizer" class="resizer"></div>
			<div class="preview-pane">
				<div id="preview" class="preview">
					<div class="loading">Loading preview...</div>
				</div>
			</div>
		</div>
		<div id="status" class="status">Connecting...</div>
	</div>
</body>
</html>`

	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(htmlStart + jsCode + htmlEnd))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}
	defer conn.Close()

	// Watch for file changes
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Println(err)
		return
	}
	defer watcher.Close()

	err = watcher.Add(*markdownFile)
	if err != nil {
		log.Println(err)
		return
	}

	for {
		select {
		case event := <-watcher.Events:
			if event.Op&fsnotify.Write == fsnotify.Write {
				conn.WriteMessage(websocket.TextMessage, []byte("reload"))
			}
		case err := <-watcher.Errors:
			log.Println("error:", err)
		}
	}
}
