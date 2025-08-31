#!/usr/bin/env bash

# Check if the required arguments are provided
if [ "$#" -lt 2 ]; then
  echo "Usage: $0 <max_snapshots> <arch>"
  echo "Example: $0 5 x86/arm"
  exit 1
fi

# Positional arguments
MAX_SNAPSHOTS=$1
ARCH=$2

# Get the list of snapshots
snapshots=$(hcloud image list --selector os=talos,testing=true,arch=${ARCH} --output json)

snapshot_ids=$(echo "$snapshots" | jq -r 'sort_by(.created) | reverse | .[].id')
count=0
echo "found images: ${snapshots}"
for snapshot_id in $snapshot_ids; do
  count=$((count + 1))
  if [ $count -gt ${MAX_SNAPSHOTS} ]; then
    echo "Deleting snapshot $snapshot_id (exceeds retention limit)"
    hcloud image delete $snapshot_id
  fi
done
