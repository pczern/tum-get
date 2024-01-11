#!/bin/bash

./tum-get

start_dir="/Users/user/uni/moodle"

find "$start_dir" -type d | while read -r dir
do

  current_dir=$(basename "$dir")
  parent_dir=$(basename "$(dirname "$dir")")
  if ls "$dir"/*.pdf 1> /dev/null 2>&1; then
      rm "$dir"/"$parent_dir"-"$current_dir".pdf
      pdfunite "$dir"/*.pdf "$dir"/"$parent_dir"-"$current_dir".pdf
  fi
done