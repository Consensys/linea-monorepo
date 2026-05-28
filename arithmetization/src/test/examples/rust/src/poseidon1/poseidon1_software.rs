#![no_std]
#![no_main]
#![feature(generic_const_exprs)]
#![allow(incomplete_features)]

// Pure-Rust Poseidon1 implementation, intended for cycle-count benchmarking
// against the custom-0 precompile (see poseidon1_with_in_embedded.rs).
//
// This file does NOT use the custom `.insn r 0x0b, ...` precompile path; the
// permutation is computed entirely in software using ordinary RV64 IM
// instructions. The harness runs the same 7 active test cases as the
// embedded-precompile variant so that the two binaries produce directly
// comparable cycle counts.
//
// Constants are the real precompile values, copied inline as per-element
// assignments so they live in `.text` (the loader does not extract `.rodata`
// at the moment). Sources:
//
//   - round constants: src/main/precompiles/poseidon/impl/_round_constants.zkc
//     (the `default_round_constants` table — the one the precompile actually
//     reads; the sibling `round_constants` table is dead code)
//   - MDS matrix:      src/main/precompiles/poseidon/impl/_mds_matrix.zkc

// === field parameters ===

const KOALABEAR_PRIME: u32 = 0x7F00_0001; // 2^31 - 2^24 + 1 = 2_130_706_433
// ALPHA = 3 ; embedded directly in fpow_alpha.

// === poseidon parameters ===

const RATE: usize = 15;
const CAPACITY: usize = 1;
const STATE_WIDTH: usize = RATE + CAPACITY; // 16
const FULL_ROUNDS_HALF: usize = 4;
const PRTL_ROUNDS: usize = 20;
const TOTAL_ROUNDS: usize = 2 * FULL_ROUNDS_HALF + PRTL_ROUNDS; // 28

// === I/O encoding (matches the precompile) ===

const BYTES_PER_INPUT_FELT: usize = 3;
const BYTES_PER_BLOCK: usize = RATE * BYTES_PER_INPUT_FELT; // 45

// === constants tables ===

const N_ROUND_CONSTANTS: usize = TOTAL_ROUNDS * STATE_WIDTH; // 448
const N_MDS_VALUES: usize = STATE_WIDTH * STATE_WIDTH;       // 256

include!("../custom_std.rs");

// === field arithmetic ===

#[inline(always)]
fn fadd(a: u32, b: u32) -> u32 {
    let s = a as u64 + b as u64;
    if s >= KOALABEAR_PRIME as u64 {
        (s - KOALABEAR_PRIME as u64) as u32
    } else {
        s as u32
    }
}

#[inline(always)]
fn fmul(a: u32, b: u32) -> u32 {
    ((a as u64).wrapping_mul(b as u64) % KOALABEAR_PRIME as u64) as u32
}

#[inline(always)]
fn fpow_alpha(x: u32) -> u32 {
    // ALPHA = 3, so x^3 = x * (x * x).
    let x2 = fmul(x, x);
    fmul(x, x2)
}

// === constants ===
//
// The 448 round constants and 256 MDS-matrix values are copied inline as
// per-element stores so they materialize in `.text` (the loader does not
// extract `.rodata`). Generated mechanically from _round_constants.zkc's
// `default_round_constants` table and _mds_matrix.zkc.

fn fill_round_constants(rc: &mut [u32; N_ROUND_CONSTANTS]) {
    rc[0] = 0x177366cd;
    rc[1] = 0x4b6952d1;
    rc[2] = 0x184dc725;
    rc[3] = 0x4368079e;
    rc[4] = 0x182b19e8;
    rc[5] = 0x398d6f81;
    rc[6] = 0x28130667;
    rc[7] = 0x4bc97c39;
    rc[8] = 0x4c0bc40e;
    rc[9] = 0x0c8a6c42;
    rc[10] = 0x77df9150;
    rc[11] = 0x38d2a89a;
    rc[12] = 0x1e9a74e8;
    rc[13] = 0x14ee8de2;
    rc[14] = 0x222debd7;
    rc[15] = 0x1802c1fd;
    rc[16] = 0x1857fc8a;
    rc[17] = 0x60ff6f42;
    rc[18] = 0x00bb6a62;
    rc[19] = 0x50faa1dc;
    rc[20] = 0x437bdfd1;
    rc[21] = 0x13428e4f;
    rc[22] = 0x288a0f4a;
    rc[23] = 0x50058122;
    rc[24] = 0x79540962;
    rc[25] = 0x1d8a143d;
    rc[26] = 0x33e39f16;
    rc[27] = 0x47384188;
    rc[28] = 0x2ed3b84c;
    rc[29] = 0x1a119f0a;
    rc[30] = 0x2c54eaff;
    rc[31] = 0x59433c04;
    rc[32] = 0x3d69e498;
    rc[33] = 0x428140c1;
    rc[34] = 0x30c22c73;
    rc[35] = 0x043fca0c;
    rc[36] = 0x1b559a2e;
    rc[37] = 0x708600f0;
    rc[38] = 0x03c07c28;
    rc[39] = 0x17857f68;
    rc[40] = 0x4b7599c4;
    rc[41] = 0x37f14756;
    rc[42] = 0x086ca7f5;
    rc[43] = 0x1c82d864;
    rc[44] = 0x2b692293;
    rc[45] = 0x5ff22dff;
    rc[46] = 0x343f350e;
    rc[47] = 0x7d43d5f3;
    rc[48] = 0x60a632e4;
    rc[49] = 0x6ec93b60;
    rc[50] = 0x1378255c;
    rc[51] = 0x7344d8cb;
    rc[52] = 0x4b371710;
    rc[53] = 0x5d2d8e8e;
    rc[54] = 0x67ce6e1b;
    rc[55] = 0x3fac84ad;
    rc[56] = 0x7cb231e7;
    rc[57] = 0x2b85254b;
    rc[58] = 0x25fe9ebf;
    rc[59] = 0x3b5d415b;
    rc[60] = 0x18fc429f;
    rc[61] = 0x3bb4d4a3;
    rc[62] = 0x06d49266;
    rc[63] = 0x403a9f2f;
    rc[64] = 0x39571c4d;
    rc[65] = 0x7b3c0402;
    rc[66] = 0x6c63f39a;
    rc[67] = 0x3c01d3a3;
    rc[68] = 0x7449846c;
    rc[69] = 0x4c3c3e6d;
    rc[70] = 0x4d1271d7;
    rc[71] = 0x19c13c2d;
    rc[72] = 0x4e19613b;
    rc[73] = 0x7c4852f3;
    rc[74] = 0x3a19f3b1;
    rc[75] = 0x33e9c2ea;
    rc[76] = 0x485cf3b2;
    rc[77] = 0x177dbf36;
    rc[78] = 0x48b3028e;
    rc[79] = 0x088e908f;
    rc[80] = 0x2309a0d7;
    rc[81] = 0x790fbb67;
    rc[82] = 0x10109755;
    rc[83] = 0x2eff1b84;
    rc[84] = 0x4427aaed;
    rc[85] = 0x45f5bdcc;
    rc[86] = 0x7b1326af;
    rc[87] = 0x2993b7b0;
    rc[88] = 0x5c4829bd;
    rc[89] = 0x64f31700;
    rc[90] = 0x57e1b67e;
    rc[91] = 0x75313910;
    rc[92] = 0x086197e9;
    rc[93] = 0x693b1f5d;
    rc[94] = 0x52c91d3a;
    rc[95] = 0x2f8e6f29;
    rc[96] = 0x6c272d73;
    rc[97] = 0x1cb98ba6;
    rc[98] = 0x7ac0cb1f;
    rc[99] = 0x76755656;
    rc[100] = 0x305ae0f0;
    rc[101] = 0x50690167;
    rc[102] = 0x1696e81c;
    rc[103] = 0x029782fc;
    rc[104] = 0x15848d04;
    rc[105] = 0x17976253;
    rc[106] = 0x52370ea3;
    rc[107] = 0x3fef9347;
    rc[108] = 0x6a65c593;
    rc[109] = 0x63b69981;
    rc[110] = 0x25555f4e;
    rc[111] = 0x27d4e26f;
    rc[112] = 0x0b4bfb94;
    rc[113] = 0x1ce72312;
    rc[114] = 0x67d586d3;
    rc[115] = 0x2dd156f1;
    rc[116] = 0x2542e717;
    rc[117] = 0x6163f3b1;
    rc[118] = 0x4d2d0d63;
    rc[119] = 0x019661be;
    rc[120] = 0x01e0830c;
    rc[121] = 0x136f5053;
    rc[122] = 0x1f5dee95;
    rc[123] = 0x088607f6;
    rc[124] = 0x46fa84a5;
    rc[125] = 0x4d401259;
    rc[126] = 0x388b5e6d;
    rc[127] = 0x428b2093;
    rc[128] = 0x7e0215ca;
    rc[129] = 0x3f33237e;
    rc[130] = 0x2fa47615;
    rc[131] = 0x0a923762;
    rc[132] = 0x6469fc2c;
    rc[133] = 0x50de36fc;
    rc[134] = 0x079dd2be;
    rc[135] = 0x5d25d408;
    rc[136] = 0x20a4c417;
    rc[137] = 0x3e919380;
    rc[138] = 0x065d3143;
    rc[139] = 0x10187995;
    rc[140] = 0x22858d82;
    rc[141] = 0x0b55b10e;
    rc[142] = 0x49c14873;
    rc[143] = 0x4dbea407;
    rc[144] = 0x510505cb;
    rc[145] = 0x74604c2c;
    rc[146] = 0x6e12422c;
    rc[147] = 0x31d2bc6f;
    rc[148] = 0x4abc755f;
    rc[149] = 0x213ffed9;
    rc[150] = 0x10864257;
    rc[151] = 0x339c39ef;
    rc[152] = 0x67ac310c;
    rc[153] = 0x603c996b;
    rc[154] = 0x4e95a863;
    rc[155] = 0x2b50485b;
    rc[156] = 0x4d93ea8f;
    rc[157] = 0x1cf81c9a;
    rc[158] = 0x0d9a13da;
    rc[159] = 0x63071071;
    rc[160] = 0x0b03eb46;
    rc[161] = 0x1ff764f8;
    rc[162] = 0x610a2010;
    rc[163] = 0x14dd47f1;
    rc[164] = 0x545afa3f;
    rc[165] = 0x6e3a8913;
    rc[166] = 0x2f362ded;
    rc[167] = 0x0f37ff11;
    rc[168] = 0x67724465;
    rc[169] = 0x3362ad09;
    rc[170] = 0x08ecdd19;
    rc[171] = 0x59c3471f;
    rc[172] = 0x32082f72;
    rc[173] = 0x793d6d25;
    rc[174] = 0x6a086a1f;
    rc[175] = 0x1eb51f40;
    rc[176] = 0x0336115b;
    rc[177] = 0x1de6e380;
    rc[178] = 0x7b6bb725;
    rc[179] = 0x315d3dcf;
    rc[180] = 0x224693c4;
    rc[181] = 0x4f5f6846;
    rc[182] = 0x3e4521f9;
    rc[183] = 0x72a313b1;
    rc[184] = 0x0b3ae1ca;
    rc[185] = 0x5c0be563;
    rc[186] = 0x515bac33;
    rc[187] = 0x11775bb9;
    rc[188] = 0x34cb426f;
    rc[189] = 0x1710dcbd;
    rc[190] = 0x769f178f;
    rc[191] = 0x45bd882f;
    rc[192] = 0x60cadbd6;
    rc[193] = 0x31c0a2a4;
    rc[194] = 0x7968f8fd;
    rc[195] = 0x6a13e997;
    rc[196] = 0x7020de0d;
    rc[197] = 0x680ed11b;
    rc[198] = 0x3c6d11ee;
    rc[199] = 0x6f65fe24;
    rc[200] = 0x26dca7d6;
    rc[201] = 0x1835b260;
    rc[202] = 0x5e9f4edc;
    rc[203] = 0x7c04ee2a;
    rc[204] = 0x1e41f994;
    rc[205] = 0x41f02326;
    rc[206] = 0x67e411aa;
    rc[207] = 0x7cf090a9;
    rc[208] = 0x18a136b5;
    rc[209] = 0x7901be2e;
    rc[210] = 0x1a6ae736;
    rc[211] = 0x06876652;
    rc[212] = 0x47fd6f3b;
    rc[213] = 0x03041342;
    rc[214] = 0x24903949;
    rc[215] = 0x00307f3d;
    rc[216] = 0x02fdbb8a;
    rc[217] = 0x6a70af55;
    rc[218] = 0x20c26749;
    rc[219] = 0x68838a05;
    rc[220] = 0x5cfd89a0;
    rc[221] = 0x12a82dbc;
    rc[222] = 0x1af2ea3f;
    rc[223] = 0x09ebe69a;
    rc[224] = 0x53b0a5f5;
    rc[225] = 0x2fa22433;
    rc[226] = 0x45017aa2;
    rc[227] = 0x4dee2566;
    rc[228] = 0x73bcda76;
    rc[229] = 0x1b2c5604;
    rc[230] = 0x69b8d30d;
    rc[231] = 0x7ad2a178;
    rc[232] = 0x212deab6;
    rc[233] = 0x59865db1;
    rc[234] = 0x165f5250;
    rc[235] = 0x3f74dfdf;
    rc[236] = 0x07c1e51d;
    rc[237] = 0x1b7e9855;
    rc[238] = 0x70daffcc;
    rc[239] = 0x3e673356;
    rc[240] = 0x7ce24cc2;
    rc[241] = 0x2d9238f8;
    rc[242] = 0x1cb6039f;
    rc[243] = 0x4f9baefc;
    rc[244] = 0x43721c99;
    rc[245] = 0x6ce9d61f;
    rc[246] = 0x297ebc1b;
    rc[247] = 0x2a42034a;
    rc[248] = 0x408b899d;
    rc[249] = 0x35248997;
    rc[250] = 0x276a54d5;
    rc[251] = 0x6e4cbe62;
    rc[252] = 0x42e04162;
    rc[253] = 0x31fa07cf;
    rc[254] = 0x50e4aab8;
    rc[255] = 0x14dcd6f6;
    rc[256] = 0x00c4861a;
    rc[257] = 0x12da790e;
    rc[258] = 0x3fa257db;
    rc[259] = 0x078f7c74;
    rc[260] = 0x0e95a5ad;
    rc[261] = 0x1e8a7721;
    rc[262] = 0x0350b631;
    rc[263] = 0x389b8cce;
    rc[264] = 0x50089702;
    rc[265] = 0x5e5b611a;
    rc[266] = 0x2f6e7433;
    rc[267] = 0x31e4feaf;
    rc[268] = 0x73e684cf;
    rc[269] = 0x4a6b0304;
    rc[270] = 0x59af8634;
    rc[271] = 0x05996652;
    rc[272] = 0x1eb24113;
    rc[273] = 0x440e2316;
    rc[274] = 0x7715278f;
    rc[275] = 0x4e0deddf;
    rc[276] = 0x000b13c9;
    rc[277] = 0x6499506e;
    rc[278] = 0x442dc23e;
    rc[279] = 0x786fad2c;
    rc[280] = 0x2260c918;
    rc[281] = 0x0c156d86;
    rc[282] = 0x04cb5854;
    rc[283] = 0x5ba9767b;
    rc[284] = 0x69dc47d0;
    rc[285] = 0x04cf37d8;
    rc[286] = 0x72ce387e;
    rc[287] = 0x2083f38f;
    rc[288] = 0x27305537;
    rc[289] = 0x00e6f4ba;
    rc[290] = 0x2b3e497d;
    rc[291] = 0x640cbbeb;
    rc[292] = 0x40948921;
    rc[293] = 0x1256b32c;
    rc[294] = 0x26c5ff9e;
    rc[295] = 0x49400010;
    rc[296] = 0x6307651d;
    rc[297] = 0x0c0b87d2;
    rc[298] = 0x32f352cf;
    rc[299] = 0x4501e164;
    rc[300] = 0x63d43281;
    rc[301] = 0x6b015892;
    rc[302] = 0x7abe3594;
    rc[303] = 0x444632df;
    rc[304] = 0x390ab06b;
    rc[305] = 0x03867b72;
    rc[306] = 0x5d027ce4;
    rc[307] = 0x0660ef2a;
    rc[308] = 0x6feff36b;
    rc[309] = 0x20cd3bda;
    rc[310] = 0x599fe9a2;
    rc[311] = 0x6c2cb4c6;
    rc[312] = 0x424d1b6e;
    rc[313] = 0x15f2471a;
    rc[314] = 0x6ce12c96;
    rc[315] = 0x42314aec;
    rc[316] = 0x575138c3;
    rc[317] = 0x6d3c7529;
    rc[318] = 0x47a946ab;
    rc[319] = 0x6b17a895;
    rc[320] = 0x6e41d597;
    rc[321] = 0x6f90b0bb;
    rc[322] = 0x1b7251e1;
    rc[323] = 0x3b9e6e2f;
    rc[324] = 0x292de946;
    rc[325] = 0x4747490a;
    rc[326] = 0x35652c49;
    rc[327] = 0x2f40fc84;
    rc[328] = 0x0a297595;
    rc[329] = 0x26ba8663;
    rc[330] = 0x599dc336;
    rc[331] = 0x14fd4bb4;
    rc[332] = 0x2459c6d5;
    rc[333] = 0x6d9172e0;
    rc[334] = 0x0628e5bf;
    rc[335] = 0x778cc2f2;
    rc[336] = 0x76e256b1;
    rc[337] = 0x2ce2681e;
    rc[338] = 0x1a3c639c;
    rc[339] = 0x769d6fe6;
    rc[340] = 0x3c19f53c;
    rc[341] = 0x1ccd5aea;
    rc[342] = 0x3c891a08;
    rc[343] = 0x282843a3;
    rc[344] = 0x1bca2b8e;
    rc[345] = 0x17622ed4;
    rc[346] = 0x5d5f862b;
    rc[347] = 0x797fc339;
    rc[348] = 0x43e7bdff;
    rc[349] = 0x3b4f82af;
    rc[350] = 0x201339b1;
    rc[351] = 0x2e1b080a;
    rc[352] = 0x0d4ae7f9;
    rc[353] = 0x0852e629;
    rc[354] = 0x6228413e;
    rc[355] = 0x50fbe5d1;
    rc[356] = 0x1263f1f7;
    rc[357] = 0x077fdb49;
    rc[358] = 0x0db1a445;
    rc[359] = 0x7b21efcd;
    rc[360] = 0x7c1142fe;
    rc[361] = 0x63a91930;
    rc[362] = 0x0a5f79bf;
    rc[363] = 0x732ae7fb;
    rc[364] = 0x23315cdb;
    rc[365] = 0x2d182a9d;
    rc[366] = 0x2b4bdae3;
    rc[367] = 0x1a509ddb;
    rc[368] = 0x4db8e670;
    rc[369] = 0x4a096555;
    rc[370] = 0x294c0465;
    rc[371] = 0x6f5b70c3;
    rc[372] = 0x45481ff9;
    rc[373] = 0x667f975a;
    rc[374] = 0x5db80b62;
    rc[375] = 0x2919febc;
    rc[376] = 0x0292a214;
    rc[377] = 0x1ed30f83;
    rc[378] = 0x2668dac5;
    rc[379] = 0x241ae0a9;
    rc[380] = 0x41f24663;
    rc[381] = 0x48b93edb;
    rc[382] = 0x2754eba2;
    rc[383] = 0x3c3d6baa;
    rc[384] = 0x47dbc236;
    rc[385] = 0x4eb9f10f;
    rc[386] = 0x5b9c2cd3;
    rc[387] = 0x0e1c9e9e;
    rc[388] = 0x15e8f173;
    rc[389] = 0x2a1646e0;
    rc[390] = 0x21d2fda6;
    rc[391] = 0x274d01af;
    rc[392] = 0x14e82176;
    rc[393] = 0x62525470;
    rc[394] = 0x553d842e;
    rc[395] = 0x360fbea8;
    rc[396] = 0x429a5750;
    rc[397] = 0x62c973dd;
    rc[398] = 0x56d90bea;
    rc[399] = 0x171199d4;
    rc[400] = 0x52321141;
    rc[401] = 0x7c5ccfcc;
    rc[402] = 0x5b7c0e4d;
    rc[403] = 0x5ea1c1e0;
    rc[404] = 0x62c37411;
    rc[405] = 0x18872bb4;
    rc[406] = 0x253db64d;
    rc[407] = 0x23d8bf80;
    rc[408] = 0x6faf33aa;
    rc[409] = 0x5804f05c;
    rc[410] = 0x25d85fb9;
    rc[411] = 0x55798ab9;
    rc[412] = 0x4b050ba8;
    rc[413] = 0x6495f91e;
    rc[414] = 0x2056d156;
    rc[415] = 0x5a7b0c3b;
    rc[416] = 0x51fb5fb8;
    rc[417] = 0x2038c00f;
    rc[418] = 0x5305aa6c;
    rc[419] = 0x744f9bef;
    rc[420] = 0x42a182ea;
    rc[421] = 0x0b3f816b;
    rc[422] = 0x6b2ab968;
    rc[423] = 0x12ce17fa;
    rc[424] = 0x25664d77;
    rc[425] = 0x0deaf12a;
    rc[426] = 0x39e2e25e;
    rc[427] = 0x01f8e4e7;
    rc[428] = 0x75294363;
    rc[429] = 0x29ef597a;
    rc[430] = 0x5cbad414;
    rc[431] = 0x270a712e;
    rc[432] = 0x05d8e787;
    rc[433] = 0x73872668;
    rc[434] = 0x1fb8181c;
    rc[435] = 0x68e8059f;
    rc[436] = 0x11ed2bb3;
    rc[437] = 0x56748f04;
    rc[438] = 0x32918577;
    rc[439] = 0x2ffa9d7a;
    rc[440] = 0x69747e81;
    rc[441] = 0x6f2e7f0c;
    rc[442] = 0x0d2eb326;
    rc[443] = 0x36316c2f;
    rc[444] = 0x25ca6ca8;
    rc[445] = 0x703074e6;
    rc[446] = 0x2d3531d1;
    rc[447] = 0x34914fa7;
}

fn fill_mds_matrix(mds: &mut [u32; N_MDS_VALUES]) {
    mds[0] = 0x07f00000;
    mds[1] = 0x0ef0f0f1;
    mds[2] = 0x23471c72;
    mds[3] = 0x281af287;
    mds[4] = 0x6bf33334;
    mds[5] = 0x42861862;
    mds[6] = 0x1cdd1746;
    mds[7] = 0x2c2c8591;
    mds[8] = 0x59f55556;
    mds[9] = 0x6fc28f5d;
    mds[10] = 0x5ccec4ed;
    mds[11] = 0x41da12f7;
    mds[12] = 0x7164924a;
    mds[13] = 0x5bf72c24;
    mds[14] = 0x47f77778;
    mds[15] = 0x4dd6b5ae;
    mds[16] = 0x10eeeeef;
    mds[17] = 0x07f00000;
    mds[18] = 0x0ef0f0f1;
    mds[19] = 0x23471c72;
    mds[20] = 0x281af287;
    mds[21] = 0x6bf33334;
    mds[22] = 0x42861862;
    mds[23] = 0x1cdd1746;
    mds[24] = 0x2c2c8591;
    mds[25] = 0x59f55556;
    mds[26] = 0x6fc28f5d;
    mds[27] = 0x5ccec4ed;
    mds[28] = 0x41da12f7;
    mds[29] = 0x7164924a;
    mds[30] = 0x5bf72c24;
    mds[31] = 0x47f77778;
    mds[32] = 0x63c92493;
    mds[33] = 0x10eeeeef;
    mds[34] = 0x07f00000;
    mds[35] = 0x0ef0f0f1;
    mds[36] = 0x23471c72;
    mds[37] = 0x281af287;
    mds[38] = 0x6bf33334;
    mds[39] = 0x42861862;
    mds[40] = 0x1cdd1746;
    mds[41] = 0x2c2c8591;
    mds[42] = 0x59f55556;
    mds[43] = 0x6fc28f5d;
    mds[44] = 0x5ccec4ed;
    mds[45] = 0x41da12f7;
    mds[46] = 0x7164924a;
    mds[47] = 0x5bf72c24;
    mds[48] = 0x3a9d89d9;
    mds[49] = 0x63c92493;
    mds[50] = 0x10eeeeef;
    mds[51] = 0x07f00000;
    mds[52] = 0x0ef0f0f1;
    mds[53] = 0x23471c72;
    mds[54] = 0x281af287;
    mds[55] = 0x6bf33334;
    mds[56] = 0x42861862;
    mds[57] = 0x1cdd1746;
    mds[58] = 0x2c2c8591;
    mds[59] = 0x59f55556;
    mds[60] = 0x6fc28f5d;
    mds[61] = 0x5ccec4ed;
    mds[62] = 0x41da12f7;
    mds[63] = 0x7164924a;
    mds[64] = 0x34eaaaab;
    mds[65] = 0x3a9d89d9;
    mds[66] = 0x63c92493;
    mds[67] = 0x10eeeeef;
    mds[68] = 0x07f00000;
    mds[69] = 0x0ef0f0f1;
    mds[70] = 0x23471c72;
    mds[71] = 0x281af287;
    mds[72] = 0x6bf33334;
    mds[73] = 0x42861862;
    mds[74] = 0x1cdd1746;
    mds[75] = 0x2c2c8591;
    mds[76] = 0x59f55556;
    mds[77] = 0x6fc28f5d;
    mds[78] = 0x5ccec4ed;
    mds[79] = 0x41da12f7;
    mds[80] = 0x39ba2e8c;
    mds[81] = 0x34eaaaab;
    mds[82] = 0x3a9d89d9;
    mds[83] = 0x63c92493;
    mds[84] = 0x10eeeeef;
    mds[85] = 0x07f00000;
    mds[86] = 0x0ef0f0f1;
    mds[87] = 0x23471c72;
    mds[88] = 0x281af287;
    mds[89] = 0x6bf33334;
    mds[90] = 0x42861862;
    mds[91] = 0x1cdd1746;
    mds[92] = 0x2c2c8591;
    mds[93] = 0x59f55556;
    mds[94] = 0x6fc28f5d;
    mds[95] = 0x5ccec4ed;
    mds[96] = 0x58e66667;
    mds[97] = 0x39ba2e8c;
    mds[98] = 0x34eaaaab;
    mds[99] = 0x3a9d89d9;
    mds[100] = 0x63c92493;
    mds[101] = 0x10eeeeef;
    mds[102] = 0x07f00000;
    mds[103] = 0x0ef0f0f1;
    mds[104] = 0x23471c72;
    mds[105] = 0x281af287;
    mds[106] = 0x6bf33334;
    mds[107] = 0x42861862;
    mds[108] = 0x1cdd1746;
    mds[109] = 0x2c2c8591;
    mds[110] = 0x59f55556;
    mds[111] = 0x6fc28f5d;
    mds[112] = 0x468e38e4;
    mds[113] = 0x58e66667;
    mds[114] = 0x39ba2e8c;
    mds[115] = 0x34eaaaab;
    mds[116] = 0x3a9d89d9;
    mds[117] = 0x63c92493;
    mds[118] = 0x10eeeeef;
    mds[119] = 0x07f00000;
    mds[120] = 0x0ef0f0f1;
    mds[121] = 0x23471c72;
    mds[122] = 0x281af287;
    mds[123] = 0x6bf33334;
    mds[124] = 0x42861862;
    mds[125] = 0x1cdd1746;
    mds[126] = 0x2c2c8591;
    mds[127] = 0x59f55556;
    mds[128] = 0x0fe00000;
    mds[129] = 0x468e38e4;
    mds[130] = 0x58e66667;
    mds[131] = 0x39ba2e8c;
    mds[132] = 0x34eaaaab;
    mds[133] = 0x3a9d89d9;
    mds[134] = 0x63c92493;
    mds[135] = 0x10eeeeef;
    mds[136] = 0x07f00000;
    mds[137] = 0x0ef0f0f1;
    mds[138] = 0x23471c72;
    mds[139] = 0x281af287;
    mds[140] = 0x6bf33334;
    mds[141] = 0x42861862;
    mds[142] = 0x1cdd1746;
    mds[143] = 0x2c2c8591;
    mds[144] = 0x48924925;
    mds[145] = 0x0fe00000;
    mds[146] = 0x468e38e4;
    mds[147] = 0x58e66667;
    mds[148] = 0x39ba2e8c;
    mds[149] = 0x34eaaaab;
    mds[150] = 0x3a9d89d9;
    mds[151] = 0x63c92493;
    mds[152] = 0x10eeeeef;
    mds[153] = 0x07f00000;
    mds[154] = 0x0ef0f0f1;
    mds[155] = 0x23471c72;
    mds[156] = 0x281af287;
    mds[157] = 0x6bf33334;
    mds[158] = 0x42861862;
    mds[159] = 0x1cdd1746;
    mds[160] = 0x69d55556;
    mds[161] = 0x48924925;
    mds[162] = 0x0fe00000;
    mds[163] = 0x468e38e4;
    mds[164] = 0x58e66667;
    mds[165] = 0x39ba2e8c;
    mds[166] = 0x34eaaaab;
    mds[167] = 0x3a9d89d9;
    mds[168] = 0x63c92493;
    mds[169] = 0x10eeeeef;
    mds[170] = 0x07f00000;
    mds[171] = 0x0ef0f0f1;
    mds[172] = 0x23471c72;
    mds[173] = 0x281af287;
    mds[174] = 0x6bf33334;
    mds[175] = 0x42861862;
    mds[176] = 0x32cccccd;
    mds[177] = 0x69d55556;
    mds[178] = 0x48924925;
    mds[179] = 0x0fe00000;
    mds[180] = 0x468e38e4;
    mds[181] = 0x58e66667;
    mds[182] = 0x39ba2e8c;
    mds[183] = 0x34eaaaab;
    mds[184] = 0x3a9d89d9;
    mds[185] = 0x63c92493;
    mds[186] = 0x10eeeeef;
    mds[187] = 0x07f00000;
    mds[188] = 0x0ef0f0f1;
    mds[189] = 0x23471c72;
    mds[190] = 0x281af287;
    mds[191] = 0x6bf33334;
    mds[192] = 0x1fc00000;
    mds[193] = 0x32cccccd;
    mds[194] = 0x69d55556;
    mds[195] = 0x48924925;
    mds[196] = 0x0fe00000;
    mds[197] = 0x468e38e4;
    mds[198] = 0x58e66667;
    mds[199] = 0x39ba2e8c;
    mds[200] = 0x34eaaaab;
    mds[201] = 0x3a9d89d9;
    mds[202] = 0x63c92493;
    mds[203] = 0x10eeeeef;
    mds[204] = 0x07f00000;
    mds[205] = 0x0ef0f0f1;
    mds[206] = 0x23471c72;
    mds[207] = 0x281af287;
    mds[208] = 0x54aaaaab;
    mds[209] = 0x1fc00000;
    mds[210] = 0x32cccccd;
    mds[211] = 0x69d55556;
    mds[212] = 0x48924925;
    mds[213] = 0x0fe00000;
    mds[214] = 0x468e38e4;
    mds[215] = 0x58e66667;
    mds[216] = 0x39ba2e8c;
    mds[217] = 0x34eaaaab;
    mds[218] = 0x3a9d89d9;
    mds[219] = 0x63c92493;
    mds[220] = 0x10eeeeef;
    mds[221] = 0x07f00000;
    mds[222] = 0x0ef0f0f1;
    mds[223] = 0x23471c72;
    mds[224] = 0x3f800000;
    mds[225] = 0x54aaaaab;
    mds[226] = 0x1fc00000;
    mds[227] = 0x32cccccd;
    mds[228] = 0x69d55556;
    mds[229] = 0x48924925;
    mds[230] = 0x0fe00000;
    mds[231] = 0x468e38e4;
    mds[232] = 0x58e66667;
    mds[233] = 0x39ba2e8c;
    mds[234] = 0x34eaaaab;
    mds[235] = 0x3a9d89d9;
    mds[236] = 0x63c92493;
    mds[237] = 0x10eeeeef;
    mds[238] = 0x07f00000;
    mds[239] = 0x0ef0f0f1;
    mds[240] = 0x7f000000;
    mds[241] = 0x3f800000;
    mds[242] = 0x54aaaaab;
    mds[243] = 0x1fc00000;
    mds[244] = 0x32cccccd;
    mds[245] = 0x69d55556;
    mds[246] = 0x48924925;
    mds[247] = 0x0fe00000;
    mds[248] = 0x468e38e4;
    mds[249] = 0x58e66667;
    mds[250] = 0x39ba2e8c;
    mds[251] = 0x34eaaaab;
    mds[252] = 0x3a9d89d9;
    mds[253] = 0x63c92493;
    mds[254] = 0x10eeeeef;
    mds[255] = 0x07f00000;
}

// === permutation ===

fn add_round_constants(
    state: &mut [u32; STATE_WIDTH],
    rc: &[u32; N_ROUND_CONSTANTS],
    round: usize,
) {
    let base = round * STATE_WIDTH;
    let mut i = 0;
    while i < STATE_WIDTH {
        state[i] = fadd(state[i], rc[base + i]);
        i += 1;
    }
}

fn apply_mds(state: &mut [u32; STATE_WIDTH], mds: &[u32; N_MDS_VALUES]) {
    let mut tmp = [0u32; STATE_WIDTH];
    let mut i = 0;
    while i < STATE_WIDTH {
        let mut acc: u32 = 0;
        let mut j = 0;
        while j < STATE_WIDTH {
            acc = fadd(acc, fmul(mds[i * STATE_WIDTH + j], state[j]));
            j += 1;
        }
        tmp[i] = acc;
        i += 1;
    }
    *state = tmp;
}

fn full_round(
    state: &mut [u32; STATE_WIDTH],
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
    round: usize,
) {
    add_round_constants(state, rc, round);
    let mut i = 0;
    while i < STATE_WIDTH {
        state[i] = fpow_alpha(state[i]);
        i += 1;
    }
    apply_mds(state, mds);
}

fn partial_round(
    state: &mut [u32; STATE_WIDTH],
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
    round: usize,
) {
    add_round_constants(state, rc, round);
    state[0] = fpow_alpha(state[0]);
    apply_mds(state, mds);
}

fn permutation(
    state: &mut [u32; STATE_WIDTH],
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
) {
    let mut r = 0;
    while r < FULL_ROUNDS_HALF {
        full_round(state, rc, mds, r);
        r += 1;
    }
    while r < FULL_ROUNDS_HALF + PRTL_ROUNDS {
        partial_round(state, rc, mds, r);
        r += 1;
    }
    while r < TOTAL_ROUNDS {
        full_round(state, rc, mds, r);
        r += 1;
    }
}

// === sponge hash ===

#[inline(always)]
fn read_full_felt(input: &[u8], offset: usize) -> u32 {
    // 3 bytes little-endian (matches get_full_felt in poseidon.zkc).
    (input[offset] as u32)
        | ((input[offset + 1] as u32) << 8)
        | ((input[offset + 2] as u32) << 16)
}

fn read_partial_felt(input: &[u8], offset: usize, n_bytes: usize) -> u32 {
    // matches get_partial_felt in poseidon.zkc: builds the felt by Horner
    // accumulation starting from the high byte.
    let mut v: u32 = 0;
    let mut i = n_bytes;
    while i > 0 {
        v = fmul(v, 256);
        v = fadd(v, input[offset + i - 1] as u32);
        i -= 1;
    }
    v
}

fn sponge_hash(
    input: &[u8],
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
) -> [u32; STATE_WIDTH] {
    let mut state = [0u32; STATE_WIDTH];

    let size = input.len();

    // initialize_state: state[RATE] = number_of_felts (rounded up)
    let number_of_felts = (size + BYTES_PER_INPUT_FELT - 1) / BYTES_PER_INPUT_FELT;
    state[RATE] = number_of_felts as u32;

    if size == 0 {
        // matches the precompile's `if size_in_bytes == 0 { fail }`
        exit(2);
    }

    let n_blocks = (size + BYTES_PER_BLOCK - 1) / BYTES_PER_BLOCK;
    let final_block_size = size % BYTES_PER_BLOCK;
    let final_block_is_partial = final_block_size != 0;

    let mut offset: usize = 0;
    let mut block_idx: usize = 0;
    while block_idx < n_blocks {
        let block_size = if block_idx == n_blocks - 1 && final_block_is_partial {
            final_block_size
        } else {
            BYTES_PER_BLOCK
        };

        let quotient = block_size / BYTES_PER_INPUT_FELT;
        let remainder = block_size % BYTES_PER_INPUT_FELT;

        let n_full_felts = if quotient < RATE { quotient } else { RATE };
        let with_partial = remainder != 0;

        let mut i = 0;
        while i < n_full_felts {
            let felt = read_full_felt(input, offset);
            state[i] = fadd(state[i], felt);
            offset += BYTES_PER_INPUT_FELT;
            i += 1;
        }

        if with_partial {
            let felt = read_partial_felt(input, offset, remainder);
            state[n_full_felts] = fadd(state[n_full_felts], felt);
            offset += remainder;
        }

        permutation(&mut state, rc, mds);
        block_idx += 1;
    }

    state
}

// === test runners ===

fn run_range<const N: usize>(
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
) -> [u32; STATE_WIDTH]
where
    [(); N * BYTES_PER_INPUT_FELT]: Sized,
{
    let mut input = [0u8; N * BYTES_PER_INPUT_FELT];
    let mut i = 0;
    while i < N {
        input[i * BYTES_PER_INPUT_FELT]     = (i         & 0xff) as u8;
        input[i * BYTES_PER_INPUT_FELT + 1] = ((i >> 8)  & 0xff) as u8;
        input[i * BYTES_PER_INPUT_FELT + 2] = ((i >> 16) & 0xff) as u8;
        i += 1;
    }
    sponge_hash(&input, rc, mds)
}

fn run_zeros<const N: usize>(
    rc: &[u32; N_ROUND_CONSTANTS],
    mds: &[u32; N_MDS_VALUES],
) -> [u32; STATE_WIDTH]
where
    [(); N * BYTES_PER_INPUT_FELT]: Sized,
{
    let input = [0u8; N * BYTES_PER_INPUT_FELT];
    sponge_hash(&input, rc, mds)
}

// === entry point ===

// Same 6 active cases as poseidon1_with_in_embedded.rs so cycle counts are
// directly comparable between this binary and the precompile variant. The
// `core::hint::black_box` calls keep the optimizer from dead-code-eliminating
// the test runs.
//
// The 1 << 16 case takes a long time so is not covered
#[no_mangle]
fn main() -> ! {
    let mut rc = [0u32; N_ROUND_CONSTANTS];
    fill_round_constants(&mut rc);

    let mut mds = [0u32; N_MDS_VALUES];
    fill_mds_matrix(&mut mds);

    core::hint::black_box(run_range::<7>(&rc, &mds));
    core::hint::black_box(run_range::<16>(&rc, &mds));
    core::hint::black_box(run_range::<256>(&rc, &mds));
    core::hint::black_box(run_zeros::<1>(&rc, &mds));
    core::hint::black_box(run_zeros::<16>(&rc, &mds));
    core::hint::black_box(run_zeros::<256>(&rc, &mds));
    // commented out for the same reason as in the embedded variant:
    // core::hint::black_box(run_zeros::<{ 1 << 16 }>(&rc, &mds));
    // core::hint::black_box(run_zeros::<{ 1 << 18 }>(&rc, &mds));
    // core::hint::black_box(run_zeros::<{ 1 << 20 }>(&rc, &mds));

    exit(0);
}
