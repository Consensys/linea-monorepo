cmake_minimum_required(VERSION 4.0)

project(gnark-gpu
    VERSION 0.1.0
    LANGUAGES CXX CUDA
    DESCRIPTION "CUDA accelerated field arithmetic for gnark (BLS12-377 + KoalaBear/Vortex)"
)

set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CUDA_STANDARD 20)
set(CMAKE_CUDA_STANDARD_REQUIRED ON)
set(CMAKE_EXPORT_COMPILE_COMMANDS ON)
set(CMAKE_CUDA_ARCHITECTURES native)

# Build position-independent code for shared library compatibility
set(CMAKE_POSITION_INDEPENDENT_CODE ON)

# Library (static for CGO integration)
add_library(gnark_gpu STATIC
    src/plonk/kernels.cu
    src/plonk/msm.cu
    src/plonk/ntt.cu
    src/plonk/fr_ops.cu
    src/plonk/plonk_z.cu
    src/plonk/plonk_eval.cu
    src/plonk/api.cu
    src/plonk2/g1.cu
    src/plonk2/kernels.cu
    src/plonk2/msm.cu
    src/vortex/kb.cu
)
set_target_properties(gnark_gpu PROPERTIES
    OUTPUT_NAME gnark_gpu
    CUDA_SEPARABLE_COMPILATION OFF
    CUDA_RESOLVE_DEVICE_SYMBOLS ON
)
target_include_directories(gnark_gpu
    PUBLIC
        $<BUILD_INTERFACE:${CMAKE_CURRENT_SOURCE_DIR}/include>
    PRIVATE
        ${CMAKE_CURRENT_SOURCE_DIR}/src/plonk
        ${CMAKE_CURRENT_SOURCE_DIR}/src/plonk2
        ${CMAKE_CURRENT_SOURCE_DIR}/src/vortex
)

# Sandbox executable (for profiling and benchmarking)
add_executable(sandbox sandbox/main.cu)
set_target_properties(sandbox PROPERTIES CUDA_SEPARABLE_COMPILATION ON)
target_include_directories(sandbox PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/src/plonk
    ${CMAKE_CURRENT_SOURCE_DIR}/sandbox
)
target_link_libraries(sandbox PRIVATE gnark_gpu)

# Profiling harness for Nsight Compute / Nsight Systems
add_executable(profile_kernels profiling/profile_kernels.cu)
target_include_directories(profile_kernels PRIVATE
    ${CMAKE_CURRENT_SOURCE_DIR}/include
)
target_link_libraries(profile_kernels PRIVATE gnark_gpu)
