#!/usr/bin/env bash

if [ $# -lt 2 ]; then
    echo "Usage: $0 file_url destination_folder"
    echo "Example: $0 https://example.com/somefile.txt /path/to/destination/"
    exit
fi

check_and_download_file() {
    local file_url="$1"
    local directory="$2"
    local filename="${file_url##*/}"  # Extracts the filename from the URL

    # Check if the file exists in the directory
    if [[ ! -f "$directory/$filename" ]]; then
        # File does not exist, download it
        echo "Downloading $file_url ..."
        wget "$file_url" -P "$directory"
        echo "Download complete!"
    else
        echo "File $filename already exists in $directory."
    fi
}
echo "$0 $1 $2"
check_and_download_file "$1" "$2"
