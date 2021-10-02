# sendmail-to-matrix

Forward locally received emails (admin notifications, etc) to a Matrix room.

THIS PROJECT IS IN ALPHA, BUT IT WORKS AND I USE IT :)


## Background

Proxmox can only send email notifications. Since they didn't work with the email service I wanted to use, I decided to use Matrix (which I use elsewhere anyway).

I wanted a minimal one-way bridge, wherein I could configure Proxmox to send email notifications to a local email address (`matrix@localhost` in my case) and the body of each email would be forwarded to my Matrix room.


## Installation

This project requires `python3` and uses [matrix-nio](https://github.com/poljar/matrix-nio).

Please install these dependencies. For instance, on Debian/Ubuntu:

```bash
# Install python3 and pip
sudo apt install python3 python3-venv python3-pip

# Install matrix-nio with end-to-end encryption support
# E2EE requires libolm-dev
sudo apt install libolm-dev
pip install "matrix-nio[e2e]"
```

Then copy `sendmail-to-matrix.py` somewhere on your machine.
```bash
# Here, we use the /app folder for example
git clone https://github.com/tnyeanderson/sendmail-to-matrix.git
cd sendmail-to-matrix
cp sendmail-to-matrix.py /app/sendmail-to-matrix.py
```


## Configuration

You must add a config file that will be used by the script, or supply a `server`, `token`, and `room` with command-line parameters.

Values from a config file are overwritten by command-line parameters. See `sendmail-to-matrix.py -h` for help.

First, obtain an access token:
```bash
curl -XPOST -d '{"type":"m.login.password", "user":"example", "password":"wordpass"}' "https://homeserver:8448/_matrix/client/r0/login"
```

Then, copy `config.json.example` from this repo and edit it for your needs:
```bash
cp config.json.example /app/config.json

# Don't forget to edit the file!
```

> Note: This example places the `config.json` file in the `/app` folder. You can place it anywhere the script can read from as long as you specify it using `-f /path/to/config.json`

Your config file might look like this:
```json
{
  "server": "https://matrix.org",
  "token": "<your_access_token>",
  "room": "!myroomid:matrix.org",
  "preface": "Sent from my homelab"
}
```

Finally, add the following line to `/etc/aliases` to pipe emails sent to `myuser@localhost` to the script:
```bash
myuser: "|/app/sendmail-to-matrix.py -f /app/config.json"
```

## Testing

To test that emails get forwarded properly, use `sendmail` (press `CTRL+D` after you have finished typing your message):
```bash
$ sendmail myuser@localhost
> Subject: THIS IS NOT A TEST
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
cat email.txt | python3 /app/sendmail-to-matrix.py
```

You're done! Direct any administration-related emails (Proxmox notifications, sysadmin stuff, monitoring, the works) to `myuser@localhost` (or whatever you created as your alias) and enjoy getting notifications in a modern way.


## Caveats

- HTML tags are removed from the message body if present (and not malformed... looking at you Proxmox). Multipart emails will prefer `text/plain`.