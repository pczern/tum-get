# tum-get

Download all Zips for your specified courses on Moodle, then move them based on a pattern or just leave them in the `root` folder.

input.json
```json
{ 
  "username": "username",
  "password": "password",
  "root": "/Users/user/uni/moodle",
  "data": [
    {
      "course": "Course 1",
      "url": "https://www.moodle.tum.de/local/downloadcenter/index.php?courseid=xxxxx",
      "rules": [{
        "method": "copy",
        "file_pattern": "solutions*",
        "dest": "/Users/user/uni/sorted/course1/solutions"
      }]
    },
    {
      "course": "Course 2",
      "url": "https://www.moodle.tum.de/local/downloadcenter/index.php?courseid=xxxxx",
      "rules": [{
        "method": "rename",
        "file_pattern": "solutions*",
        "path_pattern": "solutions1*",
        "neg_file_pattern": "solutions2*",
        "neg_path_pattern": "solutions2*",
        "regex_file_pattern": "^.*solutions.*$",
        "regex_path_pattern": "^.*solutions.*$",
        "neg_regex_file_pattern": "^.*solutions2.*$",
        "neg_regex_path_pattern": "^.*solutions2.*$",
        "dest": "/Users/user/uni/sorted/course2/solutions"
      }]
    }
  ]
}
```

Note that `path_pattern`s match the full path, and `file_pattern`s match only the basenames of the files.
All patterns are matched for files in a rule before the file is written to `dest`.<br>
`method` can be `copy` or `rename`. `rename` moves the matched files to `dest` and creates a placeholder file where the original was. `copy` copies the files to `dest`, a placeholder is not created.
New files that are not in the file system yet are outputted to the console, this is why `rename` creates a placeholder so that the program knows that the file already existed.<br>


## Run
Install Golang, build the binary for your system & run it within the same folder as `input.json`.
```bash
go build .
./tum-get
```
You can also skip fetching the zip files from Moodle and just use the file pattern matching.
```bash
./tum-get -skip-fetch
```

## Tips
The tool `poppler-utils` has a utility called `pdfunite`. You can install `poppler-utils` to merge multiple PDFs from the command line.
The bash script `script.sh` merges multiple PDFs in subdirectories. It's recommended to use it with an `input.json` config that has a two level structure for `dest` like `/sorted/course/solutions`. You can run it in the console with `./script.sh`.