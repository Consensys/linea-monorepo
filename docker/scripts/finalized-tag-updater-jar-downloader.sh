#!/usr/bin/env bash

if [ $# -lt 2 ]; then
    echo "Usage: $0 github_access_token plugin_version destination_folder"
    echo "Example: $0 ghp_xxx 0.0.1 /path/to/destination/"
    exit
fi

check_and_download_file() {
    local github_access_token="$1"
    local plugin_version="$2"
    local destination_folder="$3"
    local asset_name="finalized-tag-updater-v$plugin_version"
    local file_name="$asset_name".jar

    # Check if the file exists in the directory
    if [[ ! -f "$destination_folder/$file_name" ]]; then
        # File does not exist, download it
        download_url=$(\
            curl -sL -H 'Accept:application/json' \
                -u $1: https://api.github.com/repos/Consensys/zkevm-monorepo/releases | jq -rc \
                'map(select(.tag_name | contains('"\"$asset_name\""'))) | .[] .assets[] | select(.name | contains('"\"$file_name\""')) .url'\
            )
        echo "Downloading $file_name from url=$download_url ..."
        curl -L -H 'Accept:application/octet-stream'  -u "$github_access_token": -o "$destination_folder/$file_name" "$download_url"
        echo "Download complete!"
    else
        echo "File $file_name already exists in $destination_folder"
    fi
}
echo "$0 github_access_token $2 $3"
check_and_download_file "$1" "$2" "$3"
