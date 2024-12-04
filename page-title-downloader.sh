#!/bin/sh

# Main script to download Wikipedia titles for a given language and split them
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

# Count total lines
TOTAL_LINES=$(wc -l < "${LANGUAGE}-titles.txt")
echo "Total lines: $TOTAL_LINES"

# Calculate lines per split (rounded down)
LINES_PER_SPLIT=$((TOTAL_LINES / 7))
REMAINDER=$((TOTAL_LINES % 7))

# Create split directory
mkdir -p split-titles

# Split the file
i=1
while [ $i -le 7 ]; do
    START_LINE=$(( (i - 1) * LINES_PER_SPLIT + 1 ))
    END_LINE=$(( i * LINES_PER_SPLIT ))
    
    # Adjust the last split to include any remainder lines
    if [ $i -eq 7 ]; then
        END_LINE=$((TOTAL_LINES))
    fi
    
    # Extract the specified line range
    sed -n "${START_LINE},${END_LINE}p" "${LANGUAGE}-titles.txt" > "split-titles/titles-part-${i}.txt"
    
    # Verify the number of lines in the split file
    SPLIT_LINES=$(wc -l < "split-titles/titles-part-${i}.txt")
    echo "Part $i: Lines from $START_LINE to $END_LINE (Total: $SPLIT_LINES lines)"
    
    # Increment counter
    i=$((i + 1))
done

# Return to the original directory
cd .. || exit

echo "Wikipedia titles for language '${LANGUAGE}' have been downloaded, processed, and split successfully."