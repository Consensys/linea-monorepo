const RATE_BYTES: usize = 136; // 1088 bits / 8
const OUTPUT_BYTES: usize = 32; // 256 bits / 8

/// Fixed-width input for padded Keccak entrypoints: 5440 bits = 680 bytes.
pub const KECCAK256_PADDED_BITS: usize = 5440;
pub const KECCAK256_PADDED_BYTES: usize = KECCAK256_PADDED_BITS / 8;

// Round constants
const RC: [u64; 24] = [
    0x0000000000000001,
    0x0000000000008082,
    0x800000000000808a,
    0x8000000080008000,
    0x000000000000808b,
    0x0000000080000001,
    0x8000000080008081,
    0x8000000000008009,
    0x000000000000008a,
    0x0000000000000088,
    0x0000000080008009,
    0x000000008000000a,
    0x000000008000808b,
    0x800000000000008b,
    0x8000000000008089,
    0x8000000000008003,
    0x8000000000008002,
    0x8000000000000080,
    0x000000000000800a,
    0x800000008000000a,
    0x8000000080008081,
    0x8000000000008080,
    0x0000000080000001,
    0x8000000080008008,
];

// ρ and π are applied as one fused 24-step update (matches Keccak-f[1600] reference).
const RHO: [u32; 24] = [
    1, 3, 6, 10, 15, 21, 28, 36, 45, 55, 2, 14, 27, 41, 56, 8, 25, 43, 62, 18, 39, 61, 20, 44,
];

const PI: [usize; 24] = [
    10, 7, 11, 17, 18, 3, 5, 16, 8, 21, 24, 4, 15, 23, 19, 13, 12, 2, 20, 14, 22, 9, 6, 1,
];

type State = [u64; 25];

/// THETA phase (diffusion)
fn theta(state: &mut State) {
    let mut c = [0u64; 5];
    let mut d = [0u64; 5];

    for x in 0..5 {
        c[x] = state[x] ^ state[x + 5] ^ state[x + 10] ^ state[x + 15] ^ state[x + 20];
    }

    for x in 0..5 {
        d[x] = c[(x + 4) % 5] ^ c[(x + 1) % 5].rotate_left(1);
    }

    for x in 0..5 {
        for y in 0..5 {
            state[x + 5 * y] ^= d[x];
        }
    }
}

/// ρ and π (fused Lane shift + rotation sequence from the Keccak reference).
fn rho_pi(state: &mut State) {
    let mut last = state[1];
    for x in 0..24 {
        let tmp = state[PI[x]];
        state[PI[x]] = last.rotate_left(RHO[x]);
        last = tmp;
    }
}

/// CHI phase (non-linear)
fn chi(state: &mut State) {
    let mut temp = *state;

    for x in 0..5 {
        for y in 0..5 {
            let idx = x + 5 * y;
            temp[idx] =
                state[idx] ^ ((!state[((x + 1) % 5) + 5 * y]) & state[((x + 2) % 5) + 5 * y]);
        }
    }

    *state = temp;
}

/// IOTA phase (round constant)
fn iota(state: &mut State, round: usize) {
    state[0] ^= RC[round];
}

/// Full Keccak-f[1600] permutation (24 rounds)
fn keccak_f(state: &mut State) {
    for round in 0..24 {
        theta(state);
        rho_pi(state);
        chi(state);
        iota(state, round);
    }
}

/// Convert bytes to u64 lane (little-endian)
fn bytes_to_lane(bytes: &[u8]) -> u64 {
    let mut lane = 0u64;
    for (i, &b) in bytes.iter().enumerate() {
        lane |= (b as u64) << (8 * i);
    }
    lane
}

/// Convert u64 lane to bytes (little-endian)
fn lane_to_bytes(lane: u64) -> [u8; 8] {
    [
        (lane & 0xff) as u8,
        ((lane >> 8) & 0xff) as u8,
        ((lane >> 16) & 0xff) as u8,
        ((lane >> 24) & 0xff) as u8,
        ((lane >> 32) & 0xff) as u8,
        ((lane >> 40) & 0xff) as u8,
        ((lane >> 48) & 0xff) as u8,
        ((lane >> 56) & 0xff) as u8,
    ]
}

/// Parse hex string to bytes.
/// Strips a `0x` / `0X` prefix. An odd number of hex digits is padded with a leading `0`
/// (same convention as `cast keccak 0xabc`).
fn hex_to_bytes(hex: &str) -> Result<Vec<u8>, String> {
    let hex = hex.trim_start_matches("0x").trim_start_matches("0X");
    let hex = if hex.len() % 2 == 0 {
        hex.to_string()
    } else {
        format!("0{hex}")
    };

    let mut bytes = Vec::new();
    for i in (0..hex.len()).step_by(2) {
        let byte_str = &hex[i..i + 2];
        let byte = u8::from_str_radix(byte_str, 16)
            .map_err(|_| format!("Invalid hex character at position {}", i))?;
        bytes.push(byte);
    }

    Ok(bytes)
}

/// Pad message to a multiple of RATE_BYTES
/// Ethereum Keccak-256 padding: append `0x01`, then `0x00` until the length is a multiple of the
/// rate, then OR `0x80` into the **last** byte (same as `sha3::Keccak256` — `0x01` and `0x80` may
/// combine in one byte, e.g. `0x81`).
fn pad_message(msg: &[u8]) -> Vec<u8> {
    let mut padded = msg.to_vec();
    padded.push(0x01);
    while padded.len() % RATE_BYTES != 0 {
        padded.push(0x00);
    }
    let last = padded.len() - 1;
    padded[last] |= 0x80;
    padded
}

/// Keccak-256 hash function that takes bytes
pub fn keccak256_bytes(msg: &[u8]) -> [u8; OUTPUT_BYTES] {
    // 1. Padding
    let padded = pad_message(msg);

    // 2. Initialize state to all zeros
    let mut state: State = [0u64; 25];

    // 3. Absorption phase
    for block_idx in 0..(padded.len() / RATE_BYTES) {
        let block = &padded[block_idx * RATE_BYTES..(block_idx + 1) * RATE_BYTES];

        // XOR block into the rate portion of the state (first 136 bytes = 17 lanes)
        for lane_idx in 0..17 {
            let lane_bytes = &block[lane_idx * 8..(lane_idx + 1) * 8];
            state[lane_idx] ^= bytes_to_lane(lane_bytes);
        }

        // Apply Keccak-f permutation
        keccak_f(&mut state);
    }

    // 4. Squeezing phase
    let mut output = [0u8; OUTPUT_BYTES];
    for lane_idx in 0..4 {
        let lane_bytes = lane_to_bytes(state[lane_idx]);
        for i in 0..8 {
            output[lane_idx * 8 + i] = lane_bytes[i];
        }
    }

    output
}

/// Interpret `padded` as one **big-endian** 5440-bit word (byte 0 = MSB). The logical message
/// is **left-padded with zero bits** to 5440 bits; `msg_len_bits` is its true length in bits.
/// The returned bytes pack the message in order: first message bit → MSB of the first output
/// byte (any trailing bits in the last byte are zero in the lower positions).
fn extract_left_padded_message_5440(
    padded: &[u8; KECCAK256_PADDED_BYTES],
    msg_len_bits: u64,
) -> Result<Vec<u8>, String> {
    if msg_len_bits > KECCAK256_PADDED_BITS as u64 {
        return Err(format!(
            "msg_len_bits {} exceeds {}-bit capacity",
            msg_len_bits, KECCAK256_PADDED_BITS
        ));
    }
    let len = msg_len_bits as usize;
    if len == 0 {
        return Ok(Vec::new());
    }

    let skip = KECCAK256_PADDED_BITS - len;
    let mut out = Vec::with_capacity((len + 7) / 8);
    let mut acc: u8 = 0;
    let mut acc_bits: u32 = 0;

    for bit_i in skip..skip + len {
        let byte_idx = bit_i / 8;
        let bit_in_byte = 7 - (bit_i % 8);
        let bit = (padded[byte_idx] >> bit_in_byte) & 1;
        acc = (acc << 1) | bit;
        acc_bits += 1;
        if acc_bits == 8 {
            out.push(acc);
            acc = 0;
            acc_bits = 0;
        }
    }
    if acc_bits > 0 {
        out.push(acc << (8 - acc_bits));
    }

    Ok(out)
}

/// Keccak-256 over a **5440-bit left-padded** message. `msg_len_bits` (typically passed as
/// `u64` in a circuit) selects the suffix of meaningful bits; leading zero bits are stripped.
pub fn keccak256_padded_5440(
    padded: &[u8; KECCAK256_PADDED_BYTES],
    msg_len_bits: u64,
) -> Result<[u8; OUTPUT_BYTES], String> {
    let msg = extract_left_padded_message_5440(padded, msg_len_bits)?;
    Ok(keccak256_bytes(&msg))
}

/// Keccak-256 over a 5440-bit **big-endian** buffer (`padded`) and a **64-bit length in bits**
/// (`msg_len_bits`). Leading zero bits are stripped before hashing. Digest is a `0x`-prefixed
/// hex string.
pub fn keccak256_hex(
    padded: &[u8; KECCAK256_PADDED_BYTES],
    msg_len_bits: u64,
) -> Result<String, String> {
    let hash = keccak256_padded_5440(padded, msg_len_bits)?;
    let hash_hex = hash
        .iter()
        .map(|b| format!("{:02x}", b))
        .collect::<String>();
    Ok(format!("0x{}", hash_hex))
}

fn main() {
    println!("Keccak-256 (5440-bit left-padded) Test");
    println!("=======================\n");

    let tests: Vec<(&str, u64, &str)> = vec![(
        "0x000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000008f0a7478c4baac2e5b972b4e8e5d9293a1baa404a9424664cddf6b31278b0b549743bcfca87ab8d4b93eb2e67ce82e783e9c6609463023620d3b010dca54755290ccb141d6c86002a5dffa8dcd715359c8b641e53149f78895a1593372a537af30781ddca7e5aa9b94a44916163dedba23595ee4849fcaa2a5743b1bed95e73d7ce877101211788ac69539d7ad7a748b2fef0d5f2d4718a21c379b999afca99fdd4443c71129a6d344521424d449d6b7522caf920971f99a5149b6a34e7f883cca1879806937a87ac4b67e0d53b716fbaf290bb76221d949344529651a5a1dc636bbdb0cc8130dbd07766e88a9ecb0e687b8a7d487f361b8c3ed8906e53b2ebe6bb4a2d438361488a5b61f50b1966ac82c66d16c4c5c2f791d66ba9817f2715ac25a3646bab8b85ca36d122903235b39cc859af090826e8fb670f1b2dcddd056591544a2124308ae2d3d2259281354d8b1d0e17c89315d5c7f4598d8a09595f044a3ef5f7333614678c6616559204b341e3a1add6a2f63f5834ae67c279ea48d2d959e2fca15ae0f3aa08619d1bcb518cab78b3ef47ec34d397e82acb6688eef5ecd1f42935074f3926ea6db1eb860b61f924b3dbad7fff51d6f8725105ab182da0306d2752e2598fb025aa068061dcaef4198bc9648a0b368790c362b1dc1f2703138ee59164c2c69cae2aa15d2b0a42e89459e1e06349774084f1531b03b548875248e8b9be35ad0c2bceec8770e4a4072c8228fa0efa291c7afc8845bf",
        4344,
        "ffceda84b040d6dbbb79169f8b77507411a14b937624dace0c981da90a011420",
    )];

    for (padded_hex, len_bits, expected) in tests {
        let bytes = match hex_to_bytes(padded_hex) {
            Ok(b) => b,
            Err(e) => {
                println!("Decode padded hex - Error: {}\n", e);
                continue;
            }
        };
        let padded: [u8; KECCAK256_PADDED_BYTES] = match bytes.try_into() {
            Ok(arr) => arr,
            Err(v) => {
                println!(
                    "Expected {} padded bytes, got {}\n",
                    KECCAK256_PADDED_BYTES,
                    v.len()
                );
                continue;
            }
        };

        match keccak256_hex(&padded, len_bits) {
            Ok(result) => {
                let matches = result.to_lowercase() == format!("0x{}", expected.to_lowercase());
                let status = if matches { "✅ PASS" } else { "❌ FAIL" };

                println!("len_bits: {}", len_bits);
                println!("Expected: 0x{}", expected);
                println!("Got:      {}", result);
                println!("Status:   {}\n", status);
            }
            Err(e) => {
                println!("Hash error: {}\n", e);
            }
        }
    }
}
