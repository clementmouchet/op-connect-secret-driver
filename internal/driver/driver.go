package driver

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

type requestedSecret struct {
	vault   string
	item    string
	section string
	field   string
}

type secretRequestParser struct {
	labels map[string]string
}

type itemSecretResolver struct {
	client connect.Client
}

// NewDriver creates a new instance of the driver.
func NewDriver() (*OPConnectSecretDriver, error) {
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

func (parser secretRequestParser) parse() (requestedSecret, bool, error) {
	if ref, ok := parser.labels["ref"]; ok {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] using op:// URL reference: %s\n", ref)
		vault, item, section, field, err := parseOpURL(ref)
		if err != nil {
			return requestedSecret{}, true, err
		}

		return requestedSecret{
			vault:   vault,
			item:    item,
			section: section,
			field:   field,
		}, true, nil
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] using individual labels\n")

	vault, ok := parser.labels["vault"]
	if !ok {
		return requestedSecret{}, false, fmt.Errorf("driver options must include \"vault\"")
	}

	item, ok := parser.labels["item"]
	if !ok {
		return requestedSecret{}, false, fmt.Errorf("driver options must include \"item\"")
	}

	field, ok := parser.labels["field"]
	if !ok || field == "" {
		field = "password"
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] no field specified, defaulting to 'password'\n")
	}

	section := strings.TrimSpace(parser.labels["section"])

	return requestedSecret{
		vault:   vault,
		item:    item,
		section: section,
		field:   field,
	}, false, nil
}

func (resolver itemSecretResolver) resolve(secret requestedSecret) secrets.Response {
	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] accessing vault: %s, item: %s, section: %s, field: %s\n",
		secret.vault, secret.item, secret.section, secret.field)

	itemDetails, err := resolver.client.GetItem(secret.item, secret.vault)
	if err != nil {
		errMsg := fmt.Sprintf("failed to get item '%s' from vault '%s': %v", secret.item, secret.vault, err)
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
		return secrets.Response{Err: errMsg}
	}

	_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved item '%s' (sections: %d, fields: %d, files: %d)\n",
		secret.item, len(itemDetails.Sections), len(itemDetails.Fields), len(itemDetails.Files))

	sectionID := findSectionID(secret.section, itemDetails.Sections)
	if secret.section != "" && sectionID == "" {
		errMsg := fmt.Sprintf("section '%s' not found in item '%s'", secret.section, secret.item)
		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
		return secrets.Response{Err: errMsg}
	}

	if file := findFile(secret.field, sectionID, itemDetails.Files); file != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] found file '%s' in item '%s', retrieving content\n", secret.field, secret.item)
		fileContent, err := resolver.client.GetFileContent(file)
		if err != nil {
			errMsg := fmt.Sprintf("error getting file '%s' content: %v", secret.field, err)
			_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
			return secrets.Response{Err: errMsg}
		}

		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved file content for '%s' (%d bytes)\n", secret.field, len(fileContent))
		return secrets.Response{Value: fileContent}
	}

	if fieldItem := findField(secret.field, sectionID, itemDetails.Fields); fieldItem != nil {
		_, _ = fmt.Fprintf(os.Stdout, "[OPCSD] successfully retrieved field '%s' from item '%s'\n", secret.field, secret.item)
		return secrets.Response{Value: []byte(fieldItem.Value)}
	}

	var errMsg string
	if secret.section != "" {
		errMsg = fmt.Sprintf("field '%s' in section '%s' not found in item '%s'", secret.field, secret.section, secret.item)
	} else {
		errMsg = fmt.Sprintf("field or file '%s' not found in item '%s'", secret.field, secret.item)
	}
	_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
	return secrets.Response{Err: errMsg}
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

	parser := secretRequestParser{labels: req.SecretLabels}
	secret, usedRef, err := parser.parse()
	if err != nil {
		errMsg := err.Error()
		if usedRef {
			errMsg = fmt.Sprintf("failed to parse 1Password URL: %v", err)
		}

		_, _ = fmt.Fprintf(os.Stderr, "[OPCSD] %s\n", errMsg)
		return secrets.Response{Err: errMsg}
	}

	resolver := itemSecretResolver{client: driver.client}
	return resolver.resolve(secret)
}
