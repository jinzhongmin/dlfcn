package dlfcn

/*
#cgo windows CFLAGS:  -I../../3rdparty/windows/dlfcn/include
#cgo windows LDFLAGS: -L../../3rdparty/windows/dlfcn/lib -ldl
#include <dlfcn.h>
void *rtdl_default(){ return RTLD_DEFAULT; }
void *rtdl_next(){ return RTLD_NEXT; }
*/
import "C"
import (
	"errors"
	"unsafe"

	"github.com/jinzhongmin/mem"
)

var RTLD_DEFAULT unsafe.Pointer
var RTLD_NEXT unsafe.Pointer

func init() {
	RTLD_DEFAULT = C.rtdl_default()
	RTLD_NEXT = C.rtdl_next()
}

type Mode C.int

const (
	RTLDNow    Mode = C.RTLD_NOW
	RTLDLazy   Mode = C.RTLD_LAZY
	RTLDGlobal Mode = C.RTLD_GLOBAL
	RTLDLocal  Mode = C.RTLD_LOCAL
)

func (m Mode) toC() C.int { return C.int(m) }

func DlError() error {
	e := C.dlerror()
	if e == nil {
		return nil
	}
	return errors.New(C.GoString(e))
}

type Handle struct {
	c unsafe.Pointer
}

func DlOpen(file string, mod Mode) (*Handle, error) {
	f := C.CString(file)
	defer mem.Free(unsafe.Pointer(f))

	h := C.dlopen(f, mod.toC())
	if h == nil {
		return nil, DlError()
	}
	di := new(Handle)
	di.c = h
	return di, nil
}
func (hd *Handle) Close() {
	if hd.c != nil {
		C.dlclose(hd.c)
	}
}
func (hd Handle) Symbol(name string) (unsafe.Pointer, error) {
	n := C.CString(name)
	defer mem.Free(unsafe.Pointer(n))

	p := C.dlsym(hd.c, n)
	if p == nil {
		return nil, DlError()
	}

	return p, nil
}

//warp dladdr
func DlAddr(addr unsafe.Pointer) (int, string, unsafe.Pointer, string, unsafe.Pointer) {
	di := new(C.Dl_info)
	i := C.dladdr(addr, di)
	fname := C.GoString(di.dli_fname)
	fbase := di.dli_fbase
	sname := C.GoString(di.dli_sname)
	saddr := di.dli_saddr
	return int(int32(i)), fname, fbase, sname, saddr
}

//warp dlsym
func DlSym(p unsafe.Pointer, name string) (unsafe.Pointer, error) {
	n := C.CString(name)
	defer mem.Free(unsafe.Pointer(n))
	r := C.dlsym(p, n)
	if r == nil {
		return nil, DlError()
	}
	return r, nil
}
