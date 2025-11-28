# Contributing to Go Poker Arena

First off, thank you for considering contributing to Go Poker Arena! 🎉

## 🤝 How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check the existing issues to avoid duplicates. When you create a bug report, include as many details as possible:

- **Use a clear and descriptive title**
- **Describe the exact steps to reproduce the problem**
- **Provide specific examples**
- **Describe the behavior you observed and what you expected**
- **Include screenshots if relevant**
- **Include your environment details** (OS, browser, Go version, etc.)

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion:

- **Use a clear and descriptive title**
- **Provide a detailed description of the suggested enhancement**
- **Explain why this enhancement would be useful**
- **List any similar features in other applications**

### Pull Requests

1. **Fork the repository** and create your branch from `dev`
2. **Make your changes** following our coding standards
3. **Add tests** if you're adding functionality
4. **Ensure tests pass** (`go test ./...` for backend, `npm test` for frontend)
5. **Update documentation** if needed
6. **Write a clear commit message** following our commit conventions
7. **Submit a pull request** to the `dev` branch

## 📝 Coding Standards

### Backend (Go)

- Follow [Effective Go](https://golang.org/doc/effective_go.html) guidelines
- Use `gofmt` to format your code
- Run `golangci-lint` before committing
- Write meaningful variable and function names
- Add comments for exported functions
- Keep functions small and focused

```go
// Good
func CalculateWinner(players []Player, communityCards []Card) *Player {
    // Implementation
}

// Bad
func calc(p []Player, c []Card) *Player {
    // Implementation
}
```

### Frontend (TypeScript/React)

- Use TypeScript for all new code
- Follow React best practices and hooks guidelines
- Use functional components over class components
- Keep components small and reusable
- Use meaningful prop names
- Add JSDoc comments for complex functions

```typescript
// Good
interface PlayerCardProps {
  player: Player
  isActive: boolean
  onAction: (action: GameAction) => void
}

// Bad
interface Props {
  p: any
  a: boolean
  f: Function
}
```

### Styling

- Use Tailwind CSS utility classes
- Follow mobile-first responsive design
- Maintain consistent spacing and colors
- Use Framer Motion for animations

## 🔀 Git Workflow

### Branch Naming

- `feature/description` - New features
- `fix/description` - Bug fixes
- `docs/description` - Documentation updates
- `refactor/description` - Code refactoring
- `test/description` - Test additions/updates

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation changes
- `style`: Code style changes (formatting, etc.)
- `refactor`: Code refactoring
- `test`: Test additions/updates
- `chore`: Build process or auxiliary tool changes

**Examples:**
```
feat(game): Add tournament mode support

Implement multi-table tournament functionality with
blind level increases and player elimination tracking.

Closes #123
```

```
fix(websocket): Resolve connection timeout issue

Fix WebSocket reconnection logic to properly handle
network interruptions and maintain game state.

Fixes #456
```

## 🧪 Testing

### Backend Tests

```bash
cd backend
go test -v -race -coverprofile=coverage.out ./...
```

### Frontend Tests

```bash
cd frontend
npm run lint
npm run type-check
npm run build
```

## 📚 Documentation

- Update README.md if you change functionality
- Add JSDoc/GoDoc comments for new functions
- Update API documentation for endpoint changes
- Add examples for new features

## 🎨 UI/UX Guidelines

- Maintain consistent design language
- Ensure accessibility (ARIA labels, keyboard navigation)
- Test on multiple screen sizes
- Use smooth animations (Framer Motion)
- Provide loading states and error messages

## 🚀 Deployment

- Test locally with Docker before submitting PR
- Ensure environment variables are documented
- Update deployment docs if needed

## 📞 Getting Help

- Join our [Discussions](https://github.com/parsapap/go-poker-arena/discussions)
- Ask questions in issues with the `question` label
- Check existing documentation in the `docs/` folder

## 📜 Code of Conduct

- Be respectful and inclusive
- Welcome newcomers and help them learn
- Focus on constructive feedback
- Respect differing viewpoints and experiences

## 🎯 Priority Areas

We're especially interested in contributions for:

1. **Performance Optimization** - Make it faster!
2. **Mobile Experience** - Improve responsive design
3. **Accessibility** - WCAG compliance
4. **Testing** - Increase test coverage
5. **Documentation** - Help others understand the code
6. **Animations** - More delightful interactions
7. **Game Features** - Tournament mode, spectator mode, etc.

## 🏆 Recognition

Contributors will be:
- Listed in our README
- Mentioned in release notes
- Given credit in commit messages

Thank you for contributing to Go Poker Arena! 🎰
