package main

import (
	"fmt"
	"testing"

	"github.com/1Password/connect-sdk-go/connect"
	"github.com/1Password/connect-sdk-go/onepassword"
	"github.com/docker/go-plugins-helpers/secrets"
)

// MockClient is a mock implementation of the connect.Client interface for testing
type MockClient struct {
	GetItemFunc        func(uuid, vaultUUID string) (*onepassword.Item, error)
	GetFileContentFunc func(file *onepassword.File) ([]byte, error)
}

func (m *MockClient) GetVaultByUUID(uuid string) (*onepassword.Vault, error) {
	return nil, nil
}
func (m *MockClient) GetVaultByTitle(title string) (*onepassword.Vault, error) {
	return nil, nil
}
func (m *MockClient) GetItemByUUID(uuid string, vaultQuery string) (*onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) DeleteItemByTitle(title string, vaultQuery string) error {
	return nil
}
func (m *MockClient) DownloadFile(file *onepassword.File, targetDirectory string, overwrite bool) (string, error) {
	return "", nil
}
func (m *MockClient) LoadStruct(config interface{}) error {
	return nil
}
func (m *MockClient) GetItem(uuid, vaultUUID string) (*onepassword.Item, error) {
	if m.GetItemFunc != nil {
		return m.GetItemFunc(uuid, vaultUUID)
	}
	return nil, nil
}
func (m *MockClient) GetFileContent(file *onepassword.File) ([]byte, error) {
	if m.GetFileContentFunc != nil {
		return m.GetFileContentFunc(file)
	}
	return nil, nil
}
func (m *MockClient) GetVaults() ([]onepassword.Vault, error) {
	return nil, nil
}
func (m *MockClient) GetVault(uuid string) (*onepassword.Vault, error) {
	return nil, nil
}
func (m *MockClient) GetVaultsByTitle(title string) ([]onepassword.Vault, error) {
	return nil, nil
}
func (m *MockClient) GetItems(vaultUUID string) ([]onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) GetItemsByTitle(title, vaultUUID string) ([]onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) GetItemByTitle(title, vaultUUID string) (*onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) CreateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) UpdateItem(item *onepassword.Item, vaultUUID string) (*onepassword.Item, error) {
	return nil, nil
}
func (m *MockClient) DeleteItem(item *onepassword.Item, vaultUUID string) error {
	return nil
}
func (m *MockClient) DeleteItemByID(itemUUID, vaultUUID string) error {
	return nil
}
func (m *MockClient) LoadStructFromItemByTitle(config interface{}, title, vaultUUID string) error {
	return nil
}
func (m *MockClient) LoadStructFromItem(config interface{}, itemUUID, vaultUUID string) error {
	return nil
}
func (m *MockClient) LoadStructFromItemByUUID(config interface{}, itemUUID, vaultUUID string) error {
	return nil
}
func (m *MockClient) GetFiles(itemUUID, vaultUUID string) ([]onepassword.File, error) {
	return nil, nil
}
func (m *MockClient) GetFile(fileUUID, itemUUID, vaultUUID string) (*onepassword.File, error) {
	return nil, nil
}

func TestNewDriverWithClientFactory(t *testing.T) {
	tests := []struct {
		name          string
		clientFactory func() (connect.Client, error)
		wantErr       bool
	}{
		{
			name: "successful driver creation",
			clientFactory: func() (connect.Client, error) {
				return &MockClient{}, nil
			},
			wantErr: false,
		},
		{
			name: "client creation fails",
			clientFactory: func() (connect.Client, error) {
				return nil, fmt.Errorf("failed to create client")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			driver, err := newDriverWithClientFactory(tt.clientFactory)
			if tt.wantErr {
				if err == nil {
					t.Errorf("newDriverWithClientFactory() expected error but got none")
				}
				if driver != nil {
					t.Errorf("newDriverWithClientFactory() expected nil driver on error, got %v", driver)
				}
			} else {
				if err != nil {
					t.Errorf("newDriverWithClientFactory() unexpected error = %v", err)
				}
				if driver == nil {
					t.Errorf("newDriverWithClientFactory() expected driver, got nil")
				}
				if driver.client == nil {
					t.Errorf("newDriverWithClientFactory() driver.client is nil")
				}
			}
		})
	}
}

func TestParseOpURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		wantVault   string
		wantItem    string
		wantSection string
		wantField   string
		wantErr     bool
	}{
		{
			name:        "valid URL with section and field",
			url:         "op://vault/item/section/field",
			wantVault:   "vault",
			wantItem:    "item",
			wantSection: "section",
			wantField:   "field",
			wantErr:     false,
		},
		{
			name:        "valid URL with field only",
			url:         "op://vault/item/field",
			wantVault:   "vault",
			wantItem:    "item",
			wantSection: "",
			wantField:   "field",
			wantErr:     false,
		},
		{
			name:        "valid URL without field (defaults to password)",
			url:         "op://vault/item",
			wantVault:   "vault",
			wantItem:    "item",
			wantSection: "",
			wantField:   "password",
			wantErr:     false,
		},
		{
			name:        "URL with spaces",
			url:         "op://vault name/item name/section name/field name",
			wantVault:   "vault name",
			wantItem:    "item name",
			wantSection: "section name",
			wantField:   "field name",
			wantErr:     false,
		},
		{
			name:        "invalid URL without op:// prefix",
			url:         "vault/item/field",
			wantVault:   "",
			wantItem:    "",
			wantSection: "",
			wantField:   "",
			wantErr:     true,
		},
		{
			name:        "invalid URL with only vault",
			url:         "op://vault",
			wantVault:   "",
			wantItem:    "",
			wantSection: "",
			wantField:   "",
			wantErr:     true,
		},
		{
			name:        "empty URL",
			url:         "",
			wantVault:   "",
			wantItem:    "",
			wantSection: "",
			wantField:   "",
			wantErr:     true,
		},
		{
			name:        "URL with trailing slashes",
			url:         "op://vault/item/field/",
			wantVault:   "vault",
			wantItem:    "item",
			wantSection: "field",
			wantField:   "",
			wantErr:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vault, item, section, field, err := parseOpURL(tt.url)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseOpURL() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if vault != tt.wantVault {
				t.Errorf("parseOpURL() vault = %v, want %v", vault, tt.wantVault)
			}
			if item != tt.wantItem {
				t.Errorf("parseOpURL() item = %v, want %v", item, tt.wantItem)
			}
			if section != tt.wantSection {
				t.Errorf("parseOpURL() section = %v, want %v", section, tt.wantSection)
			}
			if field != tt.wantField {
				t.Errorf("parseOpURL() field = %v, want %v", field, tt.wantField)
			}
		})
	}
}

func TestOPConnectSecretDriver_Get(t *testing.T) {
	tests := []struct {
		name      string
		request   secrets.Request
		mockItem  *onepassword.Item
		mockFile  []byte
		wantErr   bool
		wantValue string
		setupMock func(*MockClient)
	}{
		{
			name: "get field with ref URL",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"ref": "op://test-vault/test-item/test-field",
				},
			},
			mockItem: &onepassword.Item{
				Fields: []*onepassword.ItemField{
					{
						Label: "test-field",
						Value: "secret-value",
					},
				},
			},
			wantErr:   false,
			wantValue: "secret-value",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "test-field",
								Value: "secret-value",
							},
						},
					}, nil
				}
			},
		},
		{
			name: "get field with section using ref URL",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"ref": "op://test-vault/test-item/test-section/test-field",
				},
			},
			wantErr:   false,
			wantValue: "secret-value-in-section",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "test-field",
								Value: "secret-value-in-section",
								Section: &onepassword.ItemSection{
									Label: "test-section",
								},
							},
						},
					}, nil
				}
			},
		},
		{
			name: "get field with individual labels",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
					"field": "test-field",
				},
			},
			wantErr:   false,
			wantValue: "secret-value",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "test-field",
								Value: "secret-value",
							},
						},
					}, nil
				}
			},
		},
		{
			name: "get field with default password",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
				},
			},
			wantErr:   false,
			wantValue: "default-password",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "password",
								Value: "default-password",
							},
						},
					}, nil
				}
			},
		},
		{
			name: "missing vault label",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"item": "test-item",
				},
			},
			wantErr: true,
		},
		{
			name: "missing item label",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
				},
			},
			wantErr: true,
		},
		{
			name: "invalid ref URL",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"ref": "invalid-url",
				},
			},
			wantErr: true,
		},
		{
			name: "get file content",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
					"field": "test-file",
				},
			},
			wantErr:   false,
			wantValue: "file-content",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Files: []*onepassword.File{
							{
								Name: "test-file",
							},
						},
					}, nil
				}
				m.GetFileContentFunc = func(file *onepassword.File) ([]byte, error) {
					return []byte("file-content"), nil
				}
			},
		},
		{
			name: "field not found",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
					"field": "non-existent-field",
				},
			},
			wantErr: true,
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "other-field",
								Value: "other-value",
							},
						},
					}, nil
				}
			},
		},
		{
			name: "get file with section",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault":   "test-vault",
					"item":    "test-item",
					"field":   "test-file",
					"section": "test-section",
				},
			},
			wantErr:   false,
			wantValue: "file-content-in-section",
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Files: []*onepassword.File{
							{
								Name: "test-file",
								Section: &onepassword.ItemSection{
									Label: "test-section",
								},
							},
						},
					}, nil
				}
				m.GetFileContentFunc = func(file *onepassword.File) ([]byte, error) {
					return []byte("file-content-in-section"), nil
				}
			},
		},
		{
			name: "section not found error",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault":   "test-vault",
					"item":    "test-item",
					"field":   "test-field",
					"section": "non-existent-section",
				},
			},
			wantErr: true,
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Fields: []*onepassword.ItemField{
							{
								Label: "test-field",
								Value: "test-value",
							},
						},
					}, nil
				}
			},
		},
		{
			name: "get item error",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
					"field": "test-field",
				},
			},
			wantErr: true,
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return nil, fmt.Errorf("item not found")
				}
			},
		},
		{
			name: "get file content error without section",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault": "test-vault",
					"item":  "test-item",
					"field": "test-file",
				},
			},
			wantErr: true,
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Files: []*onepassword.File{
							{
								Name: "test-file",
							},
						},
					}, nil
				}
				m.GetFileContentFunc = func(file *onepassword.File) ([]byte, error) {
					return nil, fmt.Errorf("file content error")
				}
			},
		},
		{
			name: "get file content error with section",
			request: secrets.Request{
				SecretName: "test-secret",
				SecretLabels: map[string]string{
					"vault":   "test-vault",
					"item":    "test-item",
					"field":   "test-file",
					"section": "test-section",
				},
			},
			wantErr: true,
			setupMock: func(m *MockClient) {
				m.GetItemFunc = func(uuid, vaultUUID string) (*onepassword.Item, error) {
					return &onepassword.Item{
						Files: []*onepassword.File{
							{
								Name: "test-file",
								Section: &onepassword.ItemSection{
									Label: "test-section",
								},
							},
						},
					}, nil
				}
				m.GetFileContentFunc = func(file *onepassword.File) ([]byte, error) {
					return nil, fmt.Errorf("file content error")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockClient := &MockClient{}
			if tt.setupMock != nil {
				tt.setupMock(mockClient)
			}

			driver := &OPConnectSecretDriver{
				client: mockClient,
			}

			response := driver.Get(tt.request)

			if tt.wantErr {
				if response.Err == "" {
					t.Errorf("Get() expected error but got none")
				}
			} else {
				if response.Err != "" {
					t.Errorf("Get() unexpected error = %v", response.Err)
				}
				if string(response.Value) != tt.wantValue {
					t.Errorf("Get() value = %v, want %v", string(response.Value), tt.wantValue)
				}
			}
		})
	}
}
