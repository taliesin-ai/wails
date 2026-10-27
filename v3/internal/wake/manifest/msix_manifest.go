package manifest

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// MSIXManifestPublisher returns the Publisher attribute of the Identity element
// in an AppxManifest.xml document, the 1-based line the element starts on, and
// whether the attribute was present. MSIX packages use this value as the
// package publisher, so it must match the signing certificate subject.
func MSIXManifestPublisher(data []byte) (publisher string, line int, found bool, err error) {
	decoder := xml.NewDecoder(bytes.NewReader(data))
	for {
		before := decoder.InputOffset()
		token, err := decoder.Token()
		if err == io.EOF {
			return "", 0, false, nil
		}
		if err != nil {
			return "", 0, false, err
		}
		start, ok := token.(xml.StartElement)
		if !ok || start.Name.Local != "Identity" {
			continue
		}
		for _, attr := range start.Attr {
			if attr.Name.Local == "Publisher" {
				// The decoder offset is before the leading whitespace, so locate the '<'.
				position := before + int64(bytes.IndexByte(data[before:], '<'))
				return attr.Value, 1 + bytes.Count(data[:position], []byte("\n")), true, nil
			}
		}
		return "", 0, false, nil
	}
}

// validateMSIXPublisherAgainstManifest rejects a package whose publisher setting
// disagrees with the Publisher of the user-owned AppxManifest.xml. The user
// manifest is packaged as written, so a disagreement would otherwise be silent.
func validateMSIXPublisherAgainstManifest(root, manifestField string, format PackageFormat) error {
	if format.Publisher == "" || format.Manifest == "" {
		return nil
	}
	field := manifestField + ".publisher"
	data, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(format.Manifest)))
	if err != nil {
		return fieldValidationCause(manifestField+".manifest", err, "cannot read the MSIX manifest")
	}
	publisher, line, found, err := MSIXManifestPublisher(data)
	if err != nil {
		return fieldValidationCause(manifestField+".manifest", err, "the MSIX manifest is not valid XML")
	}
	if !found || publisher == format.Publisher {
		return nil
	}
	return fieldValidationError(field, "packages.msix.publisher is %q but %s line %d sets Publisher=%q; the user-owned manifest is packaged as written, so set both to the same value (the signing certificate subject)",
		format.Publisher, format.Manifest, line, publisher)
}

// MSIXPublisherMatchesSubject reports whether an MSIX publisher and a
// certificate subject name the same distinguished name. Components are compared
// as a set so RDN order and surrounding spaces do not matter.
func MSIXPublisherMatchesSubject(publisher, subject string) bool {
	left, right := distinguishedNameParts(publisher), distinguishedNameParts(subject)
	if len(left) == 0 || len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

// distinguishedNameParts splits "CN=a, O=b" into sorted, normalised components.
func distinguishedNameParts(name string) []string {
	var parts []string
	for _, part := range strings.Split(name, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%s", strings.ToUpper(strings.TrimSpace(key)), strings.TrimSpace(value)))
	}
	sort.Strings(parts)
	return parts
}
