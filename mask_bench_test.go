//go:build goexperiment.simd

package ws_test

import (
	"runtime"
	"testing"

	"github.com/EetuAH/go-ws-mask-simd-bench"
)

var result byte
var benchMaskFunc = benchMaskCopy

// idealized, hot-buffer
func benchMaskInPlace(b *testing.B, size int, fn func([]byte, []byte)) {
	key := []byte{1, 2, 3, 4}
	buf := make([]byte, size)

	// Fill buf with deterministic but non-zero data
	for i := range buf {
		buf[i] = byte(i*31 + 7)
	}

	b.SetBytes(int64(size))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		fn(buf, key)
		result ^= buf[0] // prevent elimination optimization
	}
}

// realistic, memory traffic
func benchMaskCopy(b *testing.B, size int, fn func([]byte, []byte)) {
	key := []byte{1, 2, 3, 4}

	// Two buffers to avoid XOR toggling effects
	src := make([]byte, size)
	dst := make([]byte, size)

	// Fill src with deterministic but non-zero data
	for i := range src {
		src[i] = byte(i*31 + 7)
	}

	b.ReportAllocs()
	b.SetBytes(int64(size))
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		// Copy to avoid repeated toggling of same memory
		copy(dst, src)

		fn(dst, key)

		// Touch a changing byte so compiler can't eliminate work
		result ^= dst[i%size]
	}

	// Prevent clever compiler tricks
	runtime.KeepAlive(dst)
}

// Original version

func BenchmarkMaskOriginal_48B(b *testing.B) {
	benchMaskFunc(b, 48, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_64B(b *testing.B) {
	benchMaskFunc(b, 64, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_128B(b *testing.B) {
	benchMaskFunc(b, 128, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_256B(b *testing.B) {
	benchMaskFunc(b, 256, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_512B(b *testing.B) {
	benchMaskFunc(b, 512, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_1KB(b *testing.B) {
	benchMaskFunc(b, 1024, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_4KB(b *testing.B) {
	benchMaskFunc(b, 4*1024, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_16KB(b *testing.B) {
	benchMaskFunc(b, 16*1024, ws.MaskXOR_Original)
}

func BenchmarkMaskOriginal_256KB(b *testing.B) {
	benchMaskFunc(b, 256*1024, ws.MaskXOR_Original)
}

// Original optimized version

func BenchmarkMaskOriginalOptimized_48B(b *testing.B) {
	benchMaskFunc(b, 48, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_64B(b *testing.B) {
	benchMaskFunc(b, 64, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_128B(b *testing.B) {
	benchMaskFunc(b, 128, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_256B(b *testing.B) {
	benchMaskFunc(b, 256, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_512B(b *testing.B) {
	benchMaskFunc(b, 512, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_1KB(b *testing.B) {
	benchMaskFunc(b, 1024, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_4KB(b *testing.B) {
	benchMaskFunc(b, 4*1024, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_16KB(b *testing.B) {
	benchMaskFunc(b, 16*1024, ws.MaskXOR_Original_Optimized)
}

func BenchmarkMaskOriginalOptimized_256KB(b *testing.B) {
	benchMaskFunc(b, 256*1024, ws.MaskXOR_Original_Optimized)
}

// Only builds with SIMD128 enabled

func BenchmarkMaskSIMD128_48B(b *testing.B) {
	benchMaskFunc(b, 48, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_64B(b *testing.B) {
	benchMaskFunc(b, 64, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_128B(b *testing.B) {
	benchMaskFunc(b, 128, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_256B(b *testing.B) {
	benchMaskFunc(b, 256, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_512B(b *testing.B) {
	benchMaskFunc(b, 512, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_1KB(b *testing.B) {
	benchMaskFunc(b, 1024, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_4KB(b *testing.B) {
	benchMaskFunc(b, 4*1024, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_16KB(b *testing.B) {
	benchMaskFunc(b, 16*1024, ws.MaskXOR_SIMD128)
}

func BenchmarkMaskSIMD128_256KB(b *testing.B) {
	benchMaskFunc(b, 256*1024, ws.MaskXOR_SIMD128)
}

// Only builds with SIMD256 enabled

func BenchmarkMaskSIMD256_48B(b *testing.B) {
	benchMaskFunc(b, 48, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_64B(b *testing.B) {
	benchMaskFunc(b, 64, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_128B(b *testing.B) {
	benchMaskFunc(b, 128, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_256B(b *testing.B) {
	benchMaskFunc(b, 256, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_512B(b *testing.B) {
	benchMaskFunc(b, 512, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_1KB(b *testing.B) {
	benchMaskFunc(b, 1024, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_4KB(b *testing.B) {
	benchMaskFunc(b, 4*1024, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_16KB(b *testing.B) {
	benchMaskFunc(b, 16*1024, ws.MaskXOR_SIMD256)
}

func BenchmarkMaskSIMD256_256KB(b *testing.B) {
	benchMaskFunc(b, 256*1024, ws.MaskXOR_SIMD256)
}

// Only builds with SIMD256&SIMD128 enabled

func BenchmarkMaskSIMD256_128_48B(b *testing.B) {
	benchMaskFunc(b, 48, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_64B(b *testing.B) {
	benchMaskFunc(b, 64, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_128B(b *testing.B) {
	benchMaskFunc(b, 128, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_256B(b *testing.B) {
	benchMaskFunc(b, 256, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_512B(b *testing.B) {
	benchMaskFunc(b, 512, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_1KB(b *testing.B) {
	benchMaskFunc(b, 1024, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_4KB(b *testing.B) {
	benchMaskFunc(b, 4*1024, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_16KB(b *testing.B) {
	benchMaskFunc(b, 16*1024, ws.MaskXOR_SIMD256_128)
}

func BenchmarkMaskSIMD256_128_256KB(b *testing.B) {
	benchMaskFunc(b, 256*1024, ws.MaskXOR_SIMD256_128)
}
