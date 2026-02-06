# AI Assistant Configuration for Fleet Platform

This document explains how to configure various AI assistants and IDEs for optimal development experience on this fleet management platform.

## 🚀 Quick Start

The following configuration files have been created to help AI assistants understand your project:

```
flyfleet/
├── .cursor/cursorrules              # Cursor IDE rules
├── .github/copilot-instructions.md  # GitHub Copilot instructions  
├── .claude/rules.md                 # Claude AI rules
├── .qoder/skills/skills.yaml        # Qoder skills configuration
├── .vscode/settings.json           # VS Code settings
└── .vscode/extensions.json         # Recommended VS Code extensions
```

## 🤖 AI Assistant Configurations

### GitHub Copilot
**Location**: `.github/copilot-instructions.md`

Provides comprehensive guidelines for:
- Technology stack enforcement
- Code quality standards
- Security requirements
- Performance benchmarks
- Implementation workflows
- Common code patterns
- Error handling strategies

**Best for**: Developers using GitHub Copilot in VS Code or JetBrains IDEs.

### Cursor IDE
**Location**: `.cursor/cursorrules`

Focuses on:
- Project-specific conventions
- File organization patterns
- Code generation workflows
- Testing requirements
- Git conventions
- Performance benchmarks

**Best for**: Teams using Cursor IDE with advanced AI features.

### Claude AI
**Location**: `.claude/rules.md`

Emphasizes:
- System persona and expertise
- Core development principles
- Technology stack enforcement
- Security and performance standards
- Detailed implementation workflows
- Code review checklists

**Best for**: Developers using Claude for code reviews and architectural discussions.

### Qoder
**Location**: `.qoder/skills/skills.yaml`

Structured YAML configuration with:
- Language-specific rules and patterns
- Technology stack guidelines
- Security and performance requirements
- Testing standards
- Git conventions
- Code review checklists

**Best for**: Teams using Qoder IDE with skills-based assistance.

## ⚙️ IDE Configuration

### VS Code Settings
**Location**: `.vscode/settings.json`

Configures:
- Auto-formatting on save
- Language-specific formatters
- Go tools (gofumpt, golangci-lint)
- TypeScript/React settings
- Tailwind CSS integration
- Search and file exclusions

### Recommended Extensions
**Location**: `.vscode/extensions.json`

Includes extensions for:
- Go development
- TypeScript/React
- Docker/Kubernetes
- Kotlin/Android
- Git utilities
- Formatting and linting tools

## 🛠️ How to Use

### For GitHub Copilot Users
1. Ensure your repository has the `.github/copilot-instructions.md` file
2. Copilot will automatically use these instructions
3. For best results, reference the file in your prompts:
   ```
   Following the guidelines in .github/copilot-instructions.md, help me...
   ```

### For Cursor IDE Users
1. Place the `.cursor/cursorrules` file in your project root
2. Cursor will automatically apply these rules
3. You can reference specific sections in conversations:
   ```
   According to the cursorrules file, when creating API endpoints I should...
   ```

### For Claude AI Users
1. Copy the contents of `.claude/rules.md` to Claude's custom instructions
2. Or reference the file directly in conversations:
   ```
   Following the rules in .claude/rules.md, please help me implement...
   ```

### For Qoder Users
1. The `.qoder/skills/skills.yaml` file is already in the correct location
2. Qoder will automatically load these skills
3. You can activate specific skills by mentioning them:
   ```
   Using the go skill, help me implement...
   ```

## 📋 Key Features Covered

### Technology Stack Enforcement
- **Go**: Gin, pgx, zap logging, clean architecture
- **Database**: PostgreSQL, TimescaleDB, proper indexing
- **Frontend**: React, TypeScript, Redux, Tailwind CSS
- **Mobile**: Kotlin, Jetpack Compose, Hilt, Room
- **Infrastructure**: Docker, Kubernetes, monitoring stack

### Code Quality Standards
- Structured error handling
- Context-aware operations
- Proper validation and sanitization
- Performance optimization guidelines
- Security best practices

### Development Workflows
- Database-first design approach
- Test-driven development
- Clean architecture patterns
- Proper documentation standards
- Git commit conventions

### Performance Requirements
- Query performance targets (< 100ms p99)
- API response times (< 200ms p99)
- Real-time data processing (1000+ GPS points/sec)
- Mobile and web performance benchmarks

## 🔧 Customization

Feel free to modify these configurations based on your team's specific needs:

1. **Technology Changes**: Update the tech stack sections
2. **Standards Adjustment**: Modify performance or coverage requirements
3. **Process Modifications**: Adapt workflows to your team's practices
4. **Tool Preferences**: Change formatter or linter preferences

## 🤝 Team Collaboration

Share these configurations with your team to ensure consistency:

```bash
# Clone the repository
git clone <repository-url>

# The AI configurations are already included
# Each team member should configure their preferred AI assistant
# using the appropriate files mentioned above
```

## 🆘 Troubleshooting

### AI Assistant Not Following Rules
- Ensure configuration files are in the correct locations
- Restart your IDE/assistant after adding configurations
- Explicitly reference the configuration files in your prompts

### Formatting Issues
- Check that recommended extensions are installed
- Verify VS Code settings are applied
- Ensure language-specific formatters are configured

### Performance Concerns
- Review the performance benchmarks section
- Consider adjusting requirements based on your infrastructure
- Monitor actual performance vs. targets

## 📚 Additional Resources

- [Project Documentation](./README.md)
- [Development Guidelines](./DEVELOPMENT.md)
- [Architecture Overview](./ARCHITECTURE.md)
- [Testing Strategy](./TESTING.md)

---
*Last updated: February 2026*