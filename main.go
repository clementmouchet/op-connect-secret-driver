package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
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
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to create 1Password Connect client: %v\n", err)
		return nil, fmt.Errorf("[OPCSD] failed to create 1Password Connect client: %v", err)
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] plugin initialized\n")

	return &OPConnectSecretDriver{client: client}, nil
}

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

// Helper function to find section ID by label or ID
func findSectionID(sectionIdentifier string, sections []*onepassword.ItemSection) string {
	if sections == nil || sectionIdentifier == "" {
		return ""
	}

	for _, s := range sections {
		// Check if it matches the section ID directly
		if s.ID == sectionIdentifier {
			_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found section by ID: %s (label: %s)\n", s.ID, s.Label)
			return s.ID
		}

		// Check if it matches the section label
		if s.Label == sectionIdentifier {
			_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found section by label: %s (ID: %s)\n", s.Label, s.ID)
			return s.ID
		}
	}

	_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] section '%s' not found in item (available sections: %d)\n", sectionIdentifier, len(sections))
	return ""
}

// Helper function to find a field by label, ID and section ID
func findField(fieldLabel string, sectionID string, fields []*onepassword.ItemField) *onepassword.ItemField {
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] searching for field '%s' in section '%s' (total fields: %d)\n", fieldLabel, sectionID, len(fields))

	for _, f := range fields {
		// Check if it matches the field ID or label
		if f.ID != fieldLabel && f.Label != fieldLabel {
			continue
		}

		// Get the field's section ID
		fieldSectionID := ""
		if f.Section != nil {
			fieldSectionID = f.Section.ID
		}

		// Match: both have the same section ID (including both being empty)
		if fieldSectionID == sectionID {
			_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found field '%s' in section '%s'\n", fieldLabel, sectionID)
			return f
		}

		// Log when field name matches but section doesn't
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] field '%s' found but in different section (wanted: '%s', found: '%s')\n", fieldLabel, sectionID, fieldSectionID)
	}

	_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] field '%s' not found in section '%s'\n", fieldLabel, sectionID)
	return nil
}

// Helper function to find a file by name and section ID
func findFile(fileName string, sectionID string, files []*onepassword.File) *onepassword.File {
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] searching for file '%s' in section '%s' (total files: %d)\n", fileName, sectionID, len(files))

	for _, file := range files {
		if file.Name != fileName {
			continue
		}

		// Get the file's section ID
		fileSectionID := ""
		if file.Section != nil {
			fileSectionID = file.Section.ID
		}

		// Match: both have the same section ID (including both being empty)
		if fileSectionID == sectionID {
			_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found file '%s' in section '%s'\n", fileName, sectionID)
			return file
		}

		// Log when file name matches but section doesn't
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] file '%s' found but in different section (wanted: '%s', found: '%s')\n", fileName, sectionID, fileSectionID)
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] file '%s' not found in section '%s'\n", fileName, sectionID)
	return nil
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
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] using op:// URL reference: %s\n", ref)
		var err error
		vault, item, section, field, err = parseOpURL(ref)
		if err != nil {
			errMsg := fmt.Sprintf("failed to parse 1Password URL: %v", err)
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
			return secrets.Response{Err: errMsg}
		}
	} else {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] using individual labels\n")
		// Fall back to individual labels
		var ok bool
		vault, ok = req.SecretLabels["vault"]
		if !ok {
			errMsg := "driver options must include \"vault\""
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
			return secrets.Response{Err: errMsg}
		}

		item, ok = req.SecretLabels["item"]
		if !ok {
			errMsg := "driver options must include \"item\""
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
			return secrets.Response{Err: errMsg}
		}

		field, ok = req.SecretLabels["field"]
		if !ok || field == "" {
			field = "password" // Default to "password" field
			_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] no field specified, defaulting to 'password'\n")
		}

		// Section is optional
		section = req.SecretLabels["section"]
		if section != "" {
			section = strings.TrimSpace(section)
		}
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] accessing vault: %s, item: %s, section: %s, field: %s\n", vault, item, section, field)

	// Retrieve the item from the specified vault
	itemDetails, err := client.GetItem(item, vault)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get item '%s' from vault '%s': %v", item, vault, err)
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
		return secrets.Response{Err: errMsg}
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved item '%s' (sections: %d, fields: %d, files: %d)\n",
		item, len(itemDetails.Sections), len(itemDetails.Fields), len(itemDetails.Files))

	// Resolve section ID if section label is provided
	sectionID := findSectionID(section, itemDetails.Sections)
	if section != "" && sectionID == "" {
		errMsg := fmt.Sprintf("section '%s' not found in item '%s'", section, item)
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
		return secrets.Response{Err: errMsg}
	}

	// First, check if the field is a file
	if file := findFile(field, sectionID, itemDetails.Files); file != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found file '%s' in item '%s', retrieving content\n", field, item)
		fileContent, err := client.GetFileContent(file)
		if err != nil {
			errMsg := fmt.Sprintf("error getting file '%s' content: %v", field, err)
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
			return secrets.Response{Err: errMsg}
		}
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved file content for '%s' (%d bytes)\n", field, len(fileContent))
		return secrets.Response{Value: fileContent}
	}

	// If not a file, check fields
	if fieldItem := findField(field, sectionID, itemDetails.Fields); fieldItem != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved field '%s' from item '%s'\n", field, item)
		return secrets.Response{Value: []byte(fieldItem.Value)}
	}

	// Not found - generate appropriate error message
	var errMsg string
	if section != "" {
		errMsg = fmt.Sprintf("field '%s' in section '%s' not found in item '%s'", field, section, item)
	} else {
		errMsg = fmt.Sprintf("field or file '%s' not found in item '%s'", field, item)
	}
	_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
	return secrets.Response{Err: errMsg}
}

func main() {
	driver, err := newDriver()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] failed to create 1Password Connect Driver: %v\n", err)
		os.Exit(1)
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] starting plugin handler on /run/docker/plugins/opcsd.sock\n")
	handler := secrets.NewHandler(driver)
	if err := handler.ServeUnix("/run/docker/plugins/opcsd.sock", 0); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] error serving plugin: %v\n", err)
		os.Exit(1)
	}
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] closed\n")
}
