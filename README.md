# netconf-go
Netconf CLI with tab completion and dynamic download of yang modules.

Very much alpha quality - probably requires changes before picking up and using.

Can also be compiled to a WebAssembly library for use by a web frontend. 
I have a separate Flutter web app that does this, which I hope to also make a public project shortly - contact me if interested.

## Usage

```
Usage of netconf-go:
      --address string    Address or host to connect to (default "localhost")
      --debug string      debug level (default "info")
      --host string       Hostname referring to hosts.yaml entry
      --password string   Password
      --port int          Port number to connect to (default 22)
      --t                 Use telnet to connect
      --user string       Username
```

## Augment handling

TODO

## Hosts file

Host details can be saved into a Yaml file `hosts.yaml` in the current directory.

Each host should have akey that is then passed to the 'hosts' command-line parameter to select the entry.

The entries under the key should match the equivalent command-line parameters.

The user can mix both the hosts file and command-line parameters, in which case 
if both are specified the command-line parameter will override the hosts entry.

## Contributing
Pull requests are welcome. For major changes, please open an issue first to discuss what you would like to change.

Currently virtually no automated unit-tests, but adding tests would be welcome.

# To-do list

- rearrange to use proper Go module structure and packages
- upstream goyang changes
- implement config files for hosts
- improve augment handling
- add tests