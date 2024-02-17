# Rudi LDA

This is a local mail delivery agent (LDA) that I use for sorting/processing my incoming e-mail.
The LDA sits between chasquid and my Maildir directory, which is then served by Dovecot. Notably,
the whole idea here is to _not_ use Dovecot's LDA or LMTP at all, since I want to write my own
code to process my e-mail and executing external scripts from within Dovecot using Sieve just
sucks.

## Installation

You can download a binary for the [latest release on GitHub](https://github.com/xrstf/rudi-lda/releases)
or install via Go:

```bash
go install go.xrstf.de/rudi-lda
```

## Usage

```
Usage of rudi-lda:
      --datadir string         Path to where metrics and other data files should be placed.
      --folder-script string   Rudi script that will be evaluated to determine the target folder for an incoming e-mail.
      --maildir string         Path to the root of the user's Maildir directory.
      --spam-script string     Rudi script that will be evaluated to determine if the incoming e-mail is spam.
  -V, --version                Show version info and exit immediately.
```

### chasquid

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

## License

MIT
