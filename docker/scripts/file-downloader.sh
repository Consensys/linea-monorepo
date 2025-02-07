#!/usr/bin/env bash

# Function to display usage information
usage() {
    echo "Usage: $0 file_url destination_folder"
    echo "Example: $0 https://example.com/somefile.txt /path/to/destination/"
    exit 1
}

# Function to check and download a file
check_and_download_file() {
    local file_url="$1"
    local directory="$2"
    local filename="${file_url##*/}"  # Extracts the filename from the URL

    # Validate the URL format
    if [[ ! "$file_url" =~ ^https?:// ]]; then
        echo "Error: Invalid URL format. Please provide a valid HTTP/HTTPS URL."
        exit 1
    fi

    # Create the directory if it doesn't exist
    if [[ ! -d "$directory" ]]; then
        echo "Directory $directory does not exist. Creating it..."
        mkdir -p "$directory" || {
            echo "Error: Failed to create directory $directory."
            exit 1
        }
    fi

    # Check if the file already exists
    if [[ -f "$directory/$filename" ]]; then
        echo "File $filename already exists in $directory. Skipping download."
        return 0
    fi

    # Download the file
    echo "Downloading $file_url to $directory..."
    if wget -q --show-progress "$file_url" -P "$directory"; then
        echo "Download of $filename completed successfully!"
    else
        echo "Error: Failed to download $file_url."
        exit 1
    fi
}

# Main script execution
main() {
    # Validate the number of arguments
    if [ $# -lt 2 ]; then
        usage
    fi

    check_and_download_file "$1" "$2"
}

# Execute the main function
main "$@"
