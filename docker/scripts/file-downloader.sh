#!/usr/bin/env bash

if [ $# -lt 2 ]; then
    echo "Usage: $0 file_url destination_folder"
    echo "Example: $0 https://example.com/somefile.txt /path/to/destination/"
    exit 1
fi

check_and_download_file() {
    local file_url="$1"
    local directory="$2"
    local filename="${file_url##*/}"  # Extracts the filename from the URL

    # Create the directory if it doesn't exist
    if [[ ! -d "$directory" ]]; then
        echo "Directory $directory does not exist. Creating it..."
        mkdir -p "$directory" || {
            echo "Failed to create directory $directory."
            exit 1
        }
    fi

    # Check if the file already exists
    if [[ ! -f "$directory/$filename" ]]; then
        echo "Downloading $file_url to $directory..."
        if wget -q "$file_url" -P "$directory"; then
            echo "Download of $filename completed successfully!"
        else
            echo "Failed to download $file_url."
            exit 1
        fi
    else
        echo "File $filename already exists in $directory. Skipping download"
    fi
}

# Execute the function
check_and_download_file "$1" "$2"
