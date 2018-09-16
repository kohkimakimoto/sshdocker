# sshdocker

Run docker containers over SSH.

sshdocker is a standalone SSH server written in Go. It launches a docker container when a SSH client connects it. You can run abitrary commands in the isolated container on a remote machine.

Table of Contents

* [Installation](#installation)
* [Usage](#usage)
  * [Example](#example)
  * [Pipe](#pipe)
* [Configuration](#configuration)
* [Author](#author)
* [License](#license)

## Installation

As sshdocker is a single binary command, to install it you can download the binrary from github release page and drop it in you $PATH

[Download latest version](https://github.com/kohkimakimoto/sshdocker/releases/latest)

If you use CentOS7, you can also use RPM package that is stored in the same release page. It is useful because it configures systemd service automatically.

To validate your installation, run `sshdocker` with `-h` option.

```
$ sshdocker -h
Usage: sshdocker [OPTIONS...]

sshdocker : Run docker containers over SSH.
version 0.1.0 (8bb0abc09c3c0a04370759f9efb4395291f824a0)

Options:
  -c, -config-file=FILE    Load configuration from the file.
  -h, -help                Show help.
  -v, -version             Print the version.

See: https://github.com/kohkimakimoto/sshdocker for updates, code and issues.
```

## Usage

### Example

Create configuration file `sshdocker.yml` as the following.

```yaml
addr: 0.0.0.0:2222

runtimes:
  example:
    image: centos:centos7
    command: [ "/bin/bash" ]

  php:
    image: php:7.2
    command: [ "/bin/bash" ]

```

Run sshdocker with this config.

```
$ sshdocker -c sshdocker.yml
2018/09/12 19:14:45 starting ssh server on 0.0.0.0:2222
```

Your sshdocker boots as a SSH server. Open new terminal and connect it by using ssh client.

```
$ ssh -p 2222 example@localhost
[root@8fc664e323ed /]# cat /etc/redhat-release 
CentOS Linux release 7.4.1708 (Core) 
```

You are in the `centos:centos7` container. The ssh login user `example` is used to specify **runtime** name that is defined in the above config. The **runtime** is settings how to run a docker container. For more detail see [Configuration](#configuration). At now, type `exit` command to terminate the container and disconnect from the ssh server. 

Connect another runtime `php` as the following.

```
$ ssh -p 2222 php@localhost
root@203c522b3376:/# php -v
PHP 7.2.9 (cli) (built: Sep  7 2018 20:26:47) ( NTS )
Copyright (c) 1997-2018 The PHP Group
Zend Engine v3.2.0, Copyright (c) 1998-2018 Zend Technologies
root@203c522b3376:/#
```

Now you are in `php:7.2` container. You can use PHP in this container.

### Pipe

You can pipe arbitrary data into sshdocker's SSH server. Run the following command in this repository's `example` directory to build the example docker image.

```
docker build --rm -t jq .
```

And then, add the following runtime config to your `sshdocker.yml`

```yaml
runtimes:
  # other runtimes...

  # add this definition!
  jq:
    image: jq
```

You don't need to reload the sshddocker process. Let's pipe some json into your sshdocker's SSH service.

```
curl -s https://api.github.com/users/kohkimakimoto | ssh -p 2222 jq@localhost .
```

You will get the following result.

```json
{
  "login": "kohkimakimoto",
  "id": 761462,
  "node_id": "MDQ6VXNlcjc2MTQ2Mg==",
  "avatar_url": "https://avatars0.githubusercontent.com/u/761462?v=4",
  "gravatar_id": "",
  "url": "https://api.github.com/users/kohkimakimoto",
  "html_url": "https://github.com/kohkimakimoto",
  "followers_url": "https://api.github.com/users/kohkimakimoto/followers",
  "following_url": "https://api.github.com/users/kohkimakimoto/following{/other_user}",
  "gists_url": "https://api.github.com/users/kohkimakimoto/gists{/gist_id}",
  "starred_url": "https://api.github.com/users/kohkimakimoto/starred{/owner}{/repo}",
  "subscriptions_url": "https://api.github.com/users/kohkimakimoto/subscriptions",
  "organizations_url": "https://api.github.com/users/kohkimakimoto/orgs",
  "repos_url": "https://api.github.com/users/kohkimakimoto/repos",
  "events_url": "https://api.github.com/users/kohkimakimoto/events{/privacy}",
  "received_events_url": "https://api.github.com/users/kohkimakimoto/received_events",
  "type": "User",
  "site_admin": false,
  "name": "Kohki Makimoto",
  "company": "Freelance",
  "blog": "http://kohkimakimoto.hatenablog.com/",
  "location": "Tokyo, Japan",
  "email": null,
  "hireable": null,
  "bio": null,
  "public_repos": 132,
  "public_gists": 10,
  "followers": 35,
  "following": 64,
  "created_at": "2011-05-01T04:25:17Z",
  "updated_at": "2018-09-04T12:10:25Z"
}
```

## Configuration

These are the available configuration options of sshdocker. You can specify the config file by `-c` or `-config-file` option.

```yaml
# `addr` is the ssh server listen address.
addr: 0.0.0.0:2222

# `host_key_file` is the ssh server's host key 
host_key_file: /etc/ssh/ssh_host_rsa_key

# `debug` is a debug flag. If you set it true, sshdocker outputs verbose debug messages to the stderr.
debug: false

# `container_label` is a label string for `docker run --label` options.
# sshdocker runs docker container with this label. This label is useful to search containers
# such as `docker ps --filter=label=sshdocker`.
container_label: sshdocker

# `public_key_authentication` is a flag to enable public key authentication.
# This config is very important. You should always set it true to protect from evil access.
public_key_authentication: true

# `authorized_keys_file` is the path to a authorized_keys file used by public key authentication.
authorized_keys_file: /root/.ssh/authorized_keys

# `authorized_keys` is array of the public keys. 
# you can write directly the authorized_keys in this configuration file.
authorized_keys:
  - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB..."
  - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB..."

# `runtimes` is definitions how to run docker containers.
# The property names are used as ssh login names. If you specify `example@localhost` by the ssh client,
# sshdocker uses `example` runtime settings for running a docker container.
runtimes:
  # This is an example runtime.
  example:
    # `public_key_authentication` overrides global `public_key_authentication` config.
    public_key_authentication: true
  
    # `authorized_keys_file` overrides global `authorized_keys_file` config.
    authorized_keys_file: tests/key/id_rsa.pub
    
    # `authorized_keys` overrides global `authorized_keys_file` config.
    authorized_keys:
      - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB..."
      - "ssh-rsa AAAAB3NzaC1yc2EAAAADAQAB..."

    # `image` is a docker image to create a container.
    image: centos:centos7

    # `container` is settings for creating docker container. 
    # This config can NOT be used with the above `image` config.
    # Instead of simply running a container, if you want to create a container and execute
    # a command in the container, you can this config.
    # For instance, if you want to run `/sbin/init` in the container and use bash shell process,
    # the following example is good for you.
    container:
      image: centos:centos7
      
      options: [ "--privileged" ]
      
      command: [ "/sbin/init" ]

    # `options` is `docker run` or `docker exec` command options.
    options: [ "-e", "HOGE=hogehoge" ]

    # `command` is a default command to be executed in a docker container.
    # If you pass the command by the ssh client such as `ssh -p 2222 example@localhost ls -la`,
    # sshdocker uses the `ls -la` as the command.
    command: [ "/bin/bash" ]

  # `_fallback` is a special runtime.
  # If you define the `_fallback` and ssh login name doesn't match with any other runtimes,
  # sshdcoker use the `__fallback` runtime.
  _fallback:
    # ${SSHDOCKER_SSH_USER} is a varibale that replaced to ssh login name.
    image: ${SSHDOCKER_SSH_USER}

  
  # You can add more runtimes...

```

## See Also

This project is inspired by https://github.com/gliderlabs/ssh/tree/master/_examples/ssh-docker

## Author

Kohki Makimoto <kohki.makimoto@gmail.com>

## License

The MIT License (MIT)



