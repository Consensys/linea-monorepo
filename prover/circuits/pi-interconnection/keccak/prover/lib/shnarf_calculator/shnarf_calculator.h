#include <stdbool.h>

#ifndef COMPRESSION_H
#define COMPRESSION_H

typedef struct {
    char* commitment;
    char* kzg_proof_contract;
    char* kzg_proof_sidecar;
    char* data_hash;
    char* snark_hash;
    char* expected_x;
    char* expected_y;
    char* expected_shnarf;
    char* error_message;
} response;

#endif // COMPRESSION_H
