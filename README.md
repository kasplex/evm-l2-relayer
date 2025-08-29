# Kasplex Relayer

A high-performance blockchain relayer that bridges Ethereum and Kaspa networks, enabling seamless cross-chain VM (Virtual Machine) transaction processing.

## Overview

Kasplex Relayer is a Go-based service that acts as an intermediary between Ethereum and Kaspa blockchains. It intercepts Ethereum JSON-RPC calls, processes VM transactions, and forwards them to the Kaspa network for execution. The relayer supports both raw transaction data and structured transaction objects.

## Features

- **Cross-Chain Bridging**: Seamlessly bridges Ethereum and Kaspa networks
- **VM Transaction Support**: Handles virtual machine transactions with data compression
- **JSON-RPC Proxy**: Acts as a proxy for Ethereum RPC calls
- **Transaction Routing**: Intelligently routes transactions based on method type
- **Wallet Integration**: Built-in Kaspa wallet functionality for transaction signing
- **Configurable**: Flexible configuration via TOML files and environment variables
- **High Performance**: Efficient client pooling and concurrent processing

## Architecture

The relayer consists of several key components:

- **Relayer**: Main HTTP server that handles incoming requests
- **Wallet**: Kaspa wallet implementation for transaction management
- **Client Pool**: Connection pooling for Kaspa RPC clients
- **Config Manager**: Configuration loading and validation
- **Logging**: Structured logging with configurable levels

## Prerequisites

- Go 1.23.10 or higher
- Access to Ethereum RPC endpoint
- Access to Kaspa RPC endpoint
- Kaspa private key for transaction signing

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/kasplex-evm/kasplex-relayer.git
cd kasplex-relayer

# Build the binary
make build

# The binary will be available in ./dist/relayer
```

### Using Make

```bash
# Build the project
make build

# View available commands
make help
```

## Configuration

The relayer uses TOML configuration files. A default configuration file `config.default.toml` is provided:

```toml
[Log]
Level = "info"
Outputs = ["stdout"]

[Relayer]
EthRPC = "http://127.0.0.1:8545"
KasRPC = "10.0.1.197:16210"
Port = 8080
PrivateKey = ""
ToAddress = "kaspatest:qrheh44dm6gckcqux4tefmsf55slqaj7u2qseclcp94696ha0da3skr8tlhdd"
```

### Configuration Options

#### Log Section
- `Level`: Log level (debug, info, warn, error, dpanic, panic, fatal)
- `Outputs`: Array of log output destinations

#### Relayer Section
- `EthRPC`: Ethereum RPC endpoint URL
- `KasRPC`: Kaspa RPC endpoint (host:port)
- `Port`: HTTP server port for the relayer
- `PrivateKey`: Kaspa private key for transaction signing (hex format)
- `ToAddress`: Destination Kaspa address for VM transactions

### Environment Variables

Configuration can also be set via environment variables with the `RELAYER_` prefix:

```bash
export RELAYER_RELAYER_ETHRPC="http://localhost:8545"
export RELAYER_RELAYER_KASRPC="localhost:16210"
export RELAYER_RELAYER_PORT="8080"
export RELAYER_RELAYER_PRIVATEKEY="your_private_key_here"
export RELAYER_RELAYER_TOADDRESS="kaspatest:destination_address"
```

## Usage

### Starting the Relayer

```bash
# Using default configuration
./dist/relayer run

# Using custom configuration file
./dist/relayer run -c /path/to/config.toml

# Check version
./dist/relayer version
```

### API Endpoints

The relayer exposes a single HTTP endpoint that acts as a JSON-RPC proxy:

- **POST /** - Main endpoint for all JSON-RPC calls

#### Supported Methods

- `eth_sendRawTransaction` - Intercepts and processes VM transactions
- All other methods are proxied to the configured Ethereum RPC endpoint

#### Example Request

```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "eth_sendRawTransaction",
  "params": ["0x..."]
}
```

## Transaction Flow

1. **Request Reception**: Relayer receives JSON-RPC request
2. **Method Detection**: Identifies the method type
3. **VM Processing**: For `eth_sendRawTransaction`, processes VM data
4. **Kaspa Transfer**: Creates and sends Kaspa transaction with VM data
5. **Response**: Returns transaction hash to client

## Development

### Project Structure

```
.
├── cmd/                    # Command-line interface
│   └── relayer/           # Main application entry point
├── config/                 # Configuration management
├── impl/                   # Core implementation
│   ├── relayer.go         # Main relayer logic
│   ├── wallet.go          # Kaspa wallet operations
│   ├── client_pool.go     # RPC client pooling
│   └── utils.go           # Utility functions
├── log/                    # Logging configuration
├── go.mod                  # Go module dependencies
├── Makefile               # Build automation
└── config.default.toml    # Default configuration
```

### Building

```bash
# Build for current platform
make build

# Build with version information
VERSION=v1.0.0 make build
```

### Dependencies

Key dependencies include:
- `github.com/kaspanet/kaspad` - Kaspa blockchain client
- `github.com/kaspanet/go-secp256k1` - Cryptographic operations
- `github.com/spf13/viper` - Configuration management
- `github.com/urfave/cli/v2` - Command-line interface
- `go.uber.org/zap` - Structured logging

## Security Considerations

- **Private Key Management**: Ensure private keys are stored securely
- **Network Security**: Use secure connections for RPC endpoints
- **Access Control**: Implement proper access controls for the relayer API
- **Transaction Validation**: Validate all incoming transactions before processing

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests if applicable
5. Submit a pull request

## License

This project is licensed under the MIT License.

## Support

For issues and questions:
- Create an issue on GitHub
- Check the configuration examples
- Review the logs for debugging information

## Version

Current version: v0.1.0

For build information, use:
```bash
./dist/relayer version
```
