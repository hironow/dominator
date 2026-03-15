package session

import "os/exec"

// CheckBinary verifies that a binary is available in PATH.
// Returns nil if found, error otherwise.
func CheckBinary(name string) error {
	_, err := exec.LookPath(name)
	return err
}
