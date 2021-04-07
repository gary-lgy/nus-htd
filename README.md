# NUS-HTD
A command-line tool to make and view your temperature declarations at National University of Singapore (NUS).

# How to build
Clone the reposiotry and run:

```bash
go build
```

then copy the generated executable to somewhere in your $PATH.

# How to use the tool
NUSNET ID and password can be supplied using either command-line arguments or from environment variables (`$HTD_USERNAME` and `$HTD_PASSWORD`).

```
usage: nus-htd [<flags>] <command> [<args> ...]

A command-line tool for making and viewing your daily temperature declarations at NUS.

Flags:
      --help               Show context-sensitive help (also try --help-long and --help-man).
  -u, --username=USERNAME  Your NUSNET ID. (default: $HTD_USERNAME.)
  -p, --password=PASSWORD  Your NUSNET password. (default: $HTD_PASSWORD)
      --debug              Print the received command line arguments and flag and immediately exit.

Commands:
  help [<command>...]
    Show help.

  declare [<flags>] <am or pm> <temperature>
    Declare your temperature. (d for short)

  view
    View your past declarations. (v for short)
```

For example,
To pass username and password on the command-line:
```bash
nus-htd d[eclare] -u USERNAME -p PASSWORD am 36.5
nus-htd v[iew] -u USERNAME -p PASSWORD
```
Alternatively, if you have set the environment variables, simply run:
```bash
nus-htd d[eclare] am 36.5
nus-htd v[iew]
```

# Why did I write this
1. To explore Golang's HTTP client
2. To write a small CLI tool in Golang
3. To save a few clicks when declaring my temperatures

# Credits
- [nus-htd-automation](https://github.com/jiachen247/nus-htd-automation) for reverse-engineering the authentication flow

# License
MIT
