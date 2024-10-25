# Smart Contract Testing Guidelines

## Folder Structure and Naming Conventions

### Folders
- Maintain a 1:1 structure with the contract directory for ease of reference.
- Shared utilities and helpers should be placed in `folderRoot/utilities/shared`.

### Test Files
- Follow a 1:1 structure with the contract directory for ease of reference.
- If a contract is complex and requires multiple files, create a subfolder with individual test files.

    Example:
    ```plaintext
    ./LineaRollup/Finalization.ts
    ./LineaRollup/BlobSubmission.ts
    ./LineaRollup/CalldataDataSubmission.ts
    ./LineaRollup/PermissionsOrOtherThings.ts
    ```

- Contract-specific or feature-specific helper or utility functions should be in a separate file within a child folder named `helpers/contractName.ts`.

    Example:
    ```plaintext
    ./LineaRollup/Finalization.ts
    ./LineaRollup/helpers/finalization.ts
    ./LineaRollup/helpers/blobSubmission.ts
    ```

### Utilities
- Common functions and helper files should be defined by functionality that is not specific to any given contract.
    - Example: test assertions, hashing, and general framework behavior.
- Each helper file should contain functionality with a single reason to change (e.g. hashing, timing, encoding).
- File naming convention should describe the whole functionality using adverbial language (e.g. hashing).

## Test Layout

- Organize tests by contract name, external calls, behavior, starting with failure tests followed by success tests.

### Describe
- Contract name (e.g., `Rate Limiter`).
    - External call (e.g., `When initializing`).
        - Expected behavior (if there are a few subcategories of behavior, this could be broken down further as well).
            - Failure tests first.
                - Tests should follow the order of code paths (failure ordering for ease of code reading).
            - Success tests last.

## Additional Guidelines

### Ethers Wrapping/Abstractions
- Ensure proper wrapping and abstraction using Ethers.js.

### Solidity Test Contracts
- Use Solidity test contracts where necessary.

### Event Testing
- Test events, including arguments.

### Error Testing
- Test errors, including arguments.

### Scenario Building Pre-Testing
- Build scenarios before testing to ensure comprehensive coverage.

### Coverage
- Ensure high test coverage for all contracts.

### Storage Layout and Upgrade Automated Testing
- Implement automated testing for storage layout and upgrades.
