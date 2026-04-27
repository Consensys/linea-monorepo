#pragma once

#include <cuda_runtime.h>
#include <cstdio>
#include <algorithm>
#include <vector>

namespace gnark_gpu {

// Simple CUDA event-based timer
struct GpuTimer {
    cudaEvent_t start, stop;

    GpuTimer() {
        cudaEventCreate(&start);
        cudaEventCreate(&stop);
    }

    ~GpuTimer() {
        cudaEventDestroy(start);
        cudaEventDestroy(stop);
    }

    void begin(cudaStream_t stream = 0) {
        cudaEventRecord(start, stream);
    }

    float end(cudaStream_t stream = 0) {
        cudaEventRecord(stop, stream);
        cudaEventSynchronize(stop);
        float ms;
        cudaEventElapsedTime(&ms, start, stop);
        return ms;
    }
};

// Benchmark a kernel with warmup and multiple iterations
template<typename F>
void bench(const char* name, int warmup, int iters, size_t bytes, F&& kernel) {
    GpuTimer timer;

    // Warmup
    for (int i = 0; i < warmup; i++) {
        kernel();
    }
    cudaDeviceSynchronize();

    // Timed runs
    std::vector<float> times(iters);
    for (int i = 0; i < iters; i++) {
        timer.begin();
        kernel();
        times[i] = timer.end();
    }

    // Stats
    std::sort(times.begin(), times.end());
    float median = times[iters / 2];
    float min = times[0];
    float max = times[iters - 1];

    // Throughput
    double gb = bytes / 1e9;
    double gbps = gb / (median / 1000.0);

    printf("%-30s  %8.3f ms (min %.3f, max %.3f)  %8.2f GB/s\n",
           name, median, min, max, gbps);
}

} // namespace gnark_gpu
