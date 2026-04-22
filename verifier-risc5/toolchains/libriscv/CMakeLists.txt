cmake_minimum_required(VERSION 3.20)
project(libriscv_runner LANGUAGES CXX)

if(NOT DEFINED LIBRISCV_SOURCE_DIR)
  set(LIBRISCV_SOURCE_DIR "${CMAKE_CURRENT_SOURCE_DIR}/../../third_party/libriscv")
endif()

set(RISCV_BINARY_TRANSLATION OFF CACHE BOOL "" FORCE)
set(RISCV_LIBTCC OFF CACHE BOOL "" FORCE)
set(RISCV_32I OFF CACHE BOOL "" FORCE)
set(RISCV_64I ON CACHE BOOL "" FORCE)
set(RISCV_128I OFF CACHE BOOL "" FORCE)
set(RISCV_MEMORY_TRAPS ON CACHE BOOL "" FORCE)

add_subdirectory("${LIBRISCV_SOURCE_DIR}/lib" "${CMAKE_CURRENT_BINARY_DIR}/libriscv")

add_executable(libriscv_runner runner.cpp)
target_compile_features(libriscv_runner PRIVATE cxx_std_20)
target_link_libraries(libriscv_runner PRIVATE riscv)
