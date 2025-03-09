# tum-get

Download all Zips for your specified courses on Moodle, then move them based on a pattern or just leave them in the `root` folder.

## Features
- [x] Concurrent Downloads 🔀
- [x] Extract Zips 📁
- [x] Pattern Matching 🔍
- [x] Filename Matching
- [x] Negative Patterns
- [x] Display New Files (not comparing dates, just filenames)
- [x] Open Source

## Run
Download the binary & run it within the same folder as `input.json`.
```bash
./tum-get # when you download the binary rename it or add version, os and architecture to the name
```
You can also skip fetching the zip files from Moodle and just use the file pattern matching capabilities.
```bash
./tum-get -skip-fetch
```

## Examples
Save the json in a file called `input.json`.

### Simple
Downloads the course's zip to the `moodle` folder.
```json
{ 
  "username": "username",
  "password": "password",
  "root": "/Users/user/uni/moodle",
  "data": [
    {
      "course": "Course 1",
      "url": "https://www.moodle.tum.de/local/downloadcenter/index.php?courseid=xxxxx",
     
    },
  ]
}
```
### Intermediate
Downloads the course's zip and use [Golang `filepath.Match`](https://pkg.go.dev/path/filepath#Match) to match all files in a course that contain `solutions*` in the filename, and then copies them to `sorted/course1/solutions`.
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
  ]
}
```

### Advanced
Downloads the course's zip and matches by regex all paths that have a *"solutions"* string but not a *"solutions2"* string, and copies them.
The negative patterns are added because negative lookaheads are not supported out of the box in Go's standard regexp package.
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
        "regex_path_pattern": "^.*solutions.*$",
        "neg_regex_path_pattern": "^.*solutions2.*$",
        "dest": "/Users/user/uni/sorted/course1/solutions"
      }]
    }
  ]
}
```

### All possible patterns
Matches paths and filenames against many patterns and then moves the file instead of copying it when all matches succeed.
An empty placeholder file is created for each moved file.
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

Note that `path_pattern`s match the full path, and `file_pattern`s match only the basenames of the files.<br>
All patterns are matched for files in a rule before the file is written to `dest`.<br><br>
`method` can be `copy` or `rename`.<br> `rename` moves the matched files to `dest` and creates a placeholder file where the original was.<br>`copy` copies the files to `dest`, a placeholder is not created.<br>
New files that are not in the file system yet are outputted to the console, this is why `rename` creates a placeholder so that the program knows that the file already existed.<br
The folders in `input.json` are auto-generated.

## Tips
The tool `poppler-utils` has a utility called `pdfunite`. You can install `poppler-utils` to merge multiple PDFs from the command line.
The bash script `script.sh` merges multiple PDFs in subdirectories.  It's recommended to use it with an `input.json` config that has a two level structure for `dest` like `/sorted/course/solutions`, `/sorted/course/exercises`.
```bash
brew install poppler # install poppler-utils
./script.sh /Users/user/uni/sorted # merge PDFs in subdirectories of folder "sorted"
cd /Users/user/uni/sorted
ls # view merged PDFs
```

## Bulid and Run from Source Code
To build from source install [Golang](https://go.dev/doc/install), build the binary for your system & run it within the same folder as `input.json`.
```bash
go build .
./tum-get
```
