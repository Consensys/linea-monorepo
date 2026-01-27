# Contract Integrity Verifier UI

A web-based user interface for the Contract Integrity Verifier. This tool allows you to verify deployed smart contracts against local artifacts through an intuitive browser-based interface.

## Features

- **Adapter Selection**: Choose between Ethers.js or Viem as your Web3 library
- **Config File Support**: Upload JSON or Markdown configuration files
- **Auto-detection**: Automatically detects required schema and artifact files from your config
- **Dynamic Forms**: Auto-generates input fields for environment variables based on config
- **Verification Options**: Full control over verification flags (skip bytecode, ABI, state checks)
- **Results Display**: Clear, expandable results for each contract verification
- **Session Persistence**: Sessions are stored locally and persist across page reloads

## Quick Start

```bash
# From the verifier-ui directory
pnpm install
pnpm dev
```

Open [http://localhost:3000](http://localhost:3000) in your browser.

## Usage

### 1. Select Web3 Library

Choose between **Viem** (recommended) or **Ethers** based on your preference.

### 2. Upload Configuration

Upload your verifier configuration file. Supported formats:

- **JSON**: Standard JSON configuration format
- **Markdown**: Markdown format with `verifier` code blocks

Example JSON config:

```json
{
  "chains": {
    "mainnet": {
      "chainId": 1,
      "rpcUrl": "${MAINNET_RPC_URL}",
      "explorerUrl": "https://etherscan.io"
    }
  },
  "contracts": [
    {
      "name": "MyContract",
      "chain": "mainnet",
      "address": "0x...",
      "artifactFile": "../artifacts/MyContract.json",
      "stateVerification": {
        "schemaFile": "../schemas/my-contract.json"
      }
    }
  ]
}
```

### 3. Upload Required Files

The UI will detect and list all schema and artifact files referenced in your config. Upload each required file.

### 4. Fill Environment Variables

The UI auto-detects environment variable placeholders (e.g., `${MAINNET_RPC_URL}`) and generates appropriate input fields. Fill in all required values.

### 5. Configure Options

Optionally configure verification options:

- **Verbose output**: Show detailed verification information
- **Skip bytecode**: Skip bytecode comparison
- **Skip ABI**: Skip ABI selector verification  
- **Skip state**: Skip state verification checks
- **Contract filter**: Only verify specific contracts
- **Chain filter**: Only verify contracts on specific chains

### 6. Run Verification

Click **Run Verification** to start. Results display in expandable cards showing:

- Overall status (pass/fail/warn/skip)
- Bytecode verification results
- ABI verification results
- State verification results (view calls, slots, storage paths)

## Configuration

Environment variables can be set in a `.env.local` file:

```bash
# Directory for storing uploaded files (defaults to ~/.linea-verifier-ui)
VERIFIER_UPLOADS_DIR=/path/to/uploads

# Session expiry in hours (defaults to 24)
SESSION_EXPIRY_HOURS=24

# Maximum file size in bytes (defaults to 10MB)
MAX_FILE_SIZE=10485760

# Maximum total session size in bytes (defaults to 50MB)
MAX_SESSION_SIZE=52428800
```

## File Storage

Uploaded files are stored outside the repository to prevent codebase pollution:

- Default location: `~/.linea-verifier-ui/sessions/<session-id>/`
- Sessions expire after 24 hours (configurable)
- Each session has isolated storage for config, schemas, and artifacts

## Development

```bash
# Run development server
pnpm dev

# Type checking
pnpm check-types

# Linting
pnpm lint

# Build for production
pnpm build

# Start production server
pnpm start
```

## Architecture

```
verifier-ui/
├── src/
│   ├── app/              # Next.js App Router
│   │   ├── api/          # API routes (session, upload, verify)
│   │   ├── layout.tsx    # Root layout
│   │   └── page.tsx      # Main page
│   ├── components/       # React components
│   │   ├── ui/           # Reusable UI components
│   │   └── *-section/    # Page sections
│   ├── hooks/            # React hooks
│   ├── lib/              # Utilities
│   │   ├── api.ts        # API client
│   │   ├── config-parser.ts  # Config parsing
│   │   ├── session.ts    # Session management
│   │   └── validation.ts # Zod schemas
│   ├── stores/           # Zustand stores
│   ├── types/            # TypeScript types
│   └── scss/             # Global styles
```

## Dependencies

This package depends on:

- `@consensys/linea-contract-integrity-verifier` (core)
- `@consensys/linea-contract-integrity-verifier-ethers`
- `@consensys/linea-contract-integrity-verifier-viem`

Both verifier adapters are included to support user choice at runtime.

## License

MIT OR Apache-2.0
