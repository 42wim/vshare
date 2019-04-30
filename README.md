# vshare
## Use-case
Share read-once and time-limited secrets, configurations or small files with your coworkers or customers.

vshare uses [vault](https://www.vaultproject.io) and [response-wrapping tokens](https://www.vaultproject.io/docs/concepts/response-wrapping.html) to simplify this.

## Requirements
* [vault](https://www.vaultproject.io)

## Flow
Vshare will use your `VAULT_TOKEN` and `VAULT_ADDR` environment variables to put the secret in the `cubbyhole/vshare-nanosecondtimestamp` key.
We will then read the `cubbyhole/vshare-nanosecondtimestamp` to get a response-wrapping token with your specified TTL (default 15m) which can be used only once to get the secret back.

## Download
Latest binaries [here](https://github.com/42wim/vshare/releases/latest)

## Example - User A wants to share a secret with user B

### User A
#### Set environment
Needs the vault server address and a valid token

```
export VAULT_ADDR=https://yourvaultserver.tld
export VAULT_TOKEN=avaulttoken
```

#### Enter the secret you want to share
```
$ vshare
Enter your secret to share: mysecret
This secret will self destruct when read or after 15m.

The token to share is s.o6dEYyNgjFEVJMtbGHHjLlzA
```
### User B
#### Set environment
Needs the vault server address

```
export VAULT_ADDR=https://yourvaultserver.tld
```

#### Use the token to get the secret back
```
$ vshare s.o6dEYyNgjFEVJMtbGHHjLlzA
The secret is: mysecret
```

## Example with website: User A wants to share a secret with user B
User A wants to share a secret with user B who maybe doesn't have direct vault access

### Set environment for webserver

Needs the vault server address and the hostname vshare is running on.
vshare doesn't support TLS directly so put it behind a proxy eg caddy.

```
export VAULT_ADDR=https://yourvaultserver.tld
export VSHARE_URL=https://yourvshareserver.tld
```
### Run the webserver

```
vshare -web -webport ":8080"
vault server: https://yourvaultserver.tld
vshare server URL: https://yourvshareserver.tld
⇨ http server started on [::]:8080
```

### User A
#### Set environment for cli
```
export VAULT_ADDR=https://yourvaultserver.tld
export VSHARE_URL=https://yourvshareserver.tld
export VAULT_TOKEN=avaulttoken
```

#### Enter the secret you want to share

```
Enter your secret to share: mysecret
This secret will self destruct when read or after 15m.

The token to share is s.ikctGfmUZ9Yv3wEmUfTtsfyH
Or via web https://yourvshareserver.tld/info/s.ikctGfmUZ9Yv3wEmUfTtsfyH
```

### User B

#### Use the vshare URL from user A
https://yourvshareserver.tld/info/s.ikctGfmUZ9Yv3wEmUfTtsfyH

![image](https://user-images.githubusercontent.com/1810977/56870119-d0fd7700-6a0a-11e9-8d59-8e9e6aca2c37.png)


## Example - advanced
### TTL of 1 hour

```
$ vshare -ttl 1h
Enter your secret to share: blah
This secret will self destruct when read or after 1h.

The token to share is s.2iP1z415d64THmM8iksUSxvU
Or via web https://yourvshareserver.tld/info/s.2iP1z415d64THmM8iksUSxvU
```
### Add a zipfile (maximum 400KB)

```
$ ls -al mystuff.zip
-rw-r--r-- 1 wim wim 14269 Apr 28 21:16 mystuff.zip

$ vshare -ttl 30m -file mystuff.zip
This secret will self destruct when read or after 5m.

The token to share is s.R9UKV5WAdQhXRDGo4EYqzrcc
Or via web https://yourvshareserver.tld/info/s.R9UKV5WAdQhXRDGo4EYqzrcc
```

When retrieving it, vshare knows it's a file and will prompt you to save it instead of showing on stdout

```
$ vshare s.R9UKV5WAdQhXRDGo4EYqzrcc
The secret is a file, enter a filename to safe the output in (default: "vshare.output"): secret.zip

$ ls -al secret.zip
-rw------- 1 wim wim 14269 Apr 28 21:19 secret.zip
```

## usage

```
./vshare: wraps your secret in a single-use token.
Needs VAULT_ADDR and VAULT_TOKEN env variable.

./vshare                        :will ask you for your secret
./vshare <token>                :will reveal the secret wrapped in <token>

flags:
  -file string
        will use the content of <file> as the secret
  -ttl string
        ttl of the token (default "15m")
  -web
        run as webserver. Needs VAULT_ADDR and VSHARE_URL env variable
  -webport string
        <ip>:<port> to listen on as webserver (default ":80")
```
