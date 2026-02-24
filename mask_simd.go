//go:build goexperiment.simd

package ws

import (
	"encoding/binary"
	"simd/archsimd"
)

// MaskXOR_SIMD128 applies the WebSocket masking key to b using 128-bit SIMD.
//
// For buffers >= 512 bytes, it uses 128-bit vector XOR (8x16B unrolled per iteration, 128B per loop)
// to improve throughput while avoiding AVX2 frequency downclock effects on client/mobile Intel CPUs.
//
// Smaller buffers fall back to a fully unrolled scalar 64-bit path,
// which is faster for tiny workloads due to lower setup cost.
//
// Performs better on benchMaskCopy than the MaskXOR_SIMD256
func MaskXOR_SIMD128(b []byte, key []byte) {
	key32 := binary.LittleEndian.Uint32(key)
	key64 := uint64(key32)<<32 | uint64(key32)

	// simd initialization is expensive, so we only use it for large buffers
	if len(b) >= 512 {
		var key128Bytes [16]byte
		binary.LittleEndian.PutUint64(key128Bytes[0:8], key64)
		binary.LittleEndian.PutUint64(key128Bytes[8:16], key64)

		key128 := archsimd.LoadUint8x16(&key128Bytes)

		for len(b) >= 128 {
			v := archsimd.LoadUint8x16Slice(b[0:16]).Xor(key128)
			v.StoreSlice(b[0:16])

			v = archsimd.LoadUint8x16Slice(b[16:32]).Xor(key128)
			v.StoreSlice(b[16:32])

			v = archsimd.LoadUint8x16Slice(b[32:48]).Xor(key128)
			v.StoreSlice(b[32:48])

			v = archsimd.LoadUint8x16Slice(b[48:64]).Xor(key128)
			v.StoreSlice(b[48:64])

			v = archsimd.LoadUint8x16Slice(b[64:80]).Xor(key128)
			v.StoreSlice(b[64:80])

			v = archsimd.LoadUint8x16Slice(b[80:96]).Xor(key128)
			v.StoreSlice(b[80:96])

			v = archsimd.LoadUint8x16Slice(b[96:112]).Xor(key128)
			v.StoreSlice(b[96:112])

			v = archsimd.LoadUint8x16Slice(b[112:128]).Xor(key128)
			v.StoreSlice(b[112:128])

			b = b[128:]
		}
		if len(b) == 0 {
			return
		}
	}

	for len(b) >= 64 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		v = binary.LittleEndian.Uint64(b[32:40])
		binary.LittleEndian.PutUint64(b[32:40], v^key64)

		v = binary.LittleEndian.Uint64(b[40:48])
		binary.LittleEndian.PutUint64(b[40:48], v^key64)

		v = binary.LittleEndian.Uint64(b[48:56])
		binary.LittleEndian.PutUint64(b[48:56], v^key64)

		v = binary.LittleEndian.Uint64(b[56:64])
		binary.LittleEndian.PutUint64(b[56:64], v^key64)

		b = b[64:]
	}
	if len(b) == 0 {
		return
	}

	if len(b) >= 32 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		b = b[32:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 16 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		b = b[16:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 8 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		b = b[8:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 4 {
		v := binary.LittleEndian.Uint32(b[0:4])
		binary.LittleEndian.PutUint32(b[0:4], v^key32)

		b = b[4:]
		if len(b) == 0 {
			return
		}
	}

	for i := range b {
		b[i] ^= key[i&3]
	}
}

// MaskXOR_SIMD256 applies the WebSocket masking key using 256-bit AVX2.
//
// For buffers >= 512 bytes, it processes 128 bytes per iteration using
// 4x32B vector loads/stores.
//
// On some Intel client/mobile CPUs, sustained 256-bit AVX2 may reduce
// core frequency.
//
// Smaller buffers fall back to the scalar 64-bit path.
//
// Performs better well on benchMaskInPlace, but suffers more on benchMaskCopy for < 4KB buffers.
func MaskXOR_SIMD256(b []byte, key []byte) {
	key32 := binary.LittleEndian.Uint32(key)
	key64 := uint64(key32)<<32 | uint64(key32)

	// simd initialization is expensive, so we only use it for large buffers
	if len(b) >= 512 {
		var key256Bytes [32]byte
		binary.LittleEndian.PutUint64(key256Bytes[0:8], key64)
		binary.LittleEndian.PutUint64(key256Bytes[8:16], key64)
		binary.LittleEndian.PutUint64(key256Bytes[16:24], key64)
		binary.LittleEndian.PutUint64(key256Bytes[24:32], key64)

		key256 := archsimd.LoadUint8x32(&key256Bytes)

		for len(b) >= 128 {
			v := archsimd.LoadUint8x32Slice(b[0:32]).Xor(key256)
			v.StoreSlice(b[0:32])

			v = archsimd.LoadUint8x32Slice(b[32:64]).Xor(key256)
			v.StoreSlice(b[32:64])

			v = archsimd.LoadUint8x32Slice(b[64:96]).Xor(key256)
			v.StoreSlice(b[64:96])

			v = archsimd.LoadUint8x32Slice(b[96:128]).Xor(key256)
			v.StoreSlice(b[96:128])

			b = b[128:]
		}
		if len(b) == 0 {
			return
		}
	}

	for len(b) >= 64 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		v = binary.LittleEndian.Uint64(b[32:40])
		binary.LittleEndian.PutUint64(b[32:40], v^key64)

		v = binary.LittleEndian.Uint64(b[40:48])
		binary.LittleEndian.PutUint64(b[40:48], v^key64)

		v = binary.LittleEndian.Uint64(b[48:56])
		binary.LittleEndian.PutUint64(b[48:56], v^key64)

		v = binary.LittleEndian.Uint64(b[56:64])
		binary.LittleEndian.PutUint64(b[56:64], v^key64)

		b = b[64:]
	}
	if len(b) == 0 {
		return
	}

	if len(b) >= 32 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		b = b[32:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 16 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		b = b[16:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 8 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		b = b[8:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 4 {
		v := binary.LittleEndian.Uint32(b[0:4])
		binary.LittleEndian.PutUint32(b[0:4], v^key32)

		b = b[4:]
		if len(b) == 0 {
			return
		}
	}

	for i := range b {
		b[i] ^= key[i&3]
	}
}

// MaskXOR_SIMD256_128 is a hybrid masking implementation.
//
// - >= 4096 bytes: uses 256-bit AVX2 for maximum throughput
// - >= 512 bytes:  uses 128-bit SIMD to avoid AVX2 frequency penalties
// - smaller sizes: falls back to scalar 64-bit unrolled XOR
//
// This tiered approach was chosen empirically to balance AVX2 downclock
// behavior on Intel client/mobile CPUs while retaining peak performance for large buffers.
func MaskXOR_SIMD256_128(b []byte, key []byte) {
	key32 := binary.LittleEndian.Uint32(key)
	key64 := uint64(key32)<<32 | uint64(key32)

	// simd initialization is expensive, so we only use it for large buffers
	// avx2 is very slow to initialize in benchMaskCopy, so only use for 4096kb+ buffers
	if len(b) >= 4096 {
		var key256Bytes [32]byte
		binary.LittleEndian.PutUint64(key256Bytes[0:8], key64)
		binary.LittleEndian.PutUint64(key256Bytes[8:16], key64)
		binary.LittleEndian.PutUint64(key256Bytes[16:24], key64)
		binary.LittleEndian.PutUint64(key256Bytes[24:32], key64)

		key256 := archsimd.LoadUint8x32(&key256Bytes)

		for len(b) >= 128 {
			v := archsimd.LoadUint8x32Slice(b[0:32]).Xor(key256)
			v.StoreSlice(b[0:32])

			v = archsimd.LoadUint8x32Slice(b[32:64]).Xor(key256)
			v.StoreSlice(b[32:64])

			v = archsimd.LoadUint8x32Slice(b[64:96]).Xor(key256)
			v.StoreSlice(b[64:96])

			v = archsimd.LoadUint8x32Slice(b[96:128]).Xor(key256)
			v.StoreSlice(b[96:128])

			b = b[128:]
		}
		if len(b) == 0 {
			return
		}
	}

	// initialize 128-bit SIMD for 512+ byte buffers
	if len(b) >= 512 {
		var key128Bytes [16]byte
		binary.LittleEndian.PutUint64(key128Bytes[0:8], key64)
		binary.LittleEndian.PutUint64(key128Bytes[8:16], key64)

		key128 := archsimd.LoadUint8x16(&key128Bytes)

		for len(b) >= 128 {
			v := archsimd.LoadUint8x16Slice(b[0:16]).Xor(key128)
			v.StoreSlice(b[0:16])

			v = archsimd.LoadUint8x16Slice(b[16:32]).Xor(key128)
			v.StoreSlice(b[16:32])

			v = archsimd.LoadUint8x16Slice(b[32:48]).Xor(key128)
			v.StoreSlice(b[32:48])

			v = archsimd.LoadUint8x16Slice(b[48:64]).Xor(key128)
			v.StoreSlice(b[48:64])

			v = archsimd.LoadUint8x16Slice(b[64:80]).Xor(key128)
			v.StoreSlice(b[64:80])

			v = archsimd.LoadUint8x16Slice(b[80:96]).Xor(key128)
			v.StoreSlice(b[80:96])

			v = archsimd.LoadUint8x16Slice(b[96:112]).Xor(key128)
			v.StoreSlice(b[96:112])

			v = archsimd.LoadUint8x16Slice(b[112:128]).Xor(key128)
			v.StoreSlice(b[112:128])

			b = b[128:]
		}
		if len(b) == 0 {
			return
		}
	}

	for len(b) >= 64 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		v = binary.LittleEndian.Uint64(b[32:40])
		binary.LittleEndian.PutUint64(b[32:40], v^key64)

		v = binary.LittleEndian.Uint64(b[40:48])
		binary.LittleEndian.PutUint64(b[40:48], v^key64)

		v = binary.LittleEndian.Uint64(b[48:56])
		binary.LittleEndian.PutUint64(b[48:56], v^key64)

		v = binary.LittleEndian.Uint64(b[56:64])
		binary.LittleEndian.PutUint64(b[56:64], v^key64)

		b = b[64:]
	}
	if len(b) == 0 {
		return
	}

	if len(b) >= 32 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		v = binary.LittleEndian.Uint64(b[16:24])
		binary.LittleEndian.PutUint64(b[16:24], v^key64)

		v = binary.LittleEndian.Uint64(b[24:32])
		binary.LittleEndian.PutUint64(b[24:32], v^key64)

		b = b[32:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 16 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		v = binary.LittleEndian.Uint64(b[8:16])
		binary.LittleEndian.PutUint64(b[8:16], v^key64)

		b = b[16:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 8 {
		v := binary.LittleEndian.Uint64(b[0:8])
		binary.LittleEndian.PutUint64(b[0:8], v^key64)

		b = b[8:]
		if len(b) == 0 {
			return
		}
	}

	if len(b) >= 4 {
		v := binary.LittleEndian.Uint32(b[0:4])
		binary.LittleEndian.PutUint32(b[0:4], v^key32)

		b = b[4:]
		if len(b) == 0 {
			return
		}
	}

	for i := range b {
		b[i] ^= key[i&3]
	}
}
