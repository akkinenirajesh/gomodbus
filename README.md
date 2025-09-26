# gomodbus

A powerful, enhanced command-line tool for Modbus communication that replicates all **mbpoll** functionality while adding advanced capabilities from the **gomodbus** library.

## 🚀 Overview

`gomodbus` is a feature-complete replacement for the popular `mbpoll` tool, built with modern Go and the robust gomodbus library. It provides all the functionality of mbpoll plus additional transport modes, enhanced data types, and improved error handling.

## ✨ Key Features

### 🔄 **Complete mbpoll Compatibility**
- ✅ Read/Write Coils (0x01, 0x05, 0x0F)
- ✅ Read/Write Discrete Inputs (0x02)
- ✅ Read/Write Input Registers (0x03, 0x04)
- ✅ Read/Write Holding Registers (0x06, 0x10)
- ✅ Multiple data formats (decimal, hex, 16/32-bit integers, 32-bit floats)
- ✅ Configurable polling with continuous and one-time modes
- ✅ Full RTU serial configuration (baudrate, parity, databits, stopbits)

### 🚀 **Enhanced Capabilities**
- **6 Transport Modes**: TCP, TLS, UDP, RTU, RTU-over-TCP, RTU-over-UDP
- **Rich Data Types**: 16/32/64-bit integers, 32/64-bit floats, configurable endianness
- **Advanced Error Handling**: User-friendly messages with helpful suggestions
- **Cross-Platform**: Pure Go implementation works on Windows, Linux, macOS
- **Modern Architecture**: Connection pooling, timeout management, verbose debugging

## 📦 Installation

### Prerequisites
- Go 1.19 or later
- Serial port access (for RTU modes)

### Quick Install
```bash
# Clone the repository
git clone <repository-url>
cd gomodbus

# Build the CLI
go build -o gomodbus main.go

# (Optional) Install to system PATH
sudo cp gomodbus /usr/local/bin/
```

### Dependencies
The project automatically manages dependencies:
```bash
go mod tidy  # Install/update dependencies
```

## 🎯 Usage

### Basic Syntax
```bash
gomodbus [OPTIONS] DEVICE|HOST [WRITE_VALUES...] [OPTIONS]
```

### Essential Options

| Option | Description | Default |
|--------|-------------|---------|
| `-m, --mode` | Transport mode: `tcp`, `tls`, `udp`, `rtu`, `rtuovertcp`, `rtuoverudp` | `tcp` |
| `-a, --address` | Slave address (0-255) | `1` |
| `-r, --reference` | Start reference address | `1` |
| `-c, --count` | Number of values to read (1-125) | `1` |
| `-t, --type` | Data type (see Data Types section) | `4` |
| `-p, --port` | TCP port number | `502` |

### Transport-Specific Options

#### TCP Options
```bash
gomodbus -m tcp -t 4 -r 1 -c 2 192.168.1.100
```

#### RTU Options
```bash
gomodbus -m rtu -b 19200 -d 8 -s 1 -P even -t 4 -r 1 -c 2 /dev/ttyUSB0
```

## 📊 Data Types

| Type | Description | Function Codes |
|------|-------------|----------------|
| `0` | Discrete outputs (coils) - binary | 0x01, 0x05, 0x0F |
| `1` | Discrete inputs - binary | 0x02 |
| `3` | 16-bit input registers | 0x04 |
| `3:hex` | 16-bit input registers (hex display) | 0x04 |
| `3:int` | 32-bit integers in input registers | 0x04 |
| `3:float` | 32-bit floats in input registers | 0x04 |
| `4` | 16-bit holding registers | 0x03, 0x06, 0x10 |
| `4:hex` | 16-bit holding registers (hex display) | 0x03, 0x06, 0x10 |
| `4:int` | 32-bit integers in holding registers | 0x03, 0x06, 0x10 |
| `4:float` | 32-bit floats in holding registers | 0x03, 0x06, 0x10 |

## 🌐 Transport Modes

### 1. **Modbus TCP** (Standard)
```bash
gomodbus -m tcp -t 4 -r 1 -c 2 192.168.1.100
```

### 2. **Modbus TCP over TLS** (Secure)
```bash
gomodbus -m tls -t 4 -r 1 -c 2 192.168.1.100
```

### 3. **Modbus TCP over UDP**
```bash
gomodbus -m udp -t 4 -r 1 -c 2 192.168.1.100
```

### 4. **Modbus RTU** (Serial)
```bash
gomodbus -m rtu -b 19200 -t 4 -r 1 -c 2 /dev/ttyUSB0
```

### 5. **RTU over TCP** (Tunneling)
```bash
gomodbus -m rtuovertcp -t 4 -r 1 -c 2 192.168.1.100
```

### 6. **RTU over UDP** (Tunneling)
```bash
gomodbus -m rtuoverudp -t 4 -r 1 -c 2 192.168.1.100
```

## 💡 Examples

### Reading Operations

#### Read Holding Registers (mbpoll equivalent)
```bash
gomodbus -t 4 -r 1 -c 2 192.168.1.100
```

#### Read Input Registers as 32-bit Floats
```bash
gomodbus -m rtu -t 3:float -r 1 -c 2 /dev/ttyUSB0
```

#### Read Coils with Continuous Polling
```bash
gomodbus -t 0 -r 1 -c 8 -l 500 192.168.1.100
```

#### Read with Hex Display
```bash
gomodbus -t 4:hex -r 1 -c 4 192.168.1.100
```

### Writing Operations

#### Write Single Values to Holding Registers
```bash
gomodbus -t 4 -r 1 192.168.1.100 123 456 789
```

#### Write 32-bit Integers
```bash
gomodbus -t 4:int -r 1 192.168.1.100 123456 -789012
```

#### Write 32-bit Floats
```bash
gomodbus -t 4:float -r 1 192.168.1.100 3.14 -2.71
```

#### Write Coils
```bash
gomodbus -t 0 -r 1 192.168.1.100 1 0 1 1
```

### Advanced Usage

#### RTU over TCP Tunneling
```bash
gomodbus -m rtuovertcp -b 19200 -t 4 -r 1 -c 2 192.168.1.100
```

#### Modbus TCP over TLS (Secure)
```bash
gomodbus -m tls -t 4 -r 1 -c 2 secure-modbus-server.com
```

#### High-Speed Polling with Custom Timeout
```bash
gomodbus -t 4 -r 1 -c 1 -l 100 -o 0.5 192.168.1.100
```

## ⚙️ Configuration Options

### General Options
- `-0, --zero-based`: Use 0-based addressing (PDU format)
- `-B, --big-endian`: Big endian word order for 32-bit data (default)
- `-1, --once`: Poll only once (no continuous polling)
- `-l, --poll-rate MS`: Poll rate in milliseconds (default: 1000)
- `-o, --timeout SEC`: Timeout in seconds (default: 1.0)
- `-v, --verbose`: Verbose mode for debugging

### RTU Serial Options
- `-b, --baudrate RATE`: Baudrate (1200-921600, default: 19200)
- `-d, --databits BITS`: Databits (7 or 8, default: 8)
- `-s, --stopbits BITS`: Stopbits (1 or 2, default: 1)
- `-P, --parity PARITY`: Parity (none, even, odd, default: even)

## 🔧 Error Handling

The CLI provides user-friendly error messages with helpful suggestions:

```bash
$ gomodbus
gomodbus: device or host parameter missing ! Try -h for help

$ gomodbus -c 200 192.168.1.100
gomodbus: count must be between 1 and 125

$ gomodbus -m invalid 192.168.1.100
gomodbus: unsupported mode: invalid (supported: tcp, tls, udp, rtu, rtuovertcp, rtuoverudp)
```

## 🆚 Comparison with mbpoll

| Feature | mbpoll | gomodbus |
|---------|--------|---------------|
| Transport Modes | 2 (TCP, RTU) | 6 (TCP, TLS, UDP, RTU, RTU-over-TCP, RTU-over-UDP) |
| Data Types | Basic | Extended (16/32/64-bit, configurable endianness) |
| Error Handling | Good | Enhanced (user-friendly, actionable) |
| Cross-Platform | Limited | Full (Windows, Linux, macOS) |
| Dependencies | libmodbus | Pure Go (gomodbus) |
| Maintenance | Mature | Active development |

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature-name`
3. Make your changes and add tests
4. Commit your changes: `git commit -am 'Add feature'`
5. Push to the branch: `git push origin feature-name`
6. Submit a pull request

## 📝 License

This project is licensed under the MIT License - see the LICENSE file for details.

## 🙏 Acknowledgments

- **mbpoll** - The original tool that inspired this implementation
- **gomodbus** - The excellent Go Modbus library that powers this CLI
- **libmodbus** - The C library that mbpoll is built upon

## 🔗 Links

- [mbpoll Documentation](https://manpages.ubuntu.com/manpages/jammy/man1/mbpoll.1.html)
- [gomodbus Library](https://github.com/simonvetter/modbus)
- [Modbus Specification](http://modbus.org/specs.php)

---

**gomodbus** - Enhanced Modbus communication for the modern era 🚀