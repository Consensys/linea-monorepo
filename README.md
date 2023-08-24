# Linea zkEVM
This is a monorepo for Linea, the Consensys zkEVM network.

- Homepage: https://linea.build/
- Docs: https://docs.linea.build/
- Mirror.xyz: https://linea.mirror.xyz/
- Support: https://support.linea.build
- Twitter: https://twitter.com/LineaBuild


## Getting Started

Requirements:
- Node.js v18
- Java 17


```
cd contracts
npm install
cd ..

// start stack locally
make fresh-start-all


cd e2e
npm install
npm run test:e2e:local
```

You observe files being generated in `tmp/local/` directory.
```
├── local
│  ├── prover
│  │  ├── request
│  │  │  └── 4-4-etv0.2.0-stv1.2.0-getZkProof.json
│  │  ├── requests-done
│  │  │  ├── 1-1-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  │  ├── 2-2-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  │  └── 3-3-etv0.2.0-stv1.2.0-getZkProof.json.success
│  │  └── response
│  │      ├── 1-1-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  │      ├── 2-2-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  │      └── 3-3-etv0.2.0-stv1.2.0-getZkProof.json.0.2.0.json
│  └── traces
│      ├── conflated
│      │  ├── 1-1.conflated.v0.2.0.json.gz
│      │  ├── 2-2.conflated.v0.2.0.json.gz
│      │  ├── 3-3.conflated.v0.2.0.json.gz
│      │  └── 4-4.conflated.v0.2.0.json.gz
│      ├── raw
│      │  ├── 1-0x2e1a3f506c0d5f11310301a86f608d840d3db0e28c545eaf9e9c9812e2b795e0.v0.2.0.json.gz
│      │  ├── 2-0x3e5b3bd8e21a94488bf93776480271d3fef8033152effd4e19fe6519dea53379.v0.2.0.json.gz
│      │  ├── 3-0xa5046c13502a619a7a3f091b397234dc020f6cbda1942d247d1003d4c73899b6.v0.2.0.json.gz
│      │  ├── 4-0xe9203ede2114bf9c291692c4bd2dcc7207973c267ed411d65568d1138b3ecfcc.v0.2.0.json.gz
│      │  ├── 5-0x2c8ec07d4222bed8285be3de83f0fccc989134c49826baed5340bf7aa8e3ce8f.v0.2.0.json.gz
│      │  └── 6-0x3c7b7ee369d5fe02a6865415a2d0ef4ec385812351723e35a3b54d972f9f4ceb.v0.2.0.json.gz
│      └── raw-non-canonical
```

## Tuning in conflation
For local testing development conflation deadline is set to 6s `conflation-deadline=PT6S` in `config/coordinator/coordinator-docker.config.toml` file. Hence only 2 blocks only conflation.
