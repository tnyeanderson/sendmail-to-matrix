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

This project is written in go and is released as a statically-linked binary
meaning it has no dependencies.

The recommended installation method is to download the binary from the releases
page.

Alternatively, build it yourself:
```bash
git clone https://github.com/tnyeanderson/sendmail-to-matrix.git
cd sendmail-to-matrix
CGO_ENABLED=0 go build .
```


## Configuration

You must add a config file that will be used by the script, or supply
a `server`, `token`, and `room` with command-line parameters.

Values from a config file are overwritten by command-line parameters. See
`sendmail-to-matrix --help` for help.

> NOTE: CLI options can have a few different
[formats](https://pkg.go.dev/flag#hdr-Command_line_flag_syntax), but it is
recommended to use the double-hyphen syntax (`--config-file` instead of
`-config-file`) for consistency with other standard applications.

Ensure that `sendmail-to-matrix` is executable:
```bash
chmod +x /path/to/sendmail-to-matrix
```

Generate a config file by following the prompts:
```bash
/path/to/sendmail-to-matrix generate-config
```

> Note: You can place the configuration file anywhere the script can read from
as long as you specify it using `--config-file /path/to/config.json`

Add the following line to `/etc/aliases` to pipe emails sent to
`myuser@localhost` to the script:
```bash
myuser: "|/path/to/sendmail-to-matrix --config-file /path/to/config.json"
```

> Note: The alias can also be added to the user's `~/.forward` file.

Reload your aliases:
```bash
newaliases
```


## Testing

To test that emails get forwarded properly, use `sendmail` (press `CTRL+D`
after you have finished typing your message):
```bash
$ sendmail myuser@localhost
> Subject: THIS IS NOT A TEST
>
> A song by Bikini Kill
```

You should receive the following message in your Matrix room (based on the
example configuration above):
```
Sent from my homelab
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
