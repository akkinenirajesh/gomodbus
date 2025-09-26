package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/simonvetter/modbus"
)

type Config struct {
	// Connection settings
	Mode     string // "rtu" or "tcp"
	Host     string
	Port     int
	Device   string
	Baudrate int
	Databits int
	Stopbits int
	Parity   string
	Timeout  time.Duration

	// Modbus settings
	SlaveID   int
	StartRef  int
	Count     int
	DataType  string
	ZeroBased bool
	BigEndian bool
	PollOnce  bool
	PollRate  time.Duration
	Verbose   bool

	// Write values
	WriteValues []interface{}

	// RTU specific
	RTSMode int
	RTSPin  int
}

type ModbusCLI struct {
	client *modbus.ModbusClient
	config *Config
}

func main() {
	cli := &ModbusCLI{}
	if err := cli.run(); err != nil {
		fmt.Fprintf(os.Stderr, "go-modbus-cli: %v\n", err)
		os.Exit(1)
	}
}

func (m *ModbusCLI) run() error {
	config, err := m.parseArgs()
	if err != nil {
		return err
	}
	m.config = config

	if err := m.setupClient(); err != nil {
		return err
	}

	if err := m.connect(); err != nil {
		return err
	}
	defer m.client.Close()

	return m.execute()
}

func (m *ModbusCLI) parseArgs() (*Config, error) {
	config := &Config{
		Mode:      "tcp",
		Port:      502,
		SlaveID:   1,
		StartRef:  1,
		Count:     1,
		DataType:  "4", // Default to holding register
		Baudrate:  19200,
		Databits:  8,
		Stopbits:  1,
		Parity:    "even",
		Timeout:   time.Second,
		PollRate:  time.Second,
		BigEndian: true,
	}

	args := os.Args[1:]
	i := 0

	for i < len(args) {
		arg := args[i]
		switch arg {
		case "-m", "--mode":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			config.Mode = args[i+1]
			i += 2

		case "-a", "--address":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			addrStr := args[i+1]
			if strings.Contains(addrStr, ",") || strings.Contains(addrStr, ":") {
				// Multiple addresses - for now, take the first one
				parts := strings.Split(addrStr, ",")
				if len(parts) > 0 {
					addrStr = parts[0]
				}
			}
			addr, err := strconv.Atoi(addrStr)
			if err != nil {
				return nil, fmt.Errorf("invalid slave address: %v", err)
			}
			config.SlaveID = addr
			i += 2

		case "-r", "--reference":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			ref, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid reference: %v", err)
			}
			config.StartRef = ref
			i += 2

		case "-c", "--count":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			count, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid count: %v", err)
			}
			config.Count = count
			i += 2

		case "-t", "--type":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			config.DataType = args[i+1]
			i += 2

		case "-0", "--zero-based":
			config.ZeroBased = true
			i++

		case "-B", "--big-endian":
			config.BigEndian = true
			i++

		case "-1", "--once":
			config.PollOnce = true
			i++

		case "-l", "--poll-rate":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			rate, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid poll rate: %v", err)
			}
			config.PollRate = time.Duration(rate) * time.Millisecond
			i += 2

		case "-o", "--timeout":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			timeout, err := strconv.ParseFloat(args[i+1], 64)
			if err != nil {
				return nil, fmt.Errorf("invalid timeout: %v", err)
			}
			config.Timeout = time.Duration(timeout * float64(time.Second))
			i += 2

		case "-p", "--port":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			port, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid port: %v", err)
			}
			config.Port = port
			i += 2

		case "-b", "--baudrate":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			baudrate, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid baudrate: %v", err)
			}
			config.Baudrate = baudrate
			i += 2

		case "-d", "--databits":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			databits, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid databits: %v", err)
			}
			config.Databits = databits
			i += 2

		case "-s", "--stopbits":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			stopbits, err := strconv.Atoi(args[i+1])
			if err != nil {
				return nil, fmt.Errorf("invalid stopbits: %v", err)
			}
			config.Stopbits = stopbits
			i += 2

		case "-P", "--parity":
			if i+1 >= len(args) {
				return nil, fmt.Errorf("missing value for %s", arg)
			}
			config.Parity = args[i+1]
			i += 2

		case "-v", "--verbose":
			config.Verbose = true
			i++

		case "-h", "--help":
			m.printHelp()
			os.Exit(0)

		case "-V", "--version":
			fmt.Println("go-modbus-cli v1.0.0")
			os.Exit(0)

		default:
			if arg == "--" {
				// Write values start here
				config.WriteValues = make([]interface{}, 0)
				for j := i + 1; j < len(args); j++ {
					val, err := strconv.ParseFloat(args[j], 64)
					if err != nil {
						return nil, fmt.Errorf("invalid write value: %s", args[j])
					}
					config.WriteValues = append(config.WriteValues, val)
				}
				break
			} else if strings.HasPrefix(arg, "-") {
				return nil, fmt.Errorf("unknown option: %s", arg)
			} else {
				// Positional arguments
				if config.Host == "" && config.Device == "" {
					if config.Mode == "tcp" {
						config.Host = arg
					} else {
						config.Device = arg
					}
				} else if len(config.WriteValues) == 0 {
					// Write values
					for ; i < len(args); i++ {
						val, err := strconv.ParseFloat(args[i], 64)
						if err != nil {
							return nil, fmt.Errorf("invalid write value: %s", args[i])
						}
						config.WriteValues = append(config.WriteValues, val)
					}
				}
				i++
			}
		}
	}

	if config.Host == "" && config.Device == "" {
		return nil, fmt.Errorf("device or host parameter missing ! Try -h for help")
	}

	// Validation
	if err := m.validateConfig(config); err != nil {
		return nil, err
	}

	return config, nil
}

func (m *ModbusCLI) validateConfig(config *Config) error {
	// Validate count range
	if config.Count < 1 || config.Count > 125 {
		return fmt.Errorf("count must be between 1 and 125")
	}

	// Validate slave address range
	if config.SlaveID < 0 || config.SlaveID > 255 {
		return fmt.Errorf("slave address must be between 0 and 255")
	}

	// Validate baudrate range
	if config.Baudrate < 1200 || config.Baudrate > 921600 {
		return fmt.Errorf("baudrate must be between 1200 and 921600")
	}

	// Validate databits
	if config.Databits != 7 && config.Databits != 8 {
		return fmt.Errorf("databits must be 7 or 8")
	}

	// Validate stopbits
	if config.Stopbits != 1 && config.Stopbits != 2 {
		return fmt.Errorf("stopbits must be 1 or 2")
	}

	// Validate parity
	validParities := map[string]bool{
		"none": true,
		"even": true,
		"odd":  true,
	}
	if !validParities[config.Parity] {
		return fmt.Errorf("parity must be none, even, or odd")
	}

	// Validate poll rate
	if config.PollRate < 10*time.Millisecond {
		return fmt.Errorf("poll rate must be at least 10ms")
	}

	// Validate timeout
	if config.Timeout < 10*time.Millisecond || config.Timeout > 10*time.Second {
		return fmt.Errorf("timeout must be between 0.01 and 10.00 seconds")
	}

	return nil
}

func (m *ModbusCLI) setupClient() error {
	var url string
	var err error

	switch m.config.Mode {
	case "tcp":
		url = fmt.Sprintf("tcp://%s:%d", m.config.Host, m.config.Port)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     url,
			Timeout: m.config.Timeout,
		})
	case "tls":
		url = fmt.Sprintf("tls://%s:%d", m.config.Host, m.config.Port)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     url,
			Timeout: m.config.Timeout,
		})
	case "udp":
		url = fmt.Sprintf("udp://%s:%d", m.config.Host, m.config.Port)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     url,
			Timeout: m.config.Timeout,
		})
	case "rtu":
		url = fmt.Sprintf("rtu://%s", m.config.Device)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:      url,
			Speed:    uint(m.config.Baudrate),
			DataBits: uint(m.config.Databits),
			StopBits: uint(m.config.Stopbits),
			Timeout:  m.config.Timeout,
		})
		if m.config.Parity != "none" {
			switch m.config.Parity {
			case "even":
				// Default for RTU
			case "odd":
				// Set odd parity
			}
		}
	case "rtuovertcp":
		url = fmt.Sprintf("rtuovertcp://%s:%d", m.config.Host, m.config.Port)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     url,
			Speed:   uint(m.config.Baudrate),
			Timeout: m.config.Timeout,
		})
	case "rtuoverudp":
		url = fmt.Sprintf("rtuoverudp://%s:%d", m.config.Host, m.config.Port)
		m.client, err = modbus.NewClient(&modbus.ClientConfiguration{
			URL:     url,
			Speed:   uint(m.config.Baudrate),
			Timeout: m.config.Timeout,
		})
	default:
		return fmt.Errorf("unsupported mode: %s (supported: tcp, tls, udp, rtu, rtuovertcp, rtuoverudp)", m.config.Mode)
	}

	return err
}

func (m *ModbusCLI) connect() error {
	err := m.client.Open()
	if err != nil {
		return fmt.Errorf("failed to connect: %v", err)
	}

	m.client.SetUnitId(uint8(m.config.SlaveID))

	// Set encoding based on configuration
	endian := modbus.BIG_ENDIAN
	if !m.config.BigEndian {
		endian = modbus.LITTLE_ENDIAN
	}
	m.client.SetEncoding(endian, modbus.LOW_WORD_FIRST)

	return nil
}

func (m *ModbusCLI) execute() error {
	startRef := m.config.StartRef
	if m.config.ZeroBased {
		startRef = 0
	}

	if m.config.Verbose {
		m.printConfig()
	}

	// If write values are provided, perform write operation
	if len(m.config.WriteValues) > 0 {
		return m.performWriteOperation(startRef)
	}

	// Otherwise, perform read operation
	for {
		if err := m.performOperation(startRef); err != nil {
			return err
		}

		if m.config.PollOnce {
			break
		}

		time.Sleep(m.config.PollRate)
	}

	return nil
}

func (m *ModbusCLI) performOperation(startRef int) error {
	switch m.config.DataType {
	case "0":
		return m.readCoils(startRef)
	case "1":
		return m.readDiscreteInputs(startRef)
	case "3", "3:hex", "3:int", "3:float":
		return m.readInputRegisters(startRef)
	case "4", "4:hex", "4:int", "4:float":
		return m.readHoldingRegisters(startRef)
	default:
		return fmt.Errorf("unsupported data type: %s", m.config.DataType)
	}
}

func (m *ModbusCLI) performWriteOperation(startRef int) error {
	if len(m.config.WriteValues) == 0 {
		return fmt.Errorf("no write values provided")
	}

	switch m.config.DataType {
	case "0":
		return m.writeCoils(startRef)
	case "4", "4:hex", "4:int", "4:float":
		return m.writeHoldingRegisters(startRef)
	default:
		return fmt.Errorf("write operations not supported for data type: %s", m.config.DataType)
	}
}

func (m *ModbusCLI) readCoils(startRef int) error {
	coils, err := m.client.ReadCoils(uint16(startRef), uint16(m.config.Count))
	if err != nil {
		return fmt.Errorf("failed to read coils: %v", err)
	}

	fmt.Printf("Coils (%d-%d):\n", startRef, startRef+m.config.Count-1)
	for i, coil := range coils {
		fmt.Printf("[%d]: %d\n", startRef+i, boolToInt(coil))
	}

	return nil
}

func (m *ModbusCLI) readDiscreteInputs(startRef int) error {
	inputs, err := m.client.ReadDiscreteInputs(uint16(startRef), uint16(m.config.Count))
	if err != nil {
		return fmt.Errorf("failed to read discrete inputs: %v", err)
	}

	fmt.Printf("Discrete Inputs (%d-%d):\n", startRef, startRef+m.config.Count-1)
	for i, input := range inputs {
		fmt.Printf("[%d]: %d\n", startRef+i, boolToInt(input))
	}

	return nil
}

func (m *ModbusCLI) readInputRegisters(startRef int) error {
	registers, err := m.client.ReadRegisters(uint16(startRef), uint16(m.config.Count), modbus.INPUT_REGISTER)
	if err != nil {
		return fmt.Errorf("failed to read input registers: %v", err)
	}

	return m.printRegisters(startRef, registers, "Input Registers")
}

func (m *ModbusCLI) readHoldingRegisters(startRef int) error {
	registers, err := m.client.ReadRegisters(uint16(startRef), uint16(m.config.Count), modbus.HOLDING_REGISTER)
	if err != nil {
		return fmt.Errorf("failed to read holding registers: %v", err)
	}

	return m.printRegisters(startRef, registers, "Holding Registers")
}

func (m *ModbusCLI) writeCoils(startRef int) error {
	if len(m.config.WriteValues) == 0 {
		return fmt.Errorf("no values to write")
	}

	coils := make([]bool, len(m.config.WriteValues))
	for i, val := range m.config.WriteValues {
		coils[i] = val.(float64) != 0
	}

	err := m.client.WriteCoils(uint16(startRef), coils)
	if err != nil {
		return fmt.Errorf("failed to write coils: %v", err)
	}

	fmt.Printf("Successfully wrote %d coil(s) starting at address %d\n", len(coils), startRef)
	for i, coil := range coils {
		fmt.Printf("[%d]: %d\n", startRef+i, boolToInt(coil))
	}

	return nil
}

func (m *ModbusCLI) writeHoldingRegisters(startRef int) error {
	if len(m.config.WriteValues) == 0 {
		return fmt.Errorf("no values to write")
	}

	// Convert values based on data type
	switch m.config.DataType {
	case "4":
		// 16-bit registers
		registers := make([]uint16, len(m.config.WriteValues))
		for i, val := range m.config.WriteValues {
			registers[i] = uint16(val.(float64))
		}
		err := m.client.WriteRegisters(uint16(startRef), registers)
		if err != nil {
			return fmt.Errorf("failed to write holding registers: %v", err)
		}
		fmt.Printf("Successfully wrote %d 16-bit register(s) starting at address %d\n", len(registers), startRef)
		for i, reg := range registers {
			fmt.Printf("[%d]: %d\n", startRef+i, reg)
		}

	case "4:int":
		// 32-bit integers
		if len(m.config.WriteValues)%2 != 0 {
			return fmt.Errorf("32-bit integers require even number of values")
		}
		values := make([]uint32, len(m.config.WriteValues)/2)
		for i := 0; i < len(m.config.WriteValues); i += 2 {
			high := uint32(m.config.WriteValues[i].(float64))
			low := uint32(m.config.WriteValues[i+1].(float64))
			if m.config.BigEndian {
				values[i/2] = high<<16 | low
			} else {
				values[i/2] = low<<16 | high
			}
		}
		err := m.client.WriteUint32s(uint16(startRef), values)
		if err != nil {
			return fmt.Errorf("failed to write 32-bit integers: %v", err)
		}
		fmt.Printf("Successfully wrote %d 32-bit integer(s) starting at address %d\n", len(values), startRef)
		for i, val := range values {
			fmt.Printf("[%d]: %d\n", startRef+i*2, val)
		}

	case "4:float":
		// 32-bit floats
		if len(m.config.WriteValues)%2 != 0 {
			return fmt.Errorf("32-bit floats require even number of values")
		}
		values := make([]float32, len(m.config.WriteValues)/2)
		for i := 0; i < len(m.config.WriteValues); i += 2 {
			high := uint32(m.config.WriteValues[i].(float64))
			low := uint32(m.config.WriteValues[i+1].(float64))
			var bits uint32
			if m.config.BigEndian {
				bits = high<<16 | low
			} else {
				bits = low<<16 | high
			}
			values[i/2] = math.Float32frombits(bits)
		}
		err := m.client.WriteFloat32s(uint16(startRef), values)
		if err != nil {
			return fmt.Errorf("failed to write 32-bit floats: %v", err)
		}
		fmt.Printf("Successfully wrote %d 32-bit float(s) starting at address %d\n", len(values), startRef)
		for i, val := range values {
			fmt.Printf("[%d]: %.2f\n", startRef+i*2, val)
		}
	}

	return nil
}

func (m *ModbusCLI) printRegisters(startRef int, registers []uint16, regType string) error {
	fmt.Printf("%s (%d-%d):\n", regType, startRef, startRef+m.config.Count-1)

	for i, reg := range registers {
		addr := startRef + i
		fmt.Printf("[%d]: %d", addr, reg)

		// Handle different data type formats
		switch m.config.DataType {
		case "3:hex", "4:hex":
			fmt.Printf(" (0x%04X)", reg)
		case "3:int", "4:int":
			if i%2 == 0 && i+1 < len(registers) {
				// 32-bit integer
				var val int32
				if m.config.BigEndian {
					val = int32(registers[i])<<16 | int32(registers[i+1])
				} else {
					val = int32(registers[i+1])<<16 | int32(registers[i])
				}
				fmt.Printf(" (%d as 32-bit int)", val)
				i++ // Skip next register as it's part of this value
			}
		case "3:float", "4:float":
			if i%2 == 0 && i+1 < len(registers) {
				// 32-bit float
				var val float32
				if m.config.BigEndian {
					val = math.Float32frombits(uint32(registers[i])<<16 | uint32(registers[i+1]))
				} else {
					val = math.Float32frombits(uint32(registers[i+1])<<16 | uint32(registers[i]))
				}
				fmt.Printf(" (%.2f as 32-bit float)", val)
				i++ // Skip next register as it's part of this value
			}
		}

		fmt.Println()
	}

	return nil
}

func (m *ModbusCLI) printConfig() {
	fmt.Printf("Protocol configuration: Modbus %s\n", strings.ToUpper(m.config.Mode))
	fmt.Printf("Slave configuration...: address = [%d]\n", m.config.SlaveID)
	fmt.Printf("Data type.............: %s\n", m.config.DataType)
	fmt.Printf("Communication.........: %s, %d-%d%c%d\n",
		m.config.Device, m.config.Baudrate, m.config.Databits,
		m.getParityChar(), m.config.Stopbits)
	fmt.Printf("Timeout...............: %.2f s\n", m.config.Timeout.Seconds())
	fmt.Printf("Poll rate.............: %d ms\n", int(m.config.PollRate.Milliseconds()))
}

func (m *ModbusCLI) getParityChar() byte {
	switch m.config.Parity {
	case "none":
		return 'N'
	case "even":
		return 'E'
	case "odd":
		return 'O'
	default:
		return 'E'
	}
}

func (m *ModbusCLI) printHelp() {
	fmt.Println(`go-modbus-cli - Enhanced Modbus CLI tool

USAGE:
  go-modbus-cli [OPTIONS] DEVICE|HOST [WRITE_VALUES...] [OPTIONS]

ARGUMENTS:
  DEVICE        Serial port when using Modbus RTU protocol
                (e.g., /dev/ttyUSB0, COM1)
  HOST          Host name or IP address when using Modbus TCP protocol
  WRITE_VALUES  List of values to be written (if not specified, reads data)

GENERAL OPTIONS:
  -m, --mode MODE         Mode: tcp, tls, udp, rtu, rtuovertcp, rtuoverudp (default: tcp)
  -a, --address ADDR      Slave address (1-255, default: 1)
  -r, --reference REF     Start reference (default: 1)
  -c, --count COUNT       Number of values to read (1-125, default: 1)
  -t, --type TYPE         Data type:
                            0 = Discrete output (coil)
                            1 = Discrete input
                            3 = 16-bit input register
                            3:hex = 16-bit input register (hex display)
                            3:int = 32-bit integer in input register
                            3:float = 32-bit float in input register
                            4 = 16-bit output (holding) register (default)
                            4:hex = 16-bit output register (hex display)
                            4:int = 32-bit integer in output register
                            4:float = 32-bit float in output register
  -0, --zero-based        First reference is 0 (PDU addressing)
  -B, --big-endian        Big endian word order for 32-bit data (default)
  -1, --once              Poll only once, otherwise poll continuously
  -l, --poll-rate MS      Poll rate in milliseconds (default: 1000)
  -o, --timeout SEC       Timeout in seconds (default: 1.0)

TCP OPTIONS:
  -p, --port PORT         TCP port number (default: 502)

RTU OPTIONS:
  -b, --baudrate RATE     Baudrate (1200-921600, default: 19200)
  -d, --databits BITS     Databits (7 or 8, default: 8)
  -s, --stopbits BITS     Stopbits (1 or 2, default: 1)
  -P, --parity PARITY     Parity: none, even, odd (default: even)

OTHER OPTIONS:
  -v, --verbose           Verbose mode
  -h, --help              Show this help message
  -V, --version           Show version information

EXAMPLES:
  # Read 2 holding registers starting at address 1 from TCP device
  go-modbus-cli -t 4 -r 1 -c 2 192.168.1.100

  # Read input registers as 32-bit floats from RTU device
  go-modbus-cli -m rtu -t 3:float -r 1 -c 2 /dev/ttyUSB0

  # Write values to holding registers
  go-modbus-cli -t 4 -r 1 192.168.1.100 123 456 789

  # Write 32-bit integers to holding registers
  go-modbus-cli -t 4:int -r 1 192.168.1.100 123456 -789012

  # Write 32-bit floats to holding registers
  go-modbus-cli -t 4:float -r 1 192.168.1.100 3.14 -2.71

  # Write coils
  go-modbus-cli -t 0 -r 1 192.168.1.100 1 0 1 1

  # Poll coils continuously
  go-modbus-cli -t 0 -r 1 -c 8 -l 500 192.168.1.100

  # Use RTU over TCP (tunneled serial)
  go-modbus-cli -m rtuovertcp -t 4 -r 1 -c 2 192.168.1.100

  # Use Modbus TCP over UDP
  go-modbus-cli -m udp -t 4 -r 1 -c 2 192.168.1.100

  # Use Modbus TCP over TLS
  go-modbus-cli -m tls -t 4 -r 1 -c 2 192.168.1.100`)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
