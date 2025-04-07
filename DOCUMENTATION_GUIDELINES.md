# Documentation Guidelines

This document outlines the documentation standards for the Markdown Previewer project. All contributors should follow these guidelines to ensure consistent, thorough documentation.

## Code Documentation

### Go Source Code

1. **Package Documentation**
   - Every package must have a package comment, a block comment preceding the package clause
   - Package comments should provide information about the package's purpose and usage

   ```go
   // Package parser provides utilities for parsing markdown content
   // and converting it to HTML for preview purposes.
   package parser
   ```

2. **Exported Functions, Types, and Methods**
   - All exported (capitalized) elements must have a comment
   - Comments should start with the element name
   - Include information about parameters, return values, and any error conditions

   ```go
   // ParseMarkdown converts markdown content to HTML. It accepts a byte slice
   // containing markdown content and returns the HTML representation.
   // If parsing fails, an error will be returned.
   func ParseMarkdown(content []byte) ([]byte, error) {
   ```

3. **Non-exported Elements**
   - Document non-exported elements that are complex or not immediately obvious
   - Focus on explaining "why" rather than "what"

4. **Examples**
   - Provide examples for complex functionality using Go's example testing framework
   - Ensure examples are clear and demonstrate typical usage

### Comments Best Practices

1. **Write Clear, Concise Comments**
   - Focus on why code exists rather than what it does
   - Explain anything that isn't obvious from the code itself

2. **Keep Comments Current**
   - Update comments when code changes
   - Outdated comments are worse than no comments

3. **Use Complete Sentences**
   - Start with capital letters
   - End with periods
   - Use proper grammar

## Project Documentation

### README.md

The README should include:

1. **Project Overview**
   - Clear description of what the project does
   - Key features

2. **Installation Instructions**
   - Step-by-step guide for installing the project
   - Any prerequisites

3. **Usage Examples**
   - Basic usage examples with code and screenshots
   - Common command-line options

4. **Configuration**
   - Available configuration options
   - Example configuration

5. **Contributing**
   - How to contribute to the project
   - Development setup instructions

6. **License Information**
   - Include license details or reference to LICENSE file

### Additional Documentation Files

1. **SECURITY.md**
   - Security considerations
   - Reporting vulnerabilities

2. **CHANGELOG.md**
   - Track changes between versions
   - Follow semantic versioning principles

3. **API Documentation**
   - Document any APIs provided by the application
   - Include endpoint descriptions, parameters, and responses

## Documentation Review Checklist

Before committing code, ensure documentation:

- [ ] Exists for all exported elements
- [ ] Is clear and grammatically correct
- [ ] Is up-to-date with current code behavior
- [ ] Includes examples where appropriate
- [ ] Mentions error conditions and edge cases
- [ ] Is consistent in tone and style

## Tools and Standards

1. **Tools**
   - Use `go doc` to review documentation
   - Run `golint` to catch undocumented exports
   - Consider `godocdown` for generating markdown from Go docs

2. **Standards**
   - Follow Go's official documentation style
   - Use present tense ("Gets" not "Get")
   - Be consistent with formatting

## Documentation Update Process

1. Update documentation as part of the same commit that changes code
2. Include documentation updates in pull request descriptions
3. Request documentation review as part of code review
4. Do not merge code with incomplete or outdated documentation