#!/bin/bash

./tum-get

start_dir="/Users/user/uni/sorted"

find "$start_dir" -type d | while read -r dir
do
  current_dir=$(basename "$dir")
  parent_dir=$(basename "$(dirname "$dir")")
  current_date=$(date +%Y%m%d) 
  if ls "$dir"/*.pdf 1> /dev/null 2>&1; then
      if [ -f "$dir"/"$parent_dir"-"$current_dir"-"$current_date".pdf ]; then
        rm "$dir"/"$parent_dir"-"$current_dir"*.pdf
      fi
      pdfunite "$dir"/*.pdf "$dir"/"$parent_dir"-"$current_dir"-"$current_date".pdf
  fi
done