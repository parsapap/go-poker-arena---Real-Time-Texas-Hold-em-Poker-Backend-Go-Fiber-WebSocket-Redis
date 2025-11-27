# Contributing to Go Poker Arena

Thank you for your interest in contributing! 🎉

## Getting Started

1. Fork the repository
2. Clone your fork: `git clone https://github.com/YOUR_USERNAME/go-poker-arena.git`
3. Create a branch: `git checkout -b feature/amazing-feature`
4. Make your changes
5. Run tests: `make test`
6. Commit: `git commit -m 'Add amazing feature'`
7. Push: `git push origin feature/amazing-feature`
8. Open a Pull Request

## Development Setup

```bash
# Install dependencies
go mod download

# Start infrastructure
make docker-up

# Run server
make run

# Run tests
make test
```

## Code Style

- Follow Go conventions
- Run `gofmt` before committing
- Add tests for new features
- Update documentation

## Pull Request Process

1. Update README.md with details of changes
2. Update API.md if adding new endpoints
3. Ensure all tests pass
4. Get approval from maintainers

## Reporting Bugs

Use GitHub Issues with:
- Clear title
- Steps to reproduce
- Expected vs actual behavior
- Environment details

## Feature Requests

Open an issue with:
- Clear description
- Use case
- Proposed solution

## Code of Conduct

Be respectful and inclusive. We're all here to build something great!

## Questions?

Open a discussion on GitHub or join our Discord.

Thank you for contributing! 🚀
