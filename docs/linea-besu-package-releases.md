# Linea Besu Package Release

This document describes the changes introduced to streamline **linea-sequencer** and **linea-tracer** releases, including a dedicated version catalog, automatic releases on merge to `main`, and local and CI support for building the linea-besu-package image with locally built tracer and sequencer plugins.

---

## 1. Introduction of `releases.versions.toml`

A separate Gradle version catalog file **`gradle/releases.versions.toml`** was introduced to hold release version for the tracer (arithmetization) plugin and the commit hash tag from `hyperledger/besu` for tracer and sequencer plugins to build against with.

### File format and location

- **Path:** `gradle/releases.versions.toml`
- **Example content:**

  ```toml
  [versions]
  arithmetization = "beta-v5.0-rc2"
  besuCommit = "d68679d6887abf8c407f99456cf2502004fa2bad"
  ```

### Purpose

- **Single source of truth** for tracer and sequencer release version:
  - Tracer artifacts will be named as `linea-tracer-[arithmetization]-[7-char-commit-hash].*`: e.g. `linea-tracer-beta-v5.0-rc2-9f55c53.zip`, `linea-tracer-beta-v5.0-rc2-9f55c53.jar`, `arithmetization-beta-v5.0-rc2-9f55c53.jar`, etc.
  - Sequencer artifact will be named as `linea-sequencer-[arithmetization]-[7-char-commit-hash].*`: e.g. `linea-sequencer-beta-v5.0-rc2-9f55c53.zip`, `linea-sequencer-beta-v5.0-rc2-9f55c53.jar`, etc.
  - Besu artifact will be named as `besu-[release-tag]-linea-[7-char-besu-commit-hash].tar.gz`: e.g. `besu-25.12.0-linea-d68679d.tar.gz` where `25.12.0` is the latest release tag on `hyperleger/besu` at the given `besuCommit`, the artifact will be placed under the `tmp/hyperledger-besu/build/distributions/folder` and the `besu` field in `libs.version.toml` will be automatically updated, e.g. `besu = 25.12.0-linea-d68679d`

- **Trigger workflows** when the desired besu commit, tracer/sequencer, or other plugins have been updated:
  - Build a **linea-besu-package** Docker image using **locally built** tracer and sequencer plugins with the desired besu version defined in `libs.versions.toml` (will build it from source if the besu version were not found in maven) and runs e2e tests against that image
  - Automatically trigger the maven publishing workflow for linea-besu if the desired besu version defined in `libs.versions.toml` was not found in maven (during pull requests only)
  - Automatically trigger the release workflow of linea-besu-package with **locally built** tracer and sequencer plugins when the PR merged into `main`

- **Gradle usage:** The `releases` catalog is registered in **root `settings.gradle`** so that all projects (including `:tracer` and `besu-plugins:linea-sequencer`) can reference it, e.g. `releases.versions.arithmetization.get()`.

---

## 2. Linea-Besu-Package Image with Local-Built Tracer and Sequencer in E2E (when relevant files had been changed)

When relevant files (e.g. under tracer and sequencer related folders, `gradle/releases.versions.toml`, `gradle/libs.versions.toml`, etc.) have been changed, the CI pipeline builds the **linea-besu-package** Docker image using **locally built** tracer and sequencer plugins with the desired besu version in `libs.versions.toml` (will also build the besu from source if the desired besu version were not found in maven), and runs e2e tests against that image. This validates integration at PR time instead of only at release time.

---

## 3. Automatically release Linea-Besu-Package Image with Local-Built Tracer and Sequencer when PR get pushed to main (relevant files had been changed)

On push to `main`, the CI pipeline builds and publishes the **linea-besu-package** Docker image using **locally built** tracer and sequencer plugins with the desired besu version in `libs.versions.toml`, the release would be found in https://github.com/Consensys/linea-monorepo/releases

---

## 4. Linea-Besu-Package Makefile: Local Build with Local Tracer and Sequencer

The **`linea-besu-package/Makefile`** was revised so one can build the linea-besu-package image **locally** using tracer and sequencer plugins built from source in this repo, with versions read from `gradle/releases.versions.toml`.

### Example: full local build and e2e

From the repository root:

```bash
cd linea-besu-package
make build
make run-e2e-test
# when done
make run-e2e-test-cleanup
```

This builds tracer and sequencer from source, assembles the package with those zips, builds the Docker image, and runs e2e with the arithmetization version as the expected traces API version.

### Overrides

- **`TAG`** – Docker image tag (default: `local`).
- **`PLATFORM`** – Optional platform for `docker buildx` (e.g. `linux/amd64`).
- **`LOCAL_TRACER_ZIP_FOLDER`**, **`LOCAL_SEQUENCER_ZIP_FOLDER`**, **`LOCAL_BESU_DIST_FOLDER`** – Override if your zips are in a different location.
- **`EXPECTED_TRACES_API_VERSION`** - Override for a specific expected traces-api version used by Coordinator

---

## 5. Linea-besu-package release flow examples

### When the desired besu version needs to be updated

1. Create a PR to update the `besu` field in [libs.versions.toml](../gradle/libs.versions.toml) by updating the `besuCommit` field in [releases.versions.toml](../gradle/releases.versions.toml) which denotes the desired commit hash tag from the [hyperledger/besu](https://github.com/hyperledger/besu) repo
- The triggered Github workflows would automatically build the besu at the desired commit hash tag and publish it to cloudsmith [maven](https://cloudsmith.io/~consensys/repos/linea-besu/packages/) AND it would automatically push a commit to the PR to update the `besu` field in [libs.versions.toml](../gradle/libs.versions.toml)
- To get the corresponding `besu` version from the besu commit hash before running any Github workflow, one can run `./gradlew buildAndUpdateBesuVersionInLibsVersions` locally which shall see the update in [libs.versions.toml](../gradle/libs.versions.toml) OR run Makefile command i.e. `cd linea-besu-package && make build-besu-local-and-update-lib-version`
2. Make necessary changes in tracer and seqeuencer to accommodate the besu version change, and update the corresponding `arithmetization` version in [releases.versions.toml](../gradle/releases.versions.toml) (Please note that tracer and sequencer version will be set as `[arithmetization_version]-[7_char_commit_hash]` and the linea-besu-package release version will be prefixed with the `arithmetization` version)
3. Once the testing workflows of tracer and sequencer are all passed, the CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins and runs e2e tests against that image
4. When the PR merged into `main`, the release of linea-besu-package will be triggered automatically and will be found in the linea-monorepo release [page](https://github.com/Consensys/linea-monorepo/releases)

### When arithmetization version needs to be updated

1. Create a PR to make necessary changes in tracer (and seqeuencer if needed), and update the corresponding `arithmetization` version in [releases.versions.toml](../gradle/releases.versions.toml) (Please note that tracer and sequencer version will be set as `[arithmetization_version]-[7_char_commit_hash]`)
2. Once the testing workflows of tracer and sequencer are all passed, the CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins and runs e2e tests against that image
3. When the PR merged into `main`, the release of linea-besu-package will be triggered automatically and will be found in the linea-monorepo release [page](https://github.com/Consensys/linea-monorepo/releases)

### When hotfix of besu version is needed on a production image

1. Create a hotfix PR based on the **same commit tag** as the production image, and update the `besuCommit` field in [releases.versions.toml](../gradle/releases.versions.toml)
2. Make necessary changes in tracer and seqeuencer to accommodate the besu version change, and update the corresponding hotfix `arithmetization` versions in [releases.versions.toml](../gradle/releases.versions.toml)
3. The CI pipeline will build the `linea-besu-package` image using **locally built** tracer and sequencer plugins and runs e2e tests against that image
4. Trigger manual release [workflow](https://github.com/Consensys/linea-monorepo/actions/workflows/linea-besu-package-release.yml) on the hotfix branch and the release of linea-besu-package will be found in the linea-monorepo release [page](https://github.com/Consensys/linea-monorepo/releases)

---

## 6. Related documentations

- [Linea-besu-package README](./linea-besu-package/README.md)