# txt4kindlegen
`txt for kindlegen`:
A Simple tool to convent text file to a mobi file by `kindlegen`

## compile
Install go-bindata

```
go get -u github.com/jteeuwen/go-bindata/...
```

Run the build cmd:
```
go-bindata -o assets/asset.go -pkg=assets assets/...
go build
```

## Usage
```
usage: txt4kindlegen [<flags>]

Flags:
      --help                  Show context-sensitive help (also try --help-long and --help-man).
  -c, --config="config.toml"  config file
      --init                  make a config file
```
* `--init` create a config_example.toml on current dir
* `-c,--config` specify the config file