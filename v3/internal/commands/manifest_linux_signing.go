package commands

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

// nfpmDebSignatureType maps linux.signing.identity (the dpkg-sig role) onto the
// nfpm deb signature type. dpkg-sig's default role "builder" has no nfpm
// equivalent; it is mapped to "origin", which is what dpkg-sig and apt verify
// as the package origin, and the mapping is documented in the migration guide.
func nfpmDebSignatureType(identity string) (string, error) {
	switch identity {
	case "", "builder", "origin":
		return "origin", nil
	case "maint", "archive":
		return identity, nil
	default:
		return "", fmt.Errorf("linux signing identity %q is not a deb signature role; use origin, maint, archive, or builder", identity)
	}
}

// applyNFPMSignature adds the nfpm signature section for a deb or rpm package.
// The passphrase is deliberately not part of the configuration; it is applied to
// the parsed package information in memory.
func applyNFPMSignature(config map[string]any, format, keyFile, identity string) error {
	signature := map[string]any{"key_file": keyFile}
	section := "rpm"
	switch format {
	case "deb":
		section = "deb"
		signatureType, err := nfpmDebSignatureType(identity)
		if err != nil {
			return err
		}
		signature["type"] = signatureType
	case "rpm":
	default:
		return fmt.Errorf("native linux signing does not support %q packages", format)
	}
	formatSection, _ := config[section].(map[string]any)
	if formatSection == nil {
		formatSection = map[string]any{}
	}
	formatSection["signature"] = signature
	config[section] = formatSection
	return nil
}

// linuxSigningKeyDigest identifies the signing key by content without placing
// key material in logs or cache keys. Changing the key changes the digest.
func linuxSigningKeyDigest(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("linux signing key: %w", err)
	}
	defer file.Close()
	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("linux signing key: %w", err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// debHasNativeSignature reports whether a deb archive carries a dpkg origin,
// maint or archive signature member, which is what nfpm writes.
func debHasNativeSignature(path string) (bool, error) {
	file, err := os.Open(path)
	if err != nil {
		return false, err
	}
	defer file.Close()
	magic := make([]byte, 8)
	if _, err := io.ReadFull(file, magic); err != nil || string(magic) != "!<arch>\n" {
		return false, errors.New("not a deb archive")
	}
	for {
		header := make([]byte, 60)
		if _, err := io.ReadFull(file, header); err != nil {
			if errors.Is(err, io.EOF) {
				return false, nil
			}
			return false, fmt.Errorf("read deb archive: %w", err)
		}
		name := strings.TrimRight(string(header[:16]), " ")
		switch name {
		case "_gpgorigin", "_gpgmaint", "_gpgarchive":
			return true, nil
		}
		var size int64
		if _, err := fmt.Sscanf(strings.TrimRight(string(header[48:58]), " "), "%d", &size); err != nil {
			return false, fmt.Errorf("read deb archive member size: %w", err)
		}
		if size%2 == 1 {
			size++
		}
		if _, err := file.Seek(size, io.SeekCurrent); err != nil {
			return false, err
		}
	}
}
