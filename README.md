# sqlite-vss Go Example

This is meant to demonstrate a fully-functional example of using the sqlite-vss extension, with cybertron.

## Usage

```bash
make demo # builds and runs ./bin/demo, use this to get started
DBNAME=":memory:" make demo # runs the demo in memory only

# additional targets
make # builds the binary to ./bin/demo
make extensions # downloads the os specific static build of the sqlite-vss extension. This is a dep for build.
```

On first `make demo` to build and run, the app will go through a few steps including:

- Download the OS-specific `static-*.tar.gz` release from sqlite-vss
- Pull sentence-transformers/all-MiniLM-L6-v2 into an OS-specific cache directory
- Seed the local db with a 1500 sample set of news articles

After all the setup is complete, the app will provide a prompt to enter headline phrases for searching. Each search will
show the 5 most relevant answers and their distances.

![example.gif](example.gif)
