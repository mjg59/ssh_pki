PKI certificates for SSH
========================

Introduction
------------

SSH certificates are limited in their usefulness - a certificate can only be
signed with a single CA key, so no chains of trust can be established, and
there's no way to tie them into the global PKI. But what if 🥺?

Should I use this?
------------------

No.

How do I use this?
------------------

Generate a CSR:

```
openssl req -nodes -newkey rsa:2048 -keyout PRIVATEKEY.key -out MYCSR.csr
```

and set the CN to your username. Get it signed somehow. Copy PRIVATEKEY.key to ~/.ssh/id_badidea and run:

```
ssh-keygen -f ~/.ssh/id_badidea -y >~/.ssh/id_badidea.pub
```

Take your signed certificate and encode it to base64 - if it's PEM encoded, convert to DER first:

```
openssl -inform pem -in signed.crt -outform der -out signed.der
```

```
base64 <signed.der >/tmp/encoded.crt
```

Generate a self-signed SSH certificate that embeds the base64 encoded certificate:

```
ssh-keygen -I badidea -s ~/.ssh/id_badidea -n $USER -O clear -O extension:x509=$(cat /tmp/encoded.crt) ~/.ssh/id_badidea.pub
```

and add it to your SSH agent:

```
ssh-add ~/.ssh/id_badidea
```

On your SSH server, add an AuthorizedKeysCommand to sshd_config:

```
AuthorizedKeysCommand /usr/local/bin/ssh_pki -certificate %k -user %i -rootCA /etc/ssh/ssh_root_ca
```

where ssh_root_ca is the root of the infrastructure used to sign the X509 cert.

How it works
------------

The ssh_pki agent examines the certificate presented to it and extracts the
X509 certificate from the extensions field. It ensures that this certificate
has a chain of trust to the configured root CA, and then extracts the
subject CN to verify that it matches the username of the account being
logged into. If everything checks out, it sends a response to the SSH daemon
telling it that the public key used to sign the SSH certificate is a
certificate authority. Since the SSH certificate is self-signed, this
results in the daemon accepting the presented certificate as evidence of
user identity.

So, should I use this?
----------------------

No.

Todo
----

Any sort of security analysis at all. The use of CN is entirely
inappropriate here, but the only reason I wrote this is because I realised I
could.