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
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

var DEFAULT unsafe.Pointer
var NEXT unsafe.Pointer
var charsetTransformer transform.Transformer

func SetDefaultCharset(t transform.Transformer) {
	charsetTransformer = t
}
func init() {
	DEFAULT = C.rtdl_default()
	NEXT = C.rtdl_next()
	charsetTransformer = simplifiedchinese.GBK.NewDecoder()
}

type Mode C.int

const (
	RTLDNow    Mode = C.RTLD_NOW
	RTLDLazy   Mode = C.RTLD_LAZY
	RTLDGlobal Mode = C.RTLD_GLOBAL
	RTLDLocal  Mode = C.RTLD_LOCAL
)

func (m Mode) c() C.int { return C.int(m) }

func DlError() error {
	e := C.dlerror()
	if e == nil {
		return nil
	}
	r, _, err := transform.String(charsetTransformer, C.GoString(e))
	if err != nil {
		return errors.New(C.GoString(e))
	}
	return errors.New(r)
}

type Handle struct {
	c unsafe.Pointer
}

func DlOpen(file string, mod Mode) (*Handle, error) {
	f := mem.NewStr(file)
	defer f.Free()

	h := C.dlopen((*C.char)(f.ToC()), mod.c())
	if h == nil {
		return nil, DlError()
	}

	return &Handle{c: h}, nil
}
func (hd *Handle) Close() {
	if hd.c != nil {
		C.dlclose(hd.c)
	}
}
func (hd Handle) Symbol(name string) (unsafe.Pointer, error) {
	n := mem.NewStr(name)
	defer n.Free()

	p := C.dlsym(hd.c, (*C.char)(n.ToC()))
	if p == nil {
		return nil, DlError()
	}

	return p, nil
}

type DlInfo struct {
	Fname string
	Fbase unsafe.Pointer
	Sname string
	Saddr unsafe.Pointer
}

func DlAddr(addr unsafe.Pointer) (*DlInfo, error) {
	di := new(C.Dl_info)
	i := C.dladdr(addr, di)
	if i == 0 {
		return nil, errors.New("not found")
	}

	return &DlInfo{
		Fname: C.GoString(di.dli_fname),
		Fbase: di.dli_fbase,
		Sname: C.GoString(di.dli_sname),
		Saddr: di.dli_saddr,
	}, nil
}

func DlSym(p unsafe.Pointer, name string) (unsafe.Pointer, error) {
	n := mem.NewStr(name)
	defer n.Free()
	r := C.dlsym(p, (*C.char)(n.ToC()))
	if r == nil {
		return nil, DlError()
	}
	return r, nil
}
