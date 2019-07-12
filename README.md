# Ident

[![#ident:abolivier.bzh on Matrix](https://img.shields.io/matrix/ident:matrix.org.svg?logo=matrix&label=%23ident:abolivier.bzh)](https://matrix.to/#/#ident:abolivier.bzh) [![Build Status](https://travis-ci.org/babolivier/ident.svg?branch=master)](https://travis-ci.org/babolivier/ident) [![codecov](https://codecov.io/gh/babolivier/ident/branch/master/graph/badge.svg)](https://codecov.io/gh/babolivier/ident) 

Ident will be a simple and lightweight identity server for [Matrix](https://matrix.org).

Its aim is to be a full implementation of the [identity server specification](https://matrix.org/docs/spec/identity_service/r0.2.1) and nothing more. The current goal is compliance with the release 0.2.1 of this specification.

## Features

* [x] [Status check](https://matrix.org/docs/spec/identity_service/r0.2.1#status-check)
* [x] [Key management](https://matrix.org/docs/spec/identity_service/r0.2.1#key-management)
* [x] [Invitation storage](https://matrix.org/docs/spec/identity_service/r0.2.1#invitation-storage)
* [x] [Ephemeral invitation signing](https://matrix.org/docs/spec/identity_service/r0.2.1#ephemeral-invitation-signing)
* [ ] [Association creation](https://matrix.org/docs/spec/identity_service/r0.2.1#establishing-associations)
* [ ] [Association deletion](https://matrix.org/docs/spec/identity_service/r0.2.1#post-matrix-identity-api-v1-3pid-unbind)
* [ ] [Association lookup](https://matrix.org/docs/spec/identity_service/r0.2.1#association-lookup)

## Build

```bash
git clone https://github.com/babolivier/ident.git
go build
```

## Configure

Ident needs a configuration file to start. The default location it will look for it at is `config.yaml` at the root of the repository. You can specify an alternative location using the `--config` flag when starting up Ident.

The configuration file needs to follow this structure:

```yaml
ident:
  base_url: "http://127.0.0.1:9999"
  signing_key:
    algo: ed25519
    id: 0
    seed: thees6sha8QueiWu4ooGhais7ahqu1oc # A 32-byte long string
  invites:
    email_template:
      text: "templates/text/invite.txt"
      html: "templates/html/invite.html"
    subject_template: "{{.SenderDisplayName}} invited you to Matrix!"

http:
  listen_addr: "127.0.0.1:9999"

database:
  driver: sqlite3
  conn_string: ident.db

email:
  from: "Ident <ident@example.com>"
  smtp:
    hostname: mail.example.com
    port: 465
    username: "ident@example.com"
    password: somepassword
    enable_tls: true
```

A more detailed documentation on this file will be provided in the future.