package driver

import (
	"fmt"
	"os"

	"github.com/1Password/connect-sdk-go/onepassword"
)

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
