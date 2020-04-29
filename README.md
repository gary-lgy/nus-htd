# What
A command-line tool to make and view your temperature declarations at NUS.

# How
You need to supply your NUSNET ID and password, either using the command-line flags, or from the `$HTD_USERNAME` and `$HTD_PASSWORD` environment variables.

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
If you set the required environment variables:
```bash
nus-htd d[eclare] am 36.5
nus-htd v[iew]
```

# Why
1. To explore Golang's HTTP client
2. To write a small CLI tool in Golang
3. To save a few seconds of clicking when declaring my temperatures

# Credits
- [nus-htd-automation](https://github.com/jiachen247/nus-htd-automation) for reverse-engineering the authentication flow

# License
MIT
