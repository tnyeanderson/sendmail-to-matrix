# sendmail-to-matrix

Forward locally received emails (admin notifications, etc) to a Matrix room.

THIS PROJECT IS IN ALPHA, BUT IT WORKS AND I USE IT :)


## Background

Proxmox can only send email notifications. Since they didn't work with the email service I wanted to use, I decided to use Matrix (which I use elsewhere anyway).

I wanted a minimal one-way bridge, wherein I could configure Proxmox to send email notifications to a local email address (`matrix@localhost` in my case) and the body of each email would be forwarded to my Matrix room.


## Installation

This project is written in go (with only standard libraries) and is released as a statically-linked binary meaning it has no dependencies.

The recommended installation method is to download the binary from the releases page.

Alternatively, build it yourself:
```bash
git clone https://github.com/tnyeanderson/sendmail-to-matrix.git
cd sendmail-to-matrix
go build sendmail-to-matrix.go
```


## Configuration

You must add a config file that will be used by the script, or supply a `server`, `token`, and `room` with command-line parameters.

Values from a config file are overwritten by command-line parameters. See `sendmail-to-matrix -h` for help.

> NOTE: CLI options can be have a few [formats](https://pkg.go.dev/flag#hdr-Command_line_flag_syntax), but it is recommended to use the double-hyphen syntax (`--config-file` instead of `-config-file`) for consistency with other standard applications.

First, obtain an access token:
```bash
curl -XPOST -d '{"type":"m.login.password", "user":"example", "password":"wordpass"}' "https://homeserver:8448/_matrix/client/r0/login"
```

Then, copy `config.json.example` from this repo and edit it for your needs:
```bash
cp config.json.example /app/config.json

# Don't forget to edit the file!
```

> Note: This example places the `config.json` file in the `/app` folder. You can place it anywhere the script can read from as long as you specify it using `--config-file /path/to/config.json`

Your config file might look like this:
```json
{
  "server": "https://matrix.org",
  "token": "<your_access_token>",
  "room": "!myroomid:matrix.org",
  "preface": "Sent from my homelab"
}
```

Add the following line to `/etc/aliases` to pipe emails sent to `myuser@localhost` to the script:
```bash
myuser: "|sendmail-to-matrix --config-file /app/config.json"
```

> Note: The alias can also be added to the user's `~/.forward` file.

Reload your aliases
```bash
newaliases
```


## Testing

To test that emails get forwarded properly, use `sendmail` (press `CTRL+D` after you have finished typing your message):
```bash
$ sendmail myuser@localhost
> Subject: THIS IS NOT A TEST
>
> A song by Bikini Kill
```

You should receive the following message in your Matrix room (based on the example configuration above):
```
Sent from my homelab
Subject: THIS IS NOT A TEST
A song by Bikini Kill
```

Alternatively, you can test with a file that contains an email in standard Linux mailbox form.
```bash
cat email.txt | sendmail-to-matrix
```

You're done! Direct any administration-related emails (Proxmox notifications, sysadmin stuff, monitoring, the works) to `myuser@localhost` (or whatever you created as your alias) and enjoy getting notifications in a modern way.


## Caveats

- HTML tags are removed from the message body if present (and not malformed... looking at you Proxmox). Multipart emails will prefer `text/plain`.
