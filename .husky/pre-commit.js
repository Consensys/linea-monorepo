const fs = require('fs');
const { execSync } = require('child_process');

// ENUMS

const FILE_EXTENSION = {
    TYPESCRIPT: "TYPESCRIPT",
    SOLIDITY: "SOLIDITY",
}

const FOLDER = {
    CONTRACTS: "CONTRACTS",
    SDK: "SDK",
}

const RUNTIME = {
    NODEJS: "NODEJS"
}

// MAPS

const FILE_EXTENSION_FILTERS = {
    [FILE_EXTENSION.TYPESCRIPT]: "\.ts$",
    [FILE_EXTENSION.SOLIDITY]: "\.sol$",
};

const FILE_EXTENSION_LINTING_COMMAND = {
    [FILE_EXTENSION.TYPESCRIPT]: "pnpm run lint:ts",
    [FILE_EXTENSION.SOLIDITY]: "pnpm run lint:sol",
};

const FOLDER_PATH = {
    [FOLDER.CONTRACTS]: "contracts/",
    [FOLDER.SDK]: "sdk/",
};

const FOLDER_CHANGED_FILES = {
    [FOLDER.CONTRACTS]: new Array(),
    [FOLDER.SDK]: new Array(),
};

const FOLDER_RUNTIME = {
    [FOLDER.CONTRACTS]: RUNTIME.NODEJS,
    [FOLDER.SDK]: RUNTIME.NODEJS,
};

// MAIN FUNCTION

main();

function main() {
    console.time('GET_CHANGED_FILE_LIST_TIMER');
    const changedFileList = getChangedFileList();
    partitionChangedFileList(changedFileList);
    console.timeEnd('GET_CHANGED_FILE_LIST_TIMER');

    for (const folder in FOLDER) {
        if (!isDependenciesInstalled(folder)) {
            console.error(`Dependencies not installed in ${FOLDER_PATH[folder]}, exiting...`)
            process.exit(1);
        }
        const changedFileExtensions = getChangedFileExtensions(folder);
        executeLinting(folder, changedFileExtensions);
    }

    updateGitIndex();
}

// HELPER FUNCTIONS

function getChangedFileList() {
    try {
        const cmd = 'git diff --name-only HEAD'
        const stdout = execSync(cmd, { encoding: 'utf8' });
        return stdout.split('\n').filter(file => file.trim() !== '');
    } catch (error) {
        console.error($`Error running ${cmd}:`, error.message);
        process.exit(1)
    }
}

function partitionChangedFileList(_changedFileList) {
    // Populate lists of filter matches
    for (const file of _changedFileList) {
        for (const path in FOLDER) {
            if (file.match(new RegExp(`^${FOLDER_PATH[path]}`))) {
                FOLDER_CHANGED_FILES[path].push(file);
            }
        }
    }
}

function isDependenciesInstalled(_folder) {
    const runtime = FOLDER_RUNTIME[_folder];
    const path = FOLDER_PATH[_folder];

    switch(runtime) {
        case RUNTIME.NODEJS:
            const dependencyFolder = `${path}node_modules`
            return fs.existsSync(dependencyFolder)
        default:
            console.error(`${runtime} runtime not supported.`);
            return false
    }
}

function getChangedFileExtensions(_folder) {
    // Use sets to implement early exit from iteration of all changed files, once we have matched all file extensions of interest.
    const remainingFileExtensionsSet = new Set(Object.values(FILE_EXTENSION));
    const foundFileExtensionsSet = new Set();

    for (const file of FOLDER_CHANGED_FILES[_folder]) {
        for (const fileExtension of remainingFileExtensionsSet) {
            if (file.match(new RegExp(FILE_EXTENSION_FILTERS[fileExtension]))) {
                foundFileExtensionsSet.add(fileExtension);
                remainingFileExtensionsSet.delete(fileExtension);
            }
        }

        // No more remaining file extensions to look for
        if (remainingFileExtensionsSet.size == 0) break; 
    }

    return Array.from(foundFileExtensionsSet);
}

function executeLinting(_folder, _changedFileExtensions) {
    for (const fileExtension of _changedFileExtensions) {
        const path = FOLDER_PATH[_folder];
        const cmd = FILE_EXTENSION_LINTING_COMMAND[fileExtension];
        console.log(`${fileExtension} change found in ${path}, linting...`);
        try {
            // Execute command synchronously and route output directly to the current stdout
            execSync(`
                cd ${path};
                ${cmd};
            `, { stdio: 'inherit' });
        } catch (error) {
            console.error(`Error:`, error.message);
            console.error(`Exiting...`);
            process.exit(1);
        }
    }
}

function updateGitIndex() {
    try {
        const cmd = 'git update-index --again'
        execSync(cmd, { stdio: 'inherit' });
    } catch (error) {
        console.error($`Error running ${cmd}:`, error.message);
        process.exit(1);
    }
}