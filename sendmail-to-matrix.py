#!/usr/bin/python3

import asyncio
import json
import sys
import xml.etree.ElementTree as ET
import email.message, email.policy
from nio import AsyncClient

"""CONFIG_FILE format:
{
	"homeserver": "https://matrix.example.org",
	"access_token": "<access token>"
	"room_id": "!roomid:homeservername"
	"preface": "Preface to message"
}
"""
CONFIG_FILE = "/app/credentials.json"


def get_email_body(remove_html = False):
	# Order of MIME parts. Prefer plain text
	mime_pref = ('plain', 'html', 'related')

	# Save the piped input from STDIN as a string
	mail = ''.join(line for line in sys.stdin)

	# Parse into an EmailMessage
	msg = email.message_from_string(mail, policy=email.policy.default)

	# Get the body of the email
	body = msg.get_body(mime_pref).get_content()

	if remove_html:
		body = remove_html_tags(body)

	# Return the body
	return body


def remove_html_tags(text):
	try:
		# Remove XML/HTML tags
		return ''.join(ET.fromstring(text).itertext())
	except:
		# If there is an error, just return the text as-is.
		# It probably isn't valid XML/HTML
		return text


def build_message(preface):
	# Add a new line after the preface, if there is one
	if len(preface) > 0:
		preface = preface + "\n"

	return preface + get_email_body(True)


async def main() -> None:
	# Read the credentials file
	with open(CONFIG_FILE, "r") as f:
		config = json.load(f)
		client = AsyncClient(config['homeserver'])

		# Get the properties from the config file
		client.access_token = config['access_token']
		room_id = config['room_id']
		preface = config.get('preface', '')

	# The message to send
	msg = build_message(preface) 

	# Send the message
	await client.room_send(
		room_id,
		message_type="m.room.message",
		content={
			"msgtype": "m.text",
			"body": msg 
		}
	)

	# Close the client connection after we are done with it.
	await client.close()

# This runs the async main() function
asyncio.get_event_loop().run_until_complete(main())
