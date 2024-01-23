package image

import (
	"fmt"
	"strings"
)

func wrapExecError(err error, output []byte) error {
	if err == nil || len(output) == 0 {
		return err
	}
	return fmt.Errorf("err: %s\nout: %s", err.Error(), string(output))
}

func isTypeSupported(filename string) bool {
	// could be global, but do we really need one more global variable?
	var supportedTypes = []string{".jpg", "jpeg", ".png"}

	lower := strings.ToLower(filename)
	for _, type_ := range supportedTypes {
		if strings.HasSuffix(lower, type_) {
			return true
		}
	}
	return false
}
