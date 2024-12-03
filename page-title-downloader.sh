#!/bin/bash

# Main script to download Wikipedia titles for a given language
# Usage: sh main.sh --lang=<language_code>

# Function to display usage instructions
usage() {
    echo "Usage: $0 --lang=<language_code>"
    exit 1
}

# Parse the input arguments
for arg in "$@"; do
    case $arg in
        --lang=*)
            LANGUAGE="${arg#*=}"
            shift
            ;;
        *)
            usage
            ;;
    esac
done

# Check if the language parameter is provided
if [ -z "$LANGUAGE" ]; then
    echo "Error: Language code not provided."
    usage
fi

# Create a directory to store the title database
mkdir -p title-db

# Navigate to the title database directory
cd title-db || exit

# Download the titles dump for the specified language
DUMP_URL="https://dumps.wikimedia.org/${LANGUAGE}wiki/latest/${LANGUAGE}wiki-latest-all-titles.gz"
wget "$DUMP_URL" || { echo "Error: Failed to download dump for language '$LANGUAGE'"; exit 1; }

# Extract the downloaded file
gunzip "${LANGUAGE}wiki-latest-all-titles.gz" || { echo "Error: Failed to extract the titles dump."; exit 1; }

# Process the extracted file: remove the header and extract the second column
tail -n +2 "${LANGUAGE}wiki-latest-all-titles" | cut -f 2 > "${LANGUAGE}-titles.txt" || {
    echo "Error: Failed to process the titles file."
    exit 1
}

# Return to the original directory
cd .. || exit

echo "Wikipedia titles for language '${LANGUAGE}' have been downloaded and processed successfully."
