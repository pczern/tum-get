#!/bin/bash

if ! which pdfunite > /dev/null; then
    echo "pdfunite from poppler-utils is not installed. Please install it and try again."
    exit 1
fi
./tum-get

start_dir="$1"

mkdir -p "$start_dir"

if [ ! -d "$start_dir" ]; then
    echo "Error: $start_dir is not a valid directory."
    exit 1
fi

find "$start_dir"/*/* -type d | while read -r dir
do
  current_dir=$(basename "$dir")
  parent_dir=$(basename "$(dirname "$dir")")
  current_date=$(date +%y%m%d)

  find "$dir" -type f | while read -r file
  do
    new_file=$(echo "$file" | sed 's/ä/ae/g' | sed 's/ö/oe/g' | sed 's/ü/ue/g' | sed 's/Ä/Ae/g' | sed 's/Ö/Oe/g' | sed 's/Ü/Ue/g' | sed 's/ß/ss/g')
    if [ "$file" != "$new_file" ]; then
        mv "$file" "$new_file"
    fi
  done

  if ls "$dir"/*.pdf 1> /dev/null 2>&1; then
    merge_file=("$start_dir"/"$parent_dir"-"$current_dir"-*.pdf)
    if [ -e "${merge_file[0]}" ]; then
      rm "$start_dir"/"$parent_dir"-"$current_dir"*.pdf
    fi
    pdfunite "$dir"/*.pdf "$start_dir"/"$parent_dir"-"$current_dir"-"$current_date".pdf
  fi
done