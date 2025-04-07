# Security Checklist for Markdown Previewer

This checklist should be reviewed regularly during development to ensure security best practices are followed.

## Input Validation and Sanitization

- [ ] All user-supplied input (file paths, command-line arguments) is validated
- [ ] Path traversal protection is implemented for file access
- [ ] Markdown content is sanitized before processing and rendering
- [ ] HTML output is properly escaped to prevent XSS attacks

## File System Access

- [ ] File access is limited to designated directories
- [ ] No write access to system directories
- [ ] File permissions are set appropriately
- [ ] Symlinks are handled securely

## Web Server

- [ ] Server only binds to localhost by default (unless explicitly configured)
- [ ] CORS settings are appropriate for local use
- [ ] HTTP headers are configured properly
  - [ ] X-Content-Type-Options: nosniff
  - [ ] X-Frame-Options: DENY
  - [ ] Content-Security-Policy: appropriate restrictions
- [ ] Rate limiting implemented for API endpoints

## WebSocket Security

- [ ] Origin checking implemented
- [ ] Messages are validated and sanitized
- [ ] Connection timeouts are implemented
- [ ] Proper error handling for WebSocket connections

## Error Handling

- [ ] Errors are logged appropriately
- [ ] Error messages don't leak sensitive information
- [ ] All error conditions are handled gracefully

## Dependency Management

- [ ] Dependencies are kept up to date
- [ ] Dependencies are scanned for vulnerabilities (e.g., with `go list -m -json all | nancy sleuth`)
- [ ] Minimum required dependencies are used

## Configuration

- [ ] No hard-coded secrets or credentials
- [ ] Default configuration is secure
- [ ] Configuration options are validated

## Resource Management

- [ ] Resources (file handles, goroutines, etc.) are properly closed/cleaned up
- [ ] Memory usage is monitored and limited
- [ ] Appropriate timeouts for operations

## Documentation

- [ ] Security considerations are documented
- [ ] Proper usage instructions include security notes
- [ ] Any limitations or potential risks are clearly communicated

## Regular Security Practices

- [ ] Run `gosec` analysis tool regularly
- [ ] Incorporate security checks in CI pipeline
- [ ] Keep Go updated to latest secure version