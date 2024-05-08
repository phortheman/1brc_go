# One Billion Row Challenge
## Description
Te idea of the one billion row challenge (1brc) is to implement the fastest possible way to parse 1 billion rows of data and aggregate it into a result. It was originally a Java challenge but other have taken the challenge in other languages.
[Learn more](https://1brc.dev/)

To generate my measurements and results I used [this C implementation](https://github.com/dannyvankooten/1brc) mostly because it is significantly faster than the baseline Java implementation and I found it valuable to have several smaller samples for quick tests before running a profile on the 1 billion rows.

## Profiling
This challenge was amazing to learn about Go's builtin tooling for profiling. I added the recommended flag package and the recommended lines to my code to implement. Once the profile file is created simply run
`go tool pprof -http=:8080 <PROFILE FILE>`
This will launch an http server and open your browser to inspect instead of using the CLI interface. This helped me find out where my implementation was slow so I could focus on improving it.
NOTE: I had to install xdg-utils and graphviz on WSL to get it to work

## Stages
- First tried to use a go routine to read the file while the main thread handled everything else. I learn about the overhead that go routines have and how it is unnecessary to do with the standard library bufio.Reader. (117.53s)
- Later implementations where around using implied decimal format as it is easier to work with integers instead of floats and then I just converted the int into a float at the end.
 - Implemented my own solution to convert a string into an int that is faster than the built in conversion (less safe but the data structure is known) (62.74s)
- Improved map access by switching the value to be a pointer to that stations value. The builtin map access functionality if actually very slow so doing it once per line was an improvement. (51.77s)
- Implementing my own line parser instead of using the built in LastIndex found some savings. (45.67s)
- Found out there is an optimization implemented for faster byte slice to string conversion which significantly improved performance (39.06s)

## To do
- I'm confident that my single threaded performance is at the max. The only other improvement would be to implement my own hash map in Go to improve map access. May revisit this but I feel like that isn't the Go way of doing things.
- Next steps would be to use multi-threading to spread the workload. Going to require some research to ensure efficiency/safety
