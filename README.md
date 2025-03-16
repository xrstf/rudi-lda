# Migration note

> [!IMPORTANT]
> Rudi LDA has been migrated to [codeberg.org/xrstf/rudi-lda](https://codeberg.org/xrstf/rudi-lda).

---

## Rudi LDA

This is a local mail delivery agent (LDA) that I use for sorting/processing my incoming e-mail.
The LDA sits between chasquid and my Maildir directory, which is then served by Dovecot. Notably,
the whole idea here is to _not_ use Dovecot's LDA or LMTP at all, since I want to write my own
code to process my e-mail and executing external scripts from within Dovecot using Sieve just
sucks.

### Installation

You can download a binary for the [latest release on GitHub](https://github.com/xrstf/rudi-lda/releases)
or install via Go:

```bash
go install go.xrstf.de/rudi-lda
```

### Usage

```
NAME:
   rudi-lda - Filter e-mails with Rudi and deliver them to Maildirs++

USAGE:
   rudi-lda [global options] [command [command options]] [arguments...]

COMMANDS:
   deliver   delivers e-mail into a Maildir++ folder (default command)
   spamtest  prints spam and folder script results on stdout
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h     show help (default: false)
   --version, -v  print the version (default: false)
```

```
NAME:
   rudi-lda deliver - delivers e-mail into a Maildir++ folder (default command)

USAGE:
   rudi-lda deliver [arguments...]

OPTIONS:
   --maildir value                (required) path to the root of the user's Maildir directory [$RUDILDA_MAILDIR]
   --datadir value                (required) path to where metrics and other data files should be placed [$RUDILDA_DATADIR]
   --from value, -f value         from address
   --destination value, -d value  (required) destination user
   --spam-script value            Rudi script that will be evaluated to determine if the incoming e-mail is spam [$RUDILDA_SPAM_SCRIPT]
   --folder-script value          Rudi script that will be evaluated to determine the target folder for an incoming e-mail [$RUDILDA_FOLDER_SCRIPT]
   --rentablo                     enable the rentablo.de processor (default: false) [$RUDILDA_RENTABLO]
   --sunnyportal                  enable the sunnyportal.de processor (default: false) [$RUDILDA_SUNNYPORTAL]
   --backup-spam                  write spam e-mails to $datadir/spam (default: false) [$RUDILDA_BACKUP_SPAM]
   --help, -h                     show help (default: false)
```

#### chasquid

To use Rudi-LDA as your MDA in [chasquid](https://blitiri.com.ar/p/chasquid/), update your
`chasquid.conf` and set

```yaml
mail_delivery_agent_bin: "rudi-lda"

mail_delivery_agent_args: "deliver"
mail_delivery_agent_args: "-f"
mail_delivery_agent_args: "%from%"
mail_delivery_agent_args: "-d"
mail_delivery_agent_args: "%to_user%"
```

You can then set the remainig configuration for Rudi-LDA using environment variables:

```bash
RUDILDA_SPAM_SCRIPT=/etc/rudi-lda/spam.rudi
RUDILDA_FOLDER_SCRIPT=/etc/rudi-lda/folder.rudi
RUDILDA_MAILDIR=/var/mail
RUDILDA_DATADIR=/var/lib/rudi-lda
RUDILDA_RENTABLO=true
RUDILDA_SUNNYPORTAL=true
RUDILDA_BACKUP_SPAM=true
```

### License

MIT
