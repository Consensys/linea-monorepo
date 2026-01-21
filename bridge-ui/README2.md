# Linea Bridge UI (temporary README draft)

Frontend for the Linea Bridge, built with Next.js.

## Live sites
- Production: https://bridge.linea.build/

## Scope of this README
- How to run and test the app locally
- How CI produces Docker image tags
- A high-level internal deployment workflow

## Local development

### Requirements
- Node.js LTS (18/20)
- pnpm

### Setup
1) Create a local env file:
```sh
cp .env.template .env
```
2) Edit `.env` and add any required API keys for local development.

### Run the dev server
```sh
pnpm i
pnpm run dev
```
Open: http://localhost:3000

## Configuration

### Public vs private variables
- Public configuration variables are read from `.env.production` (and from your local `.env` while developing).
- Private configuration variables used by CI/CD are stored in GitHub Actions secrets.
- Important: any variable prefixed with `NEXT_PUBLIC_` is exposed to the browser; do not put secrets there.

### Config variables
Reference list: `.env.template`, including:
- Contract addresses (mainnet and sepolia)
- Token list URLs
- Third-party API keys (WalletConnect, Infura, Alchemy, QuickNode, etc.)
- Feature flags (e.g., CCTP)
- For the full variable table, see the existing `README.md` config section (unchanged).

## Build and run the Docker image locally
This matches the Docker image used for deployments.

### Build
```sh
docker build --build-arg ENV_FILE=.env.production -t linea/bridge-ui .
```
Notes:
- The file passed via `ENV_FILE` must exist in the build context.
- Use `ENV_FILE=.env.production` for production-like config.

### Run
```sh
docker run -p 3000:3000 linea/bridge-ui
# optional: mount a specific env file
# docker run -p 3000:3000 -v $(pwd)/.env.production:/app/.env.production linea/bridge-ui
```
Open: http://localhost:3000

## CI and Docker image tags
- Docker image tags are produced by the GitHub Actions workflow “Bridge UI Build and Publish”.

### How to retrieve a tag
1) Open the latest successful workflow run for the Bridge UI build/publish pipeline.
2) Find the step named “Set Docker Tag”.
3) Copy the printed `DOCKER_TAG` value.

### Tag format
```
<GITHUB_SHA_7>-<unix_timestamp>-bridge-ui-<package_version>
```
Example: `f3afe33-1705598198-bridge-ui-0.5.3`

## End-to-end tests
E2E tests run in CI and can also be run locally.

### Run locally (from repo root)
1) In your `.env`, set: `NEXT_PUBLIC_E2E_TEST_MODE=true`
2) Build the app:
```sh
pnpm run build
```
3) From the repository root, start the local docker stack:
```sh
make start-env-with-tracing-v2-ci
```
4) Run the tests (still from repo root):
```sh
pnpm run test:e2e:headful
```

## Deployment (internal ops)
Deployments are managed internally via ArgoCD and internal deployment repositories.

High-level process:
1) Get a Bridge UI image tag from CI (see CI and Docker image tags above).
2) Update the Bridge UI image tag in the internal deployment values file (example path: `argocd/bridge-ui/values.yaml`).
3) Open a PR/MR, merge, then sync/apply via ArgoCD.

Notes:
- Internal repositories and ArgoCD dashboards are intentionally not linked here to avoid broken links for external readers.
- Access details are managed internally.
