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
git clone https://github.com/tnyeanderson/sendmail-to-matrix
cd sendmail-to-matrix
cp sendmail-to-matrix.py /app/sendmail-to-matrix.py
```


## Configuration

You must add a credential file that will be used by the script.

First, obtain an access token:
```bash
curl -XPOST -d '{"type":"m.login.password", "user":"example", "password":"wordpass"}' "https://homeserver:8448/_matrix/client/r0/login"
```

Then, copy `credentials.json.example` from this repo and edit it for your needs:
```bash
cp credentials.json.example /app/credentials.json

# Don't forget to edit the file!
```

> NOTE: At the moment, you must place this file at `/app/credentials.json`. Soon this will be a shell parameter for the script...

Your credentials file might look like this:
```json
{
  "homeserver": "https://matrix.org",
  "access_token": "<your_access_token>",
  "room_id": "!myroomid:matrix.org",
  "preface": "SENT FROM MY HOMELAB"
}
```

Finally, add the following line to `/etc/aliases` to pipe emails sent to `myuser@localhost` to the script:
```bash
myuser: "|/app/sendmail-to-matrix.py"
```

## Testing

To test that emails get forwarded properly, use `sendmail` (press `CTRL+D` after you have finished typing your message):
```bash
$ sendmail myuser@localhost
> Subject: THIS IS NOT A TEST
> A song by Bikini Kill

```

Since subject lines are ignored, you should receive the following message in your Matrix room (based on the example configuration above):
```
SENT FROM MY HOMELAB

A song by Bikini Kill

```

Alternatively, you can test with a file that contains an email in standard Linux mailbox form.
```bash
cat email.txt | python3 /app/sendmail-to-matrix.py
```

You're done! Direct any administration-related emails (Proxmox notifications, sysadmin stuff, monitoring, the works) to `myuser@localhost` (or whatever you created as your alias) and enjoy getting notifications in a modern way.


## Caveats

- HTML tags are removed from the script if present (and not malformed... looking at you Proxmox). Multipart emails will prefer `text/plain`.