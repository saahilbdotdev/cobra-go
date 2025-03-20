package cobrago

/*
#cgo CFLAGS: -I./cobra/include
#cgo LDFLAGS: -L./cobra/lib/linux/x86_64 -lpv_cobra
#include "./cobra/include/pv_cobra.h"
#include <stdlib.h>
*/
import "C"
import (
	"fmt"
	"os"
	"unsafe"
)

var accessKey string

func init() {
	accessKey = os.Getenv("COBRA_ACCESS_KEY")
	if accessKey == "" {
		panic("COBRA_ACCESS_KEY is not set")
	}
}

type CobraVAD struct {
	cobra     *C.pv_cobra_t
	Threshold float64
}

type CobraVADOption func(*CobraVAD)

func WithThreshold(threshold float64) CobraVADOption {
	return func(c *CobraVAD) {
		c.Threshold = threshold
	}
}

func New(opts ...CobraVADOption) *CobraVAD {
	var cobra *C.pv_cobra_t

	cobraVAD := &CobraVAD{cobra: cobra}

	for _, opt := range opts {
		opt(cobraVAD)
	}

	return cobraVAD
}

func (c *CobraVAD) Init() error {
	status := C.pv_cobra_init(C.CString(accessKey), &c.cobra)
	if status != C.PV_STATUS_SUCCESS {
		return fmt.Errorf("failed to initialize Cobra: %d", status)
	}

	return nil
}

func (c *CobraVAD) Close() {
	C.pv_cobra_delete(c.cobra)
}

func (c *CobraVAD) Process(frame []int16) (bool, error) {
	cPcm := (*C.int16_t)(unsafe.Pointer(&frame[0]))

	var isVoiced C.float

	status := C.pv_cobra_process(c.cobra, cPcm, &isVoiced)
	if status != C.PV_STATUS_SUCCESS {
		return false, fmt.Errorf("failed to process audio frame: %d", status)
	}

	return isVoiced > C.float(c.Threshold), nil
}
