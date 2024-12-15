# MTC CLI

A command-line interface tool for validating MoreThanCertified lesson tasks locally.

## Features

- Submit lesson tasks for validation
- Interactive progress tracking

## Installation

### With Go

```bash
go install github.com/morethancertified/mtc-cli
```

## Usage

Submit a lesson for validation:

```bash
mtc submit <lesson-token>
```

Example:

```bash
mtc submit cm4ppz694200blze51ts1234
```

## Development

The project uses several development tools and commands:

### Available Commands

Reference to justfile commands:

- `just run` - Run the application from source
- `just build` - Build the binary to bin/mtc-cli
- `just test` - Run tests
- `just fmt` - Format code
- `just clean` - Clean build artifacts
- `just install` - Install the binary
- `just uninstall` - Uninstall the binary

### Configuration

The CLI uses Viper for configuration management. By default, it creates a config file at:

- `$HOME/.config/mtc/config.json`

You can override the config location using the `--config` flag.

### Project Structure

- `cmd/` - Command implementations
- `internal/` - Internal packages
  - `mtcapi/` - API client implementation
  - `widgets/` - TUI components

## Building for Distribution

The project uses GoReleaser for building and distributing releases. Configuration can be found in:

- [.goreleaser.yaml](.goreleaser.yaml)

## License

MIT License - See LICENSE file for details.

## Contributing

1. Clone the repository (`git clone https://github.com/morethancertified/mtc-cli`)
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. [Open a Pull Request](https://github.com/morethancertified/mtc-cli/compare)

## Dependencies

Key dependencies include:

- `github.com/spf13/cobra` - CLI framework
- `github.com/spf13/viper` - Configuration management
- `github.com/charmbracelet/bubbles` - Terminal UI components
- `github.com/go-resty/resty/v2` - HTTP client

For a complete list of dependencies, see: [go.mod](go.mod)
