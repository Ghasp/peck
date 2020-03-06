import base64
import binascii
import os

def decode_basic_atuh(auth_header: str, encoding: str = 'latin1') -> tuple:
    try:
        auth_type, encoded_credentials = auth_header.split(' ', 1)
    except ValueError:
        raise ValueError('Could not parse authorization header.')

    if auth_type.lower() != 'basic':
        raise ValueError('Unknown authorization method %s' % auth_type)

    try:
        decoded = base64.b64decode(
            encoded_credentials.encode('ascii'), validate=True
        ).decode(encoding)
    except binascii.Error:
        raise ValueError('Invalid base64 encoding.')

    try:
        # RFC 2617 HTTP Authentication
        # https://www.ietf.org/rfc/rfc2617.txt
        # the colon must be present, but the username and password may be
        # otherwise blank.
        username, password = decoded.split(':', 1)
    except ValueError:
        raise ValueError('Invalid credentials.')

    return (username, password)


def test_path(file_path):
    return os.path.exists(file_path)