![Latest GitHub release](https://img.shields.io/github/release/mashiike/qsgpm.svg)
![Github Actions test](https://github.com/mashiike/qsgpm/workflows/Test/badge.svg?branch=main)
[![Go Report Card](https://goreportcard.com/badge/mashiike/qsgpm)](https://goreportcard.com/report/mashiike/qsgpm) [![License](https://img.shields.io/badge/license-MIT-blue.svg)](https://github.com/mashiike/qsgpm/blob/master/LICENSE)
# qsgpm
A commandline tool for management of QuickSight Group and CustomPermission

## Install

### binary packages

[Releases](https://github.com/mashiike/shimesaba/releases).

## QuickStart

```console 
$ qsgpm --help                       
NAME:
   qsgpm - A commandline tool for management of QuickSight Group and CustomPermission

USAGE:
   qsgpm --config <config file>

VERSION:
   current

COMMANDS:
   help, h  Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --config value, -c value     config file path [$CONFIG, $QSGPM_CONFIG]
   --dry-run                    (default: false) [$QSGPM_DRY_RUN]
   --log-level value, -l value  output log level (debug|info|notice|warn|error) (default: "info") [$QSGPM_LOG_LEVEL]
   --help, -h                   show help (default: false)
   --version, -v                print the version (default: false)
```

The simplest configuration is:

```yaml
required_version: ">=0.0.0"

user:
  namespace: default
groups:
  - all

rules:
  - user:
      role: Admin
    groups:
      - admin

  - user:
      role: Author
    groups:
      - author
    custom_permission: DefaultAuthor

  - user:
      role: Reader
    groups:
      - reader
```

The above setting means that all users will belong to the group "all" and also to the group for each account role, and Author will have a custom permission named DefaultAuthor.

What conditions do the rules match for users? and if they match, which group they belong to and what custom permissions they should have.
The rule matches only one.
The Yaml array is evaluated from the top and the first matching rule is applied to each user.

For example, for a more complex configuration where the QuickSight user is an external user, the following is an example of another configuration.
```yaml
required_version: ">=0.0.0"

user:
  identity_type: IAM
  namespace: default
groups:
  - all

rules:
  - user:
      role: Admin
    groups:
      - admin

   - user:
      identity_type: QuickSight 
      email_suffix: "@internal.example.com"
      role: Author
    groups:
      - internal_author
      - author
    custom_permission: InternalAuthor

   - user:
      role: Author
    groups:
      - external_author
      - author
    custom_permission: ExternalAuthor

  - user:
      session_name_suffix: "@external.example.com"
      role: Author
    groups:
      - external_author
      - author
    custom_permission: ExternalAuthor
  
  - user:
      role: Author
    groups:
      - internal_author
      - author
    custom_permission: InternalAuthor


  - user:
      role: Reader
    groups:
      - reader
```

## LICENSE

MIT
