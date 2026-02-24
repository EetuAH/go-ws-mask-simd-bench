#  WebSocket Mask XOR Benchmark (Go 1.26 SIMD)

This repository benchmarks multiple WebSocket masking (XOR) implementations:

* **Original** - from [`lxzan/gws`](https://github.com/lxzan/gws/blob/main/internal/utils.go#L87)
* **OriginalOptimized** - small scalar improvements
* **SIMD128** - Go 1.26 `archsimd` (128-bit)
* **SIMD256** - AVX2 (256-bit)
* **SIMD256_128 (Hybrid)** - AVX2 for large buffers, SIMD128 fallback for mid-size buffers

> Built using **Go 1.26** with `//go:build goexperiment.simd` and the new `archsimd` package.

Benchmarks run on:

```
CPU: 13th Gen Intel(R) Core(TM) i9-13900H
goos: linux
goarch: amd64
```

---

## In-Place Benchmark (Hot Buffer)

* **Original is massively slower at 48B.**
* OriginalOptimized quickly catches up.
* SIMD128 starts pulling ahead at 512B+.
* SIMD256 dominates strongly from 1KB upward.
* Hybrid matches or slightly beats pure SIMD256 at larger sizes.
* 16KB peak shows ~94 GB/s ceiling territory.

![In-Place Benchmark](img/test_in_place.png)


---

## Copy Benchmark (Streaming / Realistic)

* Original and OriginalOptimized cluster closely.
* SIMD128 improves steadily.
* **Pure SIMD256 collapses at 512B / 1KB (AVX downclock zone).**
* Hybrid completely fixes that dip.
* ≥4KB shows clear SIMD256 advantage.
* 256KB shows memory-bandwidth limit across all variants.

![Copy Benchmark](img/test_copy.png)

---

# 🧠 Why SIMD256 Collapses at 512B–1KB?

On many Intel CPUs (including 13th gen):

* Heavy AVX2 usage can trigger **frequency downclocking**
* For small workloads, AVX startup cost > benefit
* In streaming scenarios, memory pressure amplifies this

The **Hybrid approach** avoids this by:

* Using AVX2 only for large buffers
* Falling back to 128-bit SIMD for mid-sized buffers

This gives the best overall performance profile.

---

# 🏁 Summary

| Size Range          | Best Implementation         |
| ------------------- | --------------------------- |
| ≤256B               | OriginalOptimized / SIMD128 |
| 512B–1KB            | SIMD128                     |
| ≥4KB                | SIMD256 or Hybrid           |
| Streaming workloads | Hybrid                      |


---

## Test Results (Visualized in the Graphs)

### `var benchMaskFunc = benchMaskInPlace`

```
Running tool: /home/eetu/.gvm/gos/go1.26.0/bin/go test -test.fullpath=true -benchmem -run=^$ -coverprofile=/tmp/vscode-gogVnK9U/go-code-cover -bench . github.com/EetuAH/go-ws-mask-simd-bench

goos: linux
goarch: amd64
pkg: github.com/EetuAH/go-ws-mask-simd-bench
cpu: 13th Gen Intel(R) Core(TM) i9-13900H
BenchmarkMaskOriginal_48B-20               	34071099	        29.62 ns/op	1620.76 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_64B-20               	289673161	         4.189 ns/op	15276.87 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_128B-20              	213572990	         5.569 ns/op	22982.57 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_256B-20              	139787794	         8.582 ns/op	29830.10 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_512B-20              	74901933	        14.74 ns/op	34743.74 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_1KB-20               	42666604	        27.08 ns/op	37809.00 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_4KB-20               	10785301	       111.4 ns/op	36768.14 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_16KB-20              	 2853536	       419.9 ns/op	39016.98 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_256KB-20             	  148568	      7539 ns/op	34772.44 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_48B-20      	267460255	         4.323 ns/op	11104.53 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_64B-20      	306349160	         3.838 ns/op	16676.32 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_128B-20     	222374852	         5.344 ns/op	23951.07 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_256B-20     	143018490	         8.390 ns/op	30514.22 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_512B-20     	75636919	        14.47 ns/op	35376.35 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_1KB-20      	45015630	        26.64 ns/op	38438.68 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_4KB-20      	11763145	       101.3 ns/op	40429.53 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_16KB-20     	 2977834	       402.2 ns/op	40731.06 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_256KB-20    	  158445	      7641 ns/op	34306.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_48B-20                	268915604	         4.461 ns/op	10761.08 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_64B-20                	293037900	         4.052 ns/op	15794.52 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_128B-20               	215643660	         5.578 ns/op	22946.22 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_256B-20               	138940147	         8.580 ns/op	29835.80 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_512B-20               	86917975	        13.70 ns/op	37374.11 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_1KB-20                	62050648	        19.28 ns/op	53113.54 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_4KB-20                	18244286	        66.08 ns/op	61984.51 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_16KB-20               	 4848559	       245.3 ns/op	66780.88 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_256KB-20              	  252157	      4710 ns/op	55657.12 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_48B-20                	263053634	         4.530 ns/op	10597.11 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_64B-20                	287505613	         4.056 ns/op	15778.96 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128B-20               	212372400	         5.623 ns/op	22764.50 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_256B-20               	137645920	         8.755 ns/op	29240.81 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_512B-20               	94967655	        12.68 ns/op	40390.47 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_1KB-20                	66643185	        16.07 ns/op	63737.30 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_4KB-20                	24771553	        49.81 ns/op	82240.59 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_16KB-20               	 6792140	       174.6 ns/op	93861.45 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_256KB-20              	  285396	      4112 ns/op	63746.28 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_48B-20            	257787620	         4.615 ns/op	10399.97 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_64B-20            	284436734	         4.165 ns/op	15365.42 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_128B-20           	210889581	         5.668 ns/op	22583.87 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_256B-20           	138078789	         8.684 ns/op	29480.23 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_512B-20           	87974313	        13.63 ns/op	37559.13 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_1KB-20            	62260396	        19.33 ns/op	52977.41 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_4KB-20            	24869986	        48.35 ns/op	84708.40 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_16KB-20           	 6907329	       173.6 ns/op	94360.17 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_256KB-20          	  292030	      4102 ns/op	63907.70 MB/s	       0 B/op	       0 allocs/op
PASS
coverage: 82.2% of statements
ok  	github.com/EetuAH/go-ws-mask-simd-bench	70.462s
```

---


### `var benchMaskFunc = benchMaskCopy`

```
Running tool: /home/eetu/.gvm/gos/go1.26.0/bin/go test -test.fullpath=true -benchmem -run=^$ -coverprofile=/tmp/vscode-gogVnK9U/go-code-cover -bench . github.com/EetuAH/go-ws-mask-simd-bench

goos: linux
goarch: amd64
pkg: github.com/EetuAH/go-ws-mask-simd-bench
cpu: 13th Gen Intel(R) Core(TM) i9-13900H
BenchmarkMaskOriginal_48B-20               	45289194	        26.51 ns/op	1810.50 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_64B-20               	192080500	         6.253 ns/op	10234.94 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_128B-20              	149288370	         8.010 ns/op	15981.00 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_256B-20              	98095380	        12.06 ns/op	21230.31 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_512B-20              	58255618	        20.50 ns/op	24978.07 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_1KB-20               	29019222	        34.84 ns/op	29393.29 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_4KB-20               	 8581066	       140.1 ns/op	29234.98 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_16KB-20              	 2348900	       511.6 ns/op	32024.63 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginal_256KB-20             	   88813	     12698 ns/op	20643.73 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_48B-20      	187455247	         6.426 ns/op	7469.59 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_64B-20      	204756044	         5.868 ns/op	10906.32 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_128B-20     	150110875	         7.947 ns/op	16105.98 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_256B-20     	94768135	        11.99 ns/op	21345.19 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_512B-20     	59201504	        20.27 ns/op	25256.93 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_1KB-20      	34859120	        34.37 ns/op	29790.90 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_4KB-20      	 8957829	       134.5 ns/op	30445.41 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_16KB-20     	 2432966	       493.1 ns/op	33224.28 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskOriginalOptimized_256KB-20    	   94621	     12853 ns/op	20395.17 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_48B-20                	184092405	         6.406 ns/op	7493.25 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_64B-20                	194893665	         6.163 ns/op	10384.82 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_128B-20               	150403326	         7.995 ns/op	16010.83 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_256B-20               	97388020	        11.96 ns/op	21397.72 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_512B-20               	64244688	        18.96 ns/op	27002.67 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_1KB-20                	42085699	        28.48 ns/op	35951.88 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_4KB-20                	11745216	       100.6 ns/op	40702.07 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_16KB-20               	 3517694	       341.6 ns/op	47957.78 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD128_256KB-20              	  113604	      9883 ns/op	26525.28 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_48B-20                	187073743	         6.387 ns/op	7514.82 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_64B-20                	196142006	         6.118 ns/op	10460.96 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128B-20               	144540705	         8.352 ns/op	15325.78 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_256B-20               	99002716	        12.09 ns/op	21179.19 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_512B-20               	 9295951	       123.6 ns/op	4141.25 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_1KB-20                	 9007368	       132.5 ns/op	7730.94 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_4KB-20                	14802586	        80.73 ns/op	50738.10 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_16KB-20               	 4483868	       266.2 ns/op	61540.74 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_256KB-20              	  115726	      9197 ns/op	28502.77 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_48B-20            	180835978	         6.703 ns/op	7161.23 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_64B-20            	190815795	         6.242 ns/op	10253.58 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_128B-20           	145451961	         8.249 ns/op	15516.51 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_256B-20           	100292130	        11.95 ns/op	21421.13 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_512B-20           	63567884	        18.83 ns/op	27190.95 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_1KB-20            	42054553	        28.46 ns/op	35978.96 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_4KB-20            	13163283	        80.11 ns/op	51129.60 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_16KB-20           	 4526919	       265.1 ns/op	61797.48 MB/s	       0 B/op	       0 allocs/op
BenchmarkMaskSIMD256_128_256KB-20          	  118902	      9205 ns/op	28478.19 MB/s	       0 B/op	       0 allocs/op
PASS
coverage: 82.2% of statements
ok  	github.com/EetuAH/go-ws-mask-simd-bench	70.856s
```
