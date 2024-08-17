# sendmail-to-matrix

Forward locally received emails (admin notifications, etc) to a Matrix room.

THIS PROJECT IS IN ALPHA, BUT IT WORKS AND I USE IT :)

## Background

Proxmox can only send email notifications. Since they didn't work with the
email service I wanted to use, I decided to use Matrix (which I use elsewhere
anyway).

I wanted a minimal one-way bridge, wherein I could configure Proxmox to send
email notifications to a local email address (`matrix@localhost` in my case)
and the body of each email would be forwarded to my Matrix room.

## Installation

This project is written in Go and is released as a statically-linked binary,
meaning it has no dependencies.

The recommended installation method is to download the binary from the releases
page. Then, ensure that `sendmail-to-matrix` is executable:

```bash
chmod +x /path/to/sendmail-to-matrix
```

Alternatively, build it yourself:

```bash
git clone https://github.com/tnyeanderson/sendmail-to-matrix.git
cd sendmail-to-matrix
CGO_ENABLED=0 go build .
```

## Configuration

This program uses viper for its configuration. Values can be read from a config
file (`config.json`) or can be provided as CLI flags.

The default configuration directory is `/etc/sendmail-to-matrix`. This can be
adjusted with the `--config-dir` flag.

For full usage, see:

```bash
sendmail-to-matrix --help
```

### One-time setup

By default, this program sends encrypted Matrix messages. This requires an
account recovery code to enable device verification, and a persistent SQLite
database to store the state machine used for encryption.

To perform the required one-time configuration setup for encrypted messaging,
run the interactive configuration utility:

```bash
sendmail-to-matrix setup
```

You can disable encryption using `--no-encrypt`. In this mode, no config files
are required if all values are provided as flags. However it is usually more
convenient to save the values in a config file anyway.

To create a config file for unencrypted messaging:

```bash
sendmail-to-matrix setup --no-encrypt
```

### Sendmail forwarding configuration

Add the following line to `/etc/aliases` (or to `~/.forward`) to forward emails
sent to `myuser@localhost`:

```bash
myuser: "|/path/to/sendmail-to-matrix --config-file /path/to/config.json"
```

> NOTE: Be sure to add `--no-encrypt` if you are not using encryption.

Reload your aliases:

```bash
newaliases
```

## Testing

To test that emails are being forwarded properly, use `sendmail` (press
`CTRL+D` after you have finished typing your message):

```bash
$ sendmail myuser@localhost
> Subject: THIS IS NOT A TEST
>
> A song by Bikini Kill
```

You should receive the following message in your Matrix room:

```text
Subject: THIS IS NOT A TEST
A song by Bikini Kill
```

Alternatively, you can test with a file that contains an email in standard
Linux mailbox form.

```bash
cat email.txt | /path/to/sendmail-to-matrix
```

You're done! Direct any administration-related emails (Proxmox notifications,
sysadmin stuff, monitoring, the works) to `myuser@localhost` (or whatever you
created as your alias) and enjoy getting notifications in a modern way.

## Caveats

- HTML tags are removed from the parts of type `text/html`
- Messages with type `multipart/alternative` will prefer `text/plain`.
- Leading and trailing newlines are removed from the message before sending.
