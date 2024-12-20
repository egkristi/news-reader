# News Reader

A modern news aggregator that supports multiple content types (articles, videos, podcasts) with advanced categorization and tagging.

## Features

- Multi-source news aggregation
- Support for RSS feeds, YouTube channels, and podcasts
- Automatic content categorization
- Custom tagging system
- Content filtering by type, category, and interests
- Modern, responsive UI

## Project Structure

```
news-reader/
├── cmd/
│   └── server/          # Application entrypoint
├── internal/
│   ├── handlers/        # HTTP request handlers
│   ├── models/          # Data models
│   └── services/        # Business logic
├── web/
│   ├── static/          # Static assets
│   └── templates/       # HTML templates
├── Dockerfile          # Container definition
├── Makefile           # Build automation
├── README.md          # This file
└── go.mod             # Go module definition
```

## Development

### Prerequisites

- Go 1.21 or later
- Make (optional, for using Makefile commands)

### Getting Started

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/news-reader.git
   cd news-reader
   ```

2. Install dependencies:
   ```bash
   go mod download
   ```

3. Run the application:
   ```bash
   # Using make
   make dev

   # Or using go directly
   go run ./cmd/server/main.go
   ```

4. Visit `http://localhost:8082` in your browser

### Build

```bash
# Build the binary
make build

# Run the built binary
make run
```

## Docker

### Build the Container

```bash
make docker-build
```

### Run the Container

```bash
make docker-run
```

The application will be available at `http://localhost:8082`.

## Configuration

The application accepts the following command-line flags:

- `-port`: Server port (default: "8082")
- `-prefs`: Path to preferences file (default: "preferences.json")
- `-debug`: Enable debug mode (default: true)

Example:
```bash
./news-reader -port 8083 -prefs /data/prefs.json -debug=false
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the LICENSE file for details.
