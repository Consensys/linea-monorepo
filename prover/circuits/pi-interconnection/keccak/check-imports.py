# DELETE THIS

import os
import re

def find_unique_imports(directory):
    # Regex pattern to match the required substring
    pattern = r"\"github\.com/consensys/linea-monorepo/\S+"
    # Set to store unique substrings
    unique_substrings = set()

    # Walk through the directory recursively
    for root, dirs, files in os.walk(directory):
        for file in files:
            try:
                # Get the full file path
                file_path = os.path.join(root, file)

                # Open and read the file content
                with open(file_path, "r", encoding="utf-8", errors="ignore") as f:
                    content = f.read()
                    # Find all substrings matching the pattern
                    substrings = re.findall(pattern, content)
                    # Add them to the set (to maintain uniqueness)
                    unique_substrings.update(substrings)
            except Exception as e:
                print(f"Error reading file {file_path}: {e}")

    # Sort the unique substrings alphabetically
    sorted_substrings = sorted(unique_substrings)

    # Print out all unique substrings found
    for substring in sorted_substrings:
        print(substring)

if __name__ == "__main__":
    # Set the directory to search (modify this as needed)
    root_directory = "./"  # Change this to the target directory
    find_unique_imports(root_directory)
