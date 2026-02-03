# Linea Sequencer and Tracer Releases

This document describes the changes introduced to streamline **linea-sequencer** and **linea-tracer** releases, including a dedicated version catalog, automatic releases on merge to `main`, and local and CI support for building the linea-besu-package image with locally built tracer and sequencer plugins.

---

## 1. Introduction of `releases.versions.toml`

A separate Gradle version catalog file **`gradle/releases.versions.toml`** was introduced to hold release versions for the tracer (arithmetization) and sequencer plugins, distinct from the main dependency catalog `gradle/libs.versions.toml`.

### Purpose

- **Single source of truth** for tracer and sequencer release versions used by:
  - Gradle builds (tracer, besu-plugins/linea-sequencer)
  - CI workflows (auto-release-tracer-and-sequencer, build-from-source, e2e)
  - Local linea-besu-package Makefile

- **Trigger workflows** when tracer and sequencer release versions have been updated:
  - Build a **linea-besu-package** Docker image using **locally built** tracer and sequencer plugins (from source) with the desired besu version defined in `libs.versions.toml` and runs e2e tests against that image
  - Automatically trigger the release workflows of tracer and sequencer using the release versions `releases.versions.toml` when a PR merged into `main`

### File format and location

- **Path:** `gradle/releases.versions.toml`
- **Example content:** (Same versioning semantics as in `linea-besu-package/versions.env`)

  ```toml
  [versions]
  arithmetization = "beta-v4.4-rc7.4"
  sequencer = "4.4-rc7.4.1"
  ```

- **Gradle usage:** The `releases` catalog is registered in **root `settings.gradle`** so that all projects (including `:tracer` and `besu-plugins:linea-sequencer`) can reference it, e.g. `releases.versions.arithmetization.get()`.

---

## 2. Linea-Besu-Package Image with Local-Built Tracer and Sequencer in E2E (when `releases.versions.toml` changes)

When changes affect **`gradle/releases.versions.toml`**, the CI pipeline builds the **linea-besu-package** Docker image using **locally built** tracer and sequencer plugins (from source) with the desired besu version in `libs.versions.toml` and runs e2e tests against that image. This validates integration at PR time instead of only at release time.

---

## 3. Auto-release of Linea Sequencer and Linea Tracer on Merge to Main

When a PR that updates **`gradle/releases.versions.toml`** is merged into `main`, the tracer and sequencer release workflows are triggered automatically with the versions from that file.

### Workflow

- **Name:** `Auto-release Tracer and Sequencer on releases.versions.toml change`
- **File:** `.github/workflows/linea-tracer-and-sequencer-auto-release.yml`
- **Trigger:** `push` to branch `main`, with path filter `gradle/releases.versions.toml`

### Behavior

1. On push to `main`, if the commit includes changes to `gradle/releases.versions.toml`, the workflow runs.
2. The **parse-release-versions** action reads `arithmetization` and `sequencer` versions from `gradle/releases.versions.toml`.
3. **Linea Tracer release** is triggered:
   - Workflow: `linea-tracer-plugin-release.yml`
   - Input: `release_tag=linea-tracer-<arithmetization_version>`
4. **Linea Sequencer release** is triggered:
   - Workflow: `linea-sequencer-plugin-release.yml`
   - Inputs: `release_tag=linea-sequencer-<sequencer_version>`

No manual workflow dispatch is required to release tracer or/and sequence if the release versions are updated in `releases.versions.toml` in the PR that merged into `main`. Duplicated releases of same versions will be failed.

### Supporting piece

- **`.github/actions/parse-release-versions`** – composite action that parses `gradle/releases.versions.toml` and outputs `arithmetization_version` and `sequencer_version` for use by the auto-release and other workflows.

---

## 4. Linea-Besu-Package Makefile: Local Build with Local Tracer and Sequencer

The **`linea-besu-package/Makefile`** was revised so one can build the linea-besu-package image **locally** using tracer and sequencer plugins built from source in this repo, with versions read from `gradle/releases.versions.toml`.

### Main targets

| Target | Description |
|--------|-------------|
| **`build-tracer-and-sequencer-local`** | From repo root: runs `./gradlew :tracer:artifacts` and `./gradlew besu-plugins:linea-sequencer:sequencer:artifacts` with versions from `releases.versions.toml`. |
| **`assemble`** | Assembles the package using downloaded artifacts (current `versions.env` and assemble script). |
| **`assemble-local`** | Assembles using local tracer and sequencer zips (paths derived from `LOCAL_*_ZIP_FOLDER` and versions above). |
| **`build`** | `assemble` then `build-image`. |
| **`build-local`** | `build-tracer-and-sequencer-local` → `assemble-local` → `build-image`. Builds tracer and sequencer from source, then builds the Docker image with those plugins. |
| **`run-e2e-test`** | Runs e2e tests with the coordinator config’s expected traces API version set from `EXPECTED_TRACES_API_VERSION`. |
| **`run-e2e-test-with-arithmetization-version`** | Sets `EXPECTED_TRACES_API_VERSION=$(ARITHMETIZATION_VERSION)` and runs `run-e2e-test`. |
| **`run-e2e-test-cleanup`** | Cleans up the e2e environment. |

### Example: full local build and e2e

From the repository root:

```bash
cd linea-besu-package
make build-local
make run-e2e-test-with-arithmetization-version
# when done
make run-e2e-test-cleanup
```

This builds tracer and sequencer from source (using `releases.versions.toml`), assembles the package with those zips, builds the Docker image, and runs e2e with the arithmetization version as the expected traces API version.

### Overrides

- **`TAG`** – Docker image tag (default: `local`).
- **`PLATFORM`** – Optional platform for `docker buildx` (e.g. `linux/amd64`).
- **`LOCAL_TRACER_ZIP_FOLDER`**, **`LOCAL_SEQUENCER_ZIP_FOLDER`** – Override if your zips are in a different location.
- **`EXPECTED_TRACES_API_VERSION`** - Override for a specific expected traces-api version used by Coordinator

---

## 5. Linea-besu-package release flow examples

### When linea-besu-upstream version needs to be updated

1. Create a PR to update the `besu` field in [libs.versions.toml](../gradle/libs.versions.toml)
2. Make necessary changes in tracer and seqeuencer to accommodate the besu version change
3. Once the testing workflows of tracer and sequencer are all passed, update the cooresponding versions in [releases.versions.toml](../gradle/releases.versions.toml) (Please note that there's a rule on sequencer versioning, see this [doc](https://www.notion.so/consensys/Besu-4-Linea-HowTo-Release-4ae7ea9b2f0e4ad29a7c7485c2571336#237fc61a326e8064b57ae37797163987))
4. The CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins (from source) and runs e2e tests against that image
5. When the PR merged into `main`, the releases of tracer and sequencer will be triggered automatically
6. Create a separate PR with the updated versions (i.e. `LINEA_BESU_TAR_GZ`, `LINEA_SEQUENCER_PLUGIN_VERSION`, and `LINEA_TRACER_PLUGIN_VERSION`) in [versions.env](../linea-besu-package/versions.env) to make a new release of `linea-besu-package` (see this [README](../linea-besu-package/README.md#how-to-release) for more details)

### When arithmetization version needs to be updated

1. Create a PR to make necessary changes in tracer (and seqeuencer if needed)
2. Once the testing workflows of tracer and sequencer are all passed, update the cooresponding versions in [releases.versions.toml](../gradle/releases.versions.toml) (Please note that sequencer version will need to be updated even if no changes on sequencer)
3. The CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins (from source) and runs e2e tests against that image
4. When the PR merged into `main`, the releases of tracer and sequencer will be triggered automatically
5. Create a separate PR with the updated versions (i.e. `LINEA_SEQUENCER_PLUGIN_VERSION` and `LINEA_TRACER_PLUGIN_VERSION`) in [versions.env](../linea-besu-package/versions.env) to make a new release of `linea-besu-package` (see this [README](../linea-besu-package/README.md#how-to-release) for more details)

### When hotfix of linea-besu-upstream is needed on a production image

1. Create a hotfix PR based on the **same commit tag** as the production image, and update the `besu` field in [libs.versions.toml](../gradle/libs.versions.toml)
2. Make necessary changes in tracer and seqeuencer to accommodate the besu version change
2. Once the testing workflows of tracer and sequencer are all passed, update the cooresponding hotfix versions in [releases.versions.toml](../gradle/releases.versions.toml)
3. The CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins (from source) and runs e2e tests against that image
4. Trigger manual release workflow for tracer and sequencer with the hotfix versions
5. Update the hotfix PR with the updated versions (i.e. `LINEA_BESU_TAR_GZ`, `LINEA_SEQUENCER_PLUGIN_VERSION`, and `LINEA_TRACER_PLUGIN_VERSION`) in [versions.env](../linea-besu-package/versions.env) to make a hotfix release of `linea-besu-package`

---

## 6. Related documentations

- [Linea-besu-package README](./linea-besu-package/versions.env) – how to release the full linea-besu-package and use `versions.env` for non-tracer/sequencer plugins.
- [Linea-besu-package HowTo/Release Notion doc](https://www.notion.so/consensys/Linea-besu-package-HowTo-Release-2effc61a326e80f2b384cbd635e507c6) - current release process of linea-besu-package in face of besu version update
