//go:build !cgo || (!darwin && !linux) || (!amd64 && !arm64)

package dotenvx

import "errors"

func loadNative(loadOptions) (loadResult, error) {
	return loadResult{}, errors.New("dotenvx: requires cgo on macOS or glibc Linux (amd64 or arm64)")
}
