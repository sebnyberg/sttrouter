package cmd

import "fmt"

// Output format constants
const (
	textOutputFormatJSON  = "json"
	textOutputFormatCSV   = "csv"
	textOutputFormatTable = "table"
)

// formatOutput formats the given data into the specified text output format.
// It uses type assertion to check if the data implements the required formatting methods.
func formatOutput(format string, data any) (string, error) {
	switch format {
	case textOutputFormatJSON:
		if formatter, ok := data.(interface{ ToJSON() ([]byte, error) }); ok {
			bytes, err := formatter.ToJSON()
			if err != nil {
				return "", err
			}
			return string(bytes), nil
		}
		return "", fmt.Errorf("data does not support JSON formatting")
	case textOutputFormatTable:
		if formatter, ok := data.(interface{ ToTable() string }); ok {
			return formatter.ToTable(), nil
		}
		return "", fmt.Errorf("data does not support table formatting")
	case textOutputFormatCSV:
		if formatter, ok := data.(interface{ ToCSV() string }); ok {
			return formatter.ToCSV(), nil
		}
		return "", fmt.Errorf("data does not support CSV formatting")
	default:
		return "", fmt.Errorf("unsupported format: %s", format)
	}
}
