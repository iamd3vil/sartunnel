# Sartunnel

A experimental encrypted Point-To-Point VPN created for learning.

## Usage

A public, private keys needs to be generated for both the local and remote instances. `sartunnel genkey` can be used to generate keys.

```bash
$ sudo sartunnel genkey
Public Key: kwdKk9kjd87k1soqzX8WmjmbTguZKGca3x8AjyUqm3A=
PrivateKey: yCsXbJBcJkNSt77ijv6SC34/kvR1lMwDd21ZSrN9RmI=
```

A `config.toml` needs to exist in the same directory as the binary. 

## Configuration

An example `config.toml`:

```toml
[tunnel]
interface = "tun0"
ip_range = "192.168.9.2/24"
private_key = "yCsXbJBcJkNSt77ijv6SC34/kvR1lMwDd21ZSrN9RmI="
local_address = "0.0.0.0:12345"

[peer]
remote_address = "140.60.63.91:41821"
public_key = "YZyw1eskwc6r61c5CjRf8HEjcZCc6DF+LGm8bzB7OCs="
```

- `interface`: Interface name
- `ip_range`: IP Range that needs to be assigned to the interface
- `private_key`: Private Key for the local instance
- `local_address`: If local address is present, other peer will connect to this instance using this IP. If it's present, a UDP server is started.
- `remote_address`: Address of the other peer to connect to. The other peer needs to listen on this IP and should be available to connect through UDP.
- `public_key`: Public Key of the other group

## Crypto

ECDH is used for the key exchange and XSalsa20 & Poly1305 is used for the authenticated encryption. 

> Note: There is no perfect forward secrecy. The key is not rotated while the session is going on. This is not implemented to keep it simple.