package api

/*
#include "bindings.h"

// typedefs for _cgo functions
typedef int64_t (*get_fn)(db_t *ptr, Buffer key, Buffer val);
typedef void (*set_fn)(db_t *ptr, Buffer key, Buffer val);

// forward declarations (db_cgo.go)
int64_t cGet_cgo(db_t *ptr, Buffer key, Buffer val);
void cSet_cgo(db_t *ptr, Buffer key, Buffer val);


// typedefs for _cgo functions
typedef int32_t (*human_address_fn)(api_t*, Buffer, Buffer);
typedef int32_t (*canonical_address_fn)(api_t*, Buffer, Buffer);

// forward declarations (api_cgo.go)
int32_t cHumanAddress_cgo(api_t *ptr, Buffer canon, Buffer human);
int32_t cCanonicalAddress_cgo(api_t *ptr, Buffer human, Buffer canon);
*/
import "C"

import "unsafe"

// Note: we have to include all exports in the same file (at least since they both import bindings.h),
// or get odd cgo build errors about duplicate definitions

/****** DB ********/

type KVStore interface {
	Get(key []byte) []byte
	Set(key, value []byte)
}

var db_vtable = C.DB_vtable{
	c_get: (C.get_fn)(C.cGet_cgo),
	c_set: (C.set_fn)(C.cSet_cgo),
}

func buildDB(kv KVStore) C.DB {
	return C.DB{
		state:  (*C.db_t)(unsafe.Pointer(&kv)),
		vtable: db_vtable,
	}
}

//export cGet
func cGet(ptr *C.db_t, key C.Buffer, val C.Buffer) i64 {
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)
	v := kv.Get(k)
	if len(v) == 0 {
		return 0
	}
	return writeToBuffer(val, v)
}

//export cSet
func cSet(ptr *C.db_t, key C.Buffer, val C.Buffer) {
	kv := *(*KVStore)(unsafe.Pointer(ptr))
	k := receiveSlice(key)
	v := receiveSlice(val)
	kv.Set(k, v)
}

/***** GoAPI *******/

type HumanAddress func([]byte) string
type CanonicalAddress func(string) []byte

type GoAPI struct {
    HumanAddress HumanAddress
    CanonicalAddress CanonicalAddress
}

var api_vtable = C.GoApi_vtable{
	c_human_address: (C.human_address_fn)(C.cHumanAddress_cgo),
	c_canonical_address: (C.canonical_address_fn)(C.cCanonicalAddress_cgo),
}

func buildAPI(api GoAPI) C.GoApi {
	return C.GoApi{
		state:  (*C.api_t)(unsafe.Pointer(&api)),
		vtable: api_vtable,
	}
}

//export cHumanAddress
func cHumanAddress(ptr *C.api_t, canon C.Buffer, human C.Buffer) i32 {
	api := (*GoAPI)(unsafe.Pointer(ptr))
	c := receiveSlice(canon)
	h := api.HumanAddress(c)
	if len(h) == 0 {
		return 0
	}
	return i32(writeToBuffer(human, []byte(h)))
}

//export cCanonicalAddress
func cCanonicalAddress(ptr *C.api_t, human C.Buffer, canon C.Buffer) i32 {
	api := (*GoAPI)(unsafe.Pointer(ptr))
	h := string(receiveSlice(human))
	c := api.CanonicalAddress(h)
	if len(c) == 0 {
		return 0
	}
	return i32(writeToBuffer(canon, c))
}