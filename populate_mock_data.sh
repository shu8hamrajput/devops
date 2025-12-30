#!/bin/bash
# Script to populate database with mock data
# This script can be run inside the Docker container or locally

echo "Populating Splitwise database with mock data..."

# Check if running in Docker or locally
if [ -f /.dockerenv ] || [ -n "$DOCKER_CONTAINER" ]; then
    # Running in Docker
    python3 /app/populate_mock_data.py
else
    # Running locally - check if psycopg2 is installed
    if ! python3 -c "import psycopg2" 2>/dev/null; then
        echo "Installing psycopg2..."
        pip3 install psycopg2-binary
    fi
    
    python3 populate_mock_data.py
fi

