#!/usr/bin/python3

import asyncio
import json
import sys
import argparse
import xml.etree.ElementTree as ET
import email.message, email.policy
from nio import AsyncClient


config_file_format="""

Config file format (JSON):
{
	"server": "https://matrix.example.org",
	"token": "<access token>"
	"room": "!roomid:homeservername"
	"preface": "Preface to message"
}
 
"""

# Parse command line arguments. These override config file values.
parser = argparse.ArgumentParser(
	description='Take an email message via STDIN and forward it to a Matrix room',
	epilog='You must define a server, token, and room either using a config file or via command-line parameters.' + config_file_format,
	formatter_class=argparse.RawTextHelpFormatter
)
parser.add_argument('-f', '--config-file',	help='Path to config file')
parser.add_argument('-s', '--server',		help='The matrix homeserver url')
parser.add_argument('-t', '--token',		help='Matrix account access token')
parser.add_argument('-r', '--room',			help='The matrix Room ID')
parser.add_argument('-p', '--preface',		help='Preface the matrix message with arbitrary text (optional)')
args = vars(parser.parse_args())


def get_config():
	required = ['server', 'token', 'room']
	optional = ['preface']

	# First, get the values from the config file
	config = read_config_file(args.get('config_file'))

	# Then replace values in it with command line arguments if available
	for arg in [*required, *optional]:
		config[arg] = args.get(arg) or config.get(arg)

	# Then check they are all set
	for arg in required:
		if config[arg] == None:
			# Throw an error
			raise TypeError('Missing required parameter: ' + arg)

	return config


def read_config_file(config_file):
	if config_file == None:
		return dict()

	# Read the credentials file
	with open(config_file, "r") as f:
		config = json.load(f)
		return config


def get_email() -> email.message.EmailMessage:
	# Save the piped input from STDIN as a string
	mail = ''.join(line for line in sys.stdin)

	# Parse into an EmailMessage
	return email.message_from_string(mail, policy=email.policy.default)


def get_email_body(email, remove_html = True):
	# Order of MIME parts. Prefer plain text
	mime_pref = ('plain', 'html', 'related')

	# Get the body of the email
	body = email.get_body(mime_pref).get_content()

	# Remove html tags if necessary
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


def build_message(config):
	email = get_email()
	body = get_email_body(email)
	subject = email.get('subject') or ''
	preface = config.get('preface') or ''

	# Add a new line after the preface, if there is one
	if len(preface) > 0:
		preface = preface + "\n"

	if len(subject) > 0:
		subject = 'Subject: ' + subject + '\n'

	return preface + subject + body


async def main() -> None:
	# Read the credentials file
	config = get_config()

	# Set up client
	client = AsyncClient(config['server'])
	client.access_token = config['token']

	# The message to send
	msg = build_message(config)

	# Send the message
	await client.room_send(
		config['room'],
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
