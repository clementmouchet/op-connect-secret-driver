package driver

import (
	"fmt"
	"os"
	"strings"
)

// parseOpURL parses a 1Password URL in the format "op://vault/item/field" or "op://vault/item/section/field"
func parseOpURL(url string) (vault, item, section, field string, err error) {
	if len(url) < 5 || url[:5] != "op://" {
		err = fmt.Errorf("invalid 1Password URL format, must start with op://")
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] parseOpURL error: %v (input: %s)\n", err, url)
		return "", "", "", "", err
	}

	// Remove the op:// prefix and split into exactly 3 parts: vault, item, and everything else (field path)
	parts := strings.SplitN(url[5:], "/", 3)
	if len(parts) < 2 {
		err = fmt.Errorf("invalid 1Password URL format, expected op://vault/item[/field] or op://vault/item/section/field")
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] parseOpURL error: %v (input: %s)\n", err, url)
		return "", "", "", "", err
	}

	vault = strings.TrimSpace(parts[0])
	item = strings.TrimSpace(parts[1])

	if len(parts) == 3 {
		fieldPath := strings.TrimSpace(parts[2])
		// Parse fieldPath to extract section and field if it contains a separator
		if strings.Contains(fieldPath, "/") {
			fieldParts := strings.SplitN(fieldPath, "/", 2)
			section = strings.TrimSpace(fieldParts[0])
			field = strings.TrimSpace(fieldParts[1])
		} else {
			field = fieldPath
		}
	} else {
		field = "password" // Default field if not specified
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] parsed URL - vault: %s, item: %s, section: %s, field: %s\n", vault, item, section, field)
	return vault, item, section, field, nil
}
