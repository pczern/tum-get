# tum-get

Download all zips for your specified courses on Moodle, then move them based on a pattern or just leave them in the `root` folder.

input.json
```json
{ 
  "username": "username",
  "password": "password",
  "root": "/Users/user/uni/moodle",
  "data": [
    {
      "course": "Course 1",
      "url": "https://www.moodle.tum.de/local/downloadcenter/index.php?courseid=XXX",
      "rules": [{
        "method": "copy",
        "pattern": "solutions*",
        "dest": "/Users/user/uni/course/solutions"
      }]
    }
  ]
}
```

## Run
Install Golang, build the binary for your system & run it in the same folder with `input.json`
```bash
go build
./tum-get
```