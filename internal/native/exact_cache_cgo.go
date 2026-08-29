//go:build cgo && ((darwin && (amd64 || arm64)) || (linux && (amd64 || arm64) && (sidereon_use_system_lib || (sidereon_linux_glibc && !sidereon_linux_musl) || (sidereon_linux_musl && !sidereon_linux_glibc))) || (windows && amd64))

package native

/*
#cgo CFLAGS: -I${SRCDIR}/include
#include <sidereon.h>
#include <stdlib.h>
*/
import "C"

import (
	"errors"
	"runtime"
	"unsafe"
)

const ExactCacheControlDirectory = ".sidereon-cache-v3"

type ExactCacheSingleFlightOptions struct {
	PollIntervalMS      uint64
	HeartbeatIntervalMS uint64
	LivenessTimeoutMS   uint64
	WaitTimeoutMS       uint64
}

type ExactCacheFiles struct {
	ProductPath    string
	ArchivePath    string
	ProvenancePath string
	EntryID        string
	Product        []byte
	Archive        []byte
	Provenance     []byte
}

type ExactCache struct {
	handle *positioningHandle
}

type ExactCacheEntry struct {
	handle *positioningHandle
}

type ExactCacheOwner struct {
	handle *positioningHandle
}

func exactCacheFromPointer(pointer *C.SidereonExactCache) (*ExactCache, error) {
	if pointer == nil {
		return nil, missingNativeHandle("exact-cache open")
	}
	return &ExactCache{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_exact_cache_free((*C.SidereonExactCache)(value))
	})}, nil
}

func exactCacheEntryFromPointer(pointer *C.SidereonExactCacheEntry) (*ExactCacheEntry, error) {
	if pointer == nil {
		return nil, missingNativeHandle("exact-cache entry")
	}
	return &ExactCacheEntry{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_exact_cache_entry_free((*C.SidereonExactCacheEntry)(value))
	})}, nil
}

func exactCacheOwnerFromPointer(pointer *C.SidereonExactCacheOwner) (*ExactCacheOwner, error) {
	if pointer == nil {
		return nil, missingNativeHandle("exact-cache single-flight owner")
	}
	return &ExactCacheOwner{handle: newPositioningHandle(unsafe.Pointer(pointer), func(value unsafe.Pointer) {
		C.sidereon_exact_cache_owner_free((*C.SidereonExactCacheOwner)(value))
	})}, nil
}

func ExactCacheSingleFlightOptionsInit() (ExactCacheSingleFlightOptions, error) {
	var value C.SidereonExactCacheSingleFlightOptions
	err := callStatus(func() uint32 {
		return uint32(C.sidereon_exact_cache_single_flight_options_init(&value))
	})
	return ExactCacheSingleFlightOptions{
		PollIntervalMS:      uint64(value.poll_interval_ms),
		HeartbeatIntervalMS: uint64(value.heartbeat_interval_ms),
		LivenessTimeoutMS:   uint64(value.liveness_timeout_ms),
		WaitTimeoutMS:       uint64(value.wait_timeout_ms),
	}, err
}

func ExactCacheOpen(stablePath string, identity ProductIdentity, source DistributionSource, timeoutMS uint64) (*ExactCache, error) {
	cIdentity, err := cProductIdentity(identity)
	if err != nil {
		return nil, err
	}
	var pointer *C.SidereonExactCache
	err = withStringError(stablePath, func(path *C.char) error {
		return statusErrorLocked(uint32(C.sidereon_exact_cache_open(path, &cIdentity, C.uint32_t(source), C.uint64_t(timeoutMS), &pointer)))
	})
	runtime.KeepAlive(identity)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_free(pointer) })
		}
		return nil, err
	}
	return exactCacheFromPointer(pointer)
}

func ExactCacheOpenSingleFlight(stablePath string, identity ProductIdentity, source DistributionSource, options *ExactCacheSingleFlightOptions) (*ExactCacheEntry, *ExactCacheOwner, error) {
	cIdentity, err := cProductIdentity(identity)
	if err != nil {
		return nil, nil, err
	}
	var result C.enum_SidereonExactCacheOpenResult
	var entryPointer *C.SidereonExactCacheEntry
	var ownerPointer *C.SidereonExactCacheOwner
	err = withStringError(stablePath, func(path *C.char) error {
		var cOptions *C.SidereonExactCacheSingleFlightOptions
		var initialized C.SidereonExactCacheSingleFlightOptions
		if options != nil {
			if initErr := statusErrorLocked(uint32(C.sidereon_exact_cache_single_flight_options_init(&initialized))); initErr != nil {
				return initErr
			}
			initialized.poll_interval_ms = C.uint64_t(options.PollIntervalMS)
			initialized.heartbeat_interval_ms = C.uint64_t(options.HeartbeatIntervalMS)
			initialized.liveness_timeout_ms = C.uint64_t(options.LivenessTimeoutMS)
			initialized.wait_timeout_ms = C.uint64_t(options.WaitTimeoutMS)
			cOptions = &initialized
		}
		return statusErrorLocked(uint32(C.sidereon_exact_cache_open_single_flight(
			path, &cIdentity, C.uint32_t(source), cOptions, &result, &entryPointer, &ownerPointer,
		)))
	})
	runtime.KeepAlive(identity)
	if err != nil {
		withCThread(func() {
			C.sidereon_exact_cache_entry_free(entryPointer)
			C.sidereon_exact_cache_owner_free(ownerPointer)
		})
		return nil, nil, err
	}
	switch result {
	case C.SIDEREON_EXACT_CACHE_OPEN_RESULT_HIT:
		if entryPointer == nil || ownerPointer != nil {
			withCThread(func() {
				C.sidereon_exact_cache_entry_free(entryPointer)
				C.sidereon_exact_cache_owner_free(ownerPointer)
			})
			return nil, nil, errors.New("sidereon: native exact-cache hit returned an invalid handle combination")
		}
		entry, entryErr := exactCacheEntryFromPointer(entryPointer)
		return entry, nil, entryErr
	case C.SIDEREON_EXACT_CACHE_OPEN_RESULT_OWNER:
		if entryPointer != nil || ownerPointer == nil {
			withCThread(func() {
				C.sidereon_exact_cache_entry_free(entryPointer)
				C.sidereon_exact_cache_owner_free(ownerPointer)
			})
			return nil, nil, errors.New("sidereon: native exact-cache owner result returned an invalid handle combination")
		}
		owner, ownerErr := exactCacheOwnerFromPointer(ownerPointer)
		return nil, owner, ownerErr
	default:
		withCThread(func() {
			C.sidereon_exact_cache_entry_free(entryPointer)
			C.sidereon_exact_cache_owner_free(ownerPointer)
		})
		return nil, nil, errors.New("sidereon: native exact-cache open returned an invalid result")
	}
}

func ExactCacheReadUnlocked(stablePath string, identity ProductIdentity, source DistributionSource) (*ExactCacheEntry, bool, error) {
	cIdentity, err := cProductIdentity(identity)
	if err != nil {
		return nil, false, err
	}
	var hit C.bool
	var pointer *C.SidereonExactCacheEntry
	err = withStringError(stablePath, func(path *C.char) error {
		return statusErrorLocked(uint32(C.sidereon_exact_cache_read_unlocked(path, &cIdentity, C.uint32_t(source), &hit, &pointer)))
	})
	runtime.KeepAlive(identity)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
		}
		return nil, false, err
	}
	if !bool(hit) {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
			return nil, false, errors.New("sidereon: native exact-cache miss returned an entry handle")
		}
		return nil, false, nil
	}
	entry, err := exactCacheEntryFromPointer(pointer)
	return entry, true, err
}

func (cache *ExactCache) Close() error {
	if cache == nil || cache.handle == nil {
		return nil
	}
	return cache.handle.close()
}

func (cache *ExactCache) Cleanup() error {
	if cache == nil || cache.handle == nil {
		return ErrClosed
	}
	err := cache.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_cache_cleanup((*C.SidereonExactCache)(pointer)))
		})
	})
	runtime.KeepAlive(cache)
	return err
}

func (cache *ExactCache) Read() (*ExactCacheEntry, bool, error) {
	if cache == nil || cache.handle == nil {
		return nil, false, ErrClosed
	}
	var hit C.bool
	var pointer *C.SidereonExactCacheEntry
	err := cache.handle.withExclusive(func(cachePointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_cache_read((*C.SidereonExactCache)(cachePointer), &hit, &pointer))
		})
	})
	runtime.KeepAlive(cache)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
		}
		return nil, false, err
	}
	if !bool(hit) {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
			return nil, false, errors.New("sidereon: native exact-cache miss returned an entry handle")
		}
		return nil, false, nil
	}
	entry, err := exactCacheEntryFromPointer(pointer)
	return entry, true, err
}

type exactCacheInput struct {
	pointer *C.uint8_t
	length  C.size_t
}

func copyExactCacheInputs(values ...[]byte) ([]exactCacheInput, func(), error) {
	inputs := make([]exactCacheInput, len(values))
	cleanup := func() {
		for _, input := range inputs {
			C.free(unsafe.Pointer(input.pointer))
		}
	}
	for i, value := range values {
		length, err := checkedNativeSize(len(value))
		if err != nil {
			cleanup()
			return nil, func() {}, err
		}
		inputs[i].length = length
		if len(value) == 0 {
			continue
		}
		pointer := C.CBytes(value)
		if pointer == nil {
			cleanup()
			return nil, func() {}, errors.New("sidereon: unable to allocate exact-cache input")
		}
		inputs[i].pointer = (*C.uint8_t)(pointer)
	}
	return inputs, cleanup, nil
}

func (cache *ExactCache) Publish(product, archive, provenance []byte) (*ExactCacheEntry, error) {
	if cache == nil || cache.handle == nil {
		return nil, ErrClosed
	}
	inputs, cleanup, err := copyExactCacheInputs(product, archive, provenance)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	var pointer *C.SidereonExactCacheEntry
	err = cache.handle.withExclusive(func(cachePointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_cache_publish(
				(*C.SidereonExactCache)(cachePointer),
				inputs[0].pointer, inputs[0].length,
				inputs[1].pointer, inputs[1].length,
				inputs[2].pointer, inputs[2].length,
				&pointer,
			))
		})
	})
	runtime.KeepAlive(cache)
	runtime.KeepAlive(product)
	runtime.KeepAlive(archive)
	runtime.KeepAlive(provenance)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
		}
		return nil, err
	}
	return exactCacheEntryFromPointer(pointer)
}

func (entry *ExactCacheEntry) Close() error {
	if entry == nil || entry.handle == nil {
		return nil
	}
	return entry.handle.close()
}

func (entry *ExactCacheEntry) Files() (ExactCacheFiles, error) {
	if entry == nil || entry.handle == nil {
		return ExactCacheFiles{}, ErrClosed
	}
	var files ExactCacheFiles
	err := entry.handle.with(func(pointer unsafe.Pointer) error {
		return withCThreadError(func() error {
			value := (*C.SidereonExactCacheEntry)(pointer)
			var err error
			files.Product, err = copyNativeBytesLocked("exact-cache product", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_bytes(value, C.SIDEREON_EXACT_CACHE_COMPONENT_PRODUCT, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.Archive, err = copyNativeBytesLocked("exact-cache archive", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_bytes(value, C.SIDEREON_EXACT_CACHE_COMPONENT_ARCHIVE, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.Provenance, err = copyNativeBytesLocked("exact-cache provenance", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_bytes(value, C.SIDEREON_EXACT_CACHE_COMPONENT_PROVENANCE, out, length, written, required)
			})
			if err != nil {
				return err
			}
			id, err := copyNativeBytesLocked("exact-cache entry id", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_id(value, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.EntryID = string(id)
			productPath, err := copyNativeBytesLocked("exact-cache product path", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_path(value, C.SIDEREON_EXACT_CACHE_COMPONENT_PRODUCT, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.ProductPath = string(productPath)
			archivePath, err := copyNativeBytesLocked("exact-cache archive path", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_path(value, C.SIDEREON_EXACT_CACHE_COMPONENT_ARCHIVE, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.ArchivePath = string(archivePath)
			provenancePath, err := copyNativeBytesLocked("exact-cache provenance path", func(out *C.uint8_t, length C.size_t, written, required *C.size_t) C.enum_SidereonStatus {
				return C.sidereon_exact_cache_entry_copy_path(value, C.SIDEREON_EXACT_CACHE_COMPONENT_PROVENANCE, out, length, written, required)
			})
			if err != nil {
				return err
			}
			files.ProvenancePath = string(provenancePath)
			return nil
		})
	})
	runtime.KeepAlive(entry)
	return files, err
}

func (owner *ExactCacheOwner) Close() error {
	if owner == nil || owner.handle == nil {
		return nil
	}
	return owner.handle.close()
}

func (owner *ExactCacheOwner) Heartbeat() error {
	if owner == nil || owner.handle == nil {
		return ErrClosed
	}
	err := owner.handle.withExclusive(func(pointer unsafe.Pointer) error {
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_cache_owner_heartbeat((*C.SidereonExactCacheOwner)(pointer)))
		})
	})
	runtime.KeepAlive(owner)
	return err
}

func (owner *ExactCacheOwner) Publish(product, archive, provenance []byte) (*ExactCacheEntry, error) {
	if owner == nil || owner.handle == nil {
		return nil, ErrClosed
	}
	inputs, cleanup, err := copyExactCacheInputs(product, archive, provenance)
	if err != nil {
		return nil, err
	}
	defer cleanup()
	var pointer *C.SidereonExactCacheEntry
	attempted := false
	err = owner.handle.withExclusive(func(ownerPointer unsafe.Pointer) error {
		attempted = true
		return callStatus(func() uint32 {
			return uint32(C.sidereon_exact_cache_owner_publish(
				(*C.SidereonExactCacheOwner)(ownerPointer),
				inputs[0].pointer, inputs[0].length,
				inputs[1].pointer, inputs[1].length,
				inputs[2].pointer, inputs[2].length,
				&pointer,
			))
		})
	})
	if attempted {
		_ = owner.handle.close()
	}
	runtime.KeepAlive(owner)
	runtime.KeepAlive(product)
	runtime.KeepAlive(archive)
	runtime.KeepAlive(provenance)
	if err != nil {
		if pointer != nil {
			withCThread(func() { C.sidereon_exact_cache_entry_free(pointer) })
		}
		return nil, err
	}
	return exactCacheEntryFromPointer(pointer)
}
