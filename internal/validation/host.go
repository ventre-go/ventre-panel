// Package validation provides input validation helpers.
package validation

import (
	"fmt"
	"strconv"
	"strings"
)

// CSVErrorCode identifies a host CSV validation failure without including
// sensitive row content.
type CSVErrorCode string

const (
	CSVErrFields           CSVErrorCode = "fields"
	CSVErrIPRequired       CSVErrorCode = "ip_required"
	CSVErrPortNumber       CSVErrorCode = "port_number"
	CSVErrPortRange        CSVErrorCode = "port_range"
	CSVErrUsernameRequired CSVErrorCode = "username_required"
	CSVErrPasswordRequired CSVErrorCode = "password_required"
)

// CSVError reports a redacted CSV import validation error.
type CSVError struct {
	Code  CSVErrorCode
	Line  int
	Value string
	Got   int
	Port  int
}

func (e CSVError) Error() string {
	switch e.Code {
	case CSVErrFields:
		return fmt.Sprintf("line %d: expected 4 fields (ip,port,username,password), got %d", e.Line, e.Got)
	case CSVErrIPRequired:
		return fmt.Sprintf("line %d: IP must not be empty", e.Line)
	case CSVErrPortNumber:
		return fmt.Sprintf("line %d: invalid port %q: must be a number", e.Line, e.Value)
	case CSVErrPortRange:
		return fmt.Sprintf("line %d: port %d out of range (1-65535)", e.Line, e.Port)
	case CSVErrUsernameRequired:
		return fmt.Sprintf("line %d: username must not be empty", e.Line)
	case CSVErrPasswordRequired:
		return fmt.Sprintf("line %d: password must not be empty", e.Line)
	default:
		return fmt.Sprintf("line %d: invalid CSV row", e.Line)
	}
}

// ParsedHost holds a single parsed host line from CSV paste.
type ParsedHost struct {
	IP       string
	Port     int
	Username string
	Password string
	LineNum  int
}

// ParseHostsCSV parses multi-line CSV host input.
// Format: ip,port,username,password
// Empty lines are skipped. Port defaults to 22. IP, username, and password are required.
func ParseHostsCSV(input string) ([]ParsedHost, []error) {
	lines := strings.Split(input, "\n")
	var hosts []ParsedHost
	var errs []error

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		fields := splitCSVLine(trimmed)
		if len(fields) != 4 {
			errs = append(errs, CSVError{Code: CSVErrFields, Line: i + 1, Got: len(fields)})
			continue
		}

		ip := strings.TrimSpace(fields[0])
		if ip == "" {
			errs = append(errs, CSVError{Code: CSVErrIPRequired, Line: i + 1})
			continue
		}

		port := 22
		if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
			var err error
			port, err = strconv.Atoi(strings.TrimSpace(fields[1]))
			if err != nil {
				errs = append(errs, CSVError{Code: CSVErrPortNumber, Line: i + 1, Value: strings.TrimSpace(fields[1])})
				continue
			}
			if port < 1 || port > 65535 {
				errs = append(errs, CSVError{Code: CSVErrPortRange, Line: i + 1, Port: port})
				continue
			}
		}

		username := strings.TrimSpace(fields[2])
		if username == "" {
			errs = append(errs, CSVError{Code: CSVErrUsernameRequired, Line: i + 1})
			continue
		}

		password := ""
		if len(fields) > 3 {
			password = strings.TrimSpace(fields[3])
		}
		if password == "" {
			errs = append(errs, CSVError{Code: CSVErrPasswordRequired, Line: i + 1})
			continue
		}

		hosts = append(hosts, ParsedHost{
			IP:       ip,
			Port:     port,
			Username: username,
			Password: password,
			LineNum:  i + 1,
		})
	}

	return hosts, errs
}

func splitCSVLine(line string) []string {
	var fields []string
	var current strings.Builder
	inQuotes := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuotes = !inQuotes
		case c == ',' && !inQuotes:
			fields = append(fields, current.String())
			current.Reset()
		default:
			current.WriteByte(c)
		}
	}
	fields = append(fields, current.String())
	return fields
}

// ValidateHostTarget checks required fields for a single host.
func ValidateHostTarget(ip string, port int, username string, password string) error {
	if ip == "" {
		return fmt.Errorf("IP must not be empty")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	if username == "" {
		return fmt.Errorf("username must not be empty")
	}
	if password == "" {
		return fmt.Errorf("password must not be empty")
	}
	return nil
}

// ValidatePort checks if a port string is a valid port number.
func ValidatePort(s string) error {
	if s == "" {
		return nil // default 22 is fine
	}
	port, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("port must be a number")
	}
	if port < 1 || port > 65535 {
		return fmt.Errorf("port must be between 1 and 65535")
	}
	return nil
}
