package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/docker/go-plugins-helpers/secrets"
)

// OPConnectSecretDriver is the struct that implements the Docker secrets.Driver interface.
type OPConnectSecretDriver struct {
	client connect.Client
}

// newDriver creates a new instance of the driver.
func newDriver() (*OPConnectSecretDriver, error) {
	return newDriverWithClientFactory(connect.NewClientFromEnvironment)
}

// newDriverWithClientFactory creates a new instance of the driver with a custom client factory.
func newDriverWithClientFactory(clientFactory func() (connect.Client, error)) (*OPConnectSecretDriver, error) {
	client, err := clientFactory()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to create 1Password Connect client %v\n", err)
		return nil, fmt.Errorf("[OPCSD] failed to create 1Password Connect client: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] plugin initialized\n")

	return &OPConnectSecretDriver{client: client}, nil
}

// parseOpURL parses a 1Password URL in the format "op://vault/item/field" or "op://vault/item/section/field"
func parseOpURL(url string) (vault, item, section, field string, err error) {
	if len(url) < 5 || url[:5] != "op://" {
		return "", "", "", "", fmt.Errorf("invalid 1Password URL format, must start with op://")
	}

	// Remove the op:// prefix and split into exactly 3 parts: vault, item, and everything else (field path)
	parts := strings.SplitN(url[5:], "/", 3)
	if len(parts) < 2 {
		return "", "", "", "", fmt.Errorf("invalid 1Password URL format, expected op://vault/item[/field] or op://vault/item/section/field")
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

	//fmt.Fprintf(os.Stdout, "[OPCSD] parsed URL - vault: %s, item: %s, section: %s, field: %s\n", vault, item, section, field)
	return vault, item, section, field, nil
}

// Get retrieves a secret from 1Password.
// The request format is expected to be JSON with "vault" and "item" + "field" or "ref" keys.
// Example Secret in a Compose file:
//
//	 secrets:
//		 db_password:
//		   driver: op-secret-driver
//		   labels:
//		     vault: "your-vault-uuid-or-name"
//		     item: "your-item-uuid-or-name"
//		     field: "password" # optional, defaults to "password"
//		     section: "section-name" # optional, only needed if field is in a section
//
//	 	db_password:
//		  driver: op-secret-driver
//		  labels:
//		    ref: "op://Test/Test Secret/section 1/password"
func (driver *OPConnectSecretDriver) Get(req secrets.Request) secrets.Response {
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] Getting secrets for req %s %v\n", req.SecretName, req.SecretLabels)

	var client = driver.client
	var vault, item, section, field string

	// Check if using op:// URL format
	if ref, ok := req.SecretLabels["ref"]; ok {
		var err error
		vault, item, section, field, err = parseOpURL(ref)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to parse 1Password URL: %v\n", err)
			return secrets.Response{Err: err.Error()}
		}
	} else {
		// Fall back to individual labels
		var ok bool
		vault, ok = req.SecretLabels["vault"]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] driver options must include \"vault\"\n")
			return secrets.Response{Err: `driver options must include "vault"`}
		}

		item, ok = req.SecretLabels["item"]
		if !ok {
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] driver options must include \"item\"\n")
			return secrets.Response{Err: `driver options must include "item"`}
		}

		field, ok = req.SecretLabels["field"]
		if !ok || field == "" {
			field = "password" // Default to "password" field
		}

		// Section is optional
		section = req.SecretLabels["section"]
		if ok {
			section = strings.TrimSpace(section)
		}
	}

	//fmt.Fprintf(os.Stdout, "[OPCSD] Accessing vault: %s, item: %s, section: %s, field: %s\n", vault, item, section, field)

	// Retrieve the item from the specified vault
	itemDetails, err := client.GetItem(item, vault)
	if err != nil {
		_ = fmt.Errorf("[OPCSD] failed to get item '%s' from vault '%s': %v", item, vault, err)
		return secrets.Response{Err: fmt.Sprintf("[OPCSD] failed to get item '%s' from vault '%s': %v", item, vault, err)}
	}

	// First, check if the field is a file
	for _, file := range itemDetails.Files {
		if file.Name == field {
			// If a section is specified, ensure it matches
			if section != "" {
				if file.Section != nil && file.Section.Label == section {
					_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] Found file '%s' in section '%s' in item '%s'\n", field, section, item)
					fileContent, err := client.GetFileContent(file)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to get file content: %v\n", err)
						return secrets.Response{Err: fmt.Sprintf("[OPCSD] error getting file '%s' content: %v", field, err)}
					}
					return secrets.Response{Value: fileContent}
				}
			} else {
				// No section specified, only return if file has no section
				if file.Section == nil {
					_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] Found file '%s' in item '%s'\n", field, item)
					fileContent, err := client.GetFileContent(file)
					if err != nil {
						_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to get file content: %v\n", err)
						return secrets.Response{Err: fmt.Sprintf("[OPCSD] error getting file '%s' content: %v", field, err)}
					}
					return secrets.Response{Value: fileContent}
				}
			}
		}
	}

	// If not a file, check fields
	for _, f := range itemDetails.Fields {
		// Match field label
		if f.Label == field {
			// If a section is specified, ensure it matches
			if section != "" {
				if f.Section != nil && f.Section.Label == section {
					_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] Found secret '%s' in section '%s' in item '%s'\n", field, section, item)
					return secrets.Response{Value: []byte(f.Value)}
				}
			} else {
				// No section specified, only return if field has no section
				if f.Section == nil {
					_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] Found secret '%s' in item '%s'\n", field, item)
					return secrets.Response{Value: []byte(f.Value)}
				}
			}
		}
	}

	if section != "" {
		return secrets.Response{Err: fmt.Sprintf("[OPCSD] field '%s' in section '%s' not found in item '%s'", field, section, item)}
	}
	return secrets.Response{Err: fmt.Sprintf("[OPCSD] field or file '%s' not found in item '%s'", field, item)}
}

func main() {
	driver, err := newDriver()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to create 1Password Connect Driver: %v\n", err)
		os.Exit(1)
	}

	handler := secrets.NewHandler(driver)
	if err := handler.ServeUnix("/run/docker/plugins/opcsd.sock", 0); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] error serving plugin: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] closed\n")
}
