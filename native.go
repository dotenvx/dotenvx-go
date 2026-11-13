//go:build cgo && (darwin || linux) && (amd64 || arm64)

package dotenvx

/*
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/lib/darwin_amd64
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/lib/darwin_arm64
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/lib/linux_amd64
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/lib/linux_arm64
#cgo linux LDFLAGS: -ldotenvx_go_native -lgcc_s -lutil -lrt -lpthread -lm -ldl
#cgo darwin LDFLAGS: -ldotenvx_go_native -liconv
#include <stdlib.h>

char *dotenvx_go_load(const char *input);
void dotenvx_go_free(char *output);
*/
import "C"

import (
	"encoding/json"
	"errors"
	"unsafe"
)

func loadNative(options loadOptions) (loadResult, error) {
	var result loadResult
	input, err := json.Marshal(options)
	if err != nil {
		return result, err
	}

	request := C.CString(string(input))
	defer C.free(unsafe.Pointer(request))
	response := C.dotenvx_go_load(request)
	if response == nil {
		return result, errors.New("dotenvx: native loader failed")
	}
	defer C.dotenvx_go_free(response)

	if err := json.Unmarshal([]byte(C.GoString(response)), &result); err != nil {
		return result, err
	}
	if result.Error != "" {
		return result, errors.New(result.Error)
	}
	return result, nil
}
