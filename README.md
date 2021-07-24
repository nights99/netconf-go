# netconf-go
Netconf CLI with tab completion and dynamic download of yang modules.

Very much alpha quality - probably requires changes before picking up and using.

Can also be compiled to a WebAssembly library for use by a web frontend. 
I have a separate Flutter web app that does this, which I hope to also make a public project shortly - contact me if interested.

## Usage

```
Usage of /home/jon/go/bin/netconf-go:
  -address string
        Address or host to connect to (default "localhost")
  -debug string
        debug level (default "info")
  -port int
        Port number to connect to (default 10555)
```

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Currently virtually no automated unit-tests, but adding tests would be welcome.
