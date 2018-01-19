# HSSH - HTTP to SSH

## install
```commandline
go get github.com/vedhavyas/hssh/cmd/hssh/...
```

## Starting server
If your gobin is set under Path, you call hssh
```commandline
hssh
```

for more details
```commandline
hssh --help
```

## Usage
Make a post to `\ssh` with server_ip an command_string
```commandline
curl -XPOST "http://localhost:8080/ssh" -d '{"server_ip": "54.173.10.86", "command_string": "cat /usr/share/dict/words"}'
```

## result
Once the post call is made, hssh will pick the ssh config from `~/.ssh/config` and fallback to `/etc/ssh/ssh_config`
Hssh streams the data if the output is larger than buffer. To test this, use curl. Postman doesn't support this at the moment.
