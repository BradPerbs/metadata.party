# Contributing to metadata.party

First off, thank you for considering contributing to metadata.party! 🎉

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the issue list as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title**
* **Describe the exact steps to reproduce the problem**
* **Provide the URL you were trying to extract metadata from**
* **Describe the behavior you observed and what behavior you expected**
* **Include the full error message or response**

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title**
* **Provide a detailed description of the suggested enhancement**
* **Explain why this enhancement would be useful**
* **List any alternative solutions you've considered**

### Pull Requests

1. Fork the repo and create your branch from `main`
2. If you've added code, ensure it follows Go best practices
3. Ensure your code passes `go vet` and `go fmt`
4. Make sure your commits are well-formatted and descriptive
5. Issue that pull request!

## Development Setup

```bash
# Clone your fork
git clone https://github.com/yourusername/metadata.party.git
cd metadata.party

# Install dependencies
go mod download

# Run the server
go run main.go

# Test your changes
curl -X POST http://localhost:8080/extract \
  -H "Content-Type: application/json" \
  -d '{"url": "https://example.com"}'
```

## Code Style

* Follow standard Go conventions
* Run `go fmt` before committing
* Run `go vet` to catch common mistakes
* Keep functions focused and reasonably sized
* Comment exported functions and complex logic

## Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

## Testing

While we don't have automated tests yet, please manually test your changes with various URLs:

* News websites
* Social media platforms
* Blogs
* E-commerce sites
* Sites with unusual or missing metadata

## License

By contributing, you agree that your contributions will be licensed under the MIT License.

