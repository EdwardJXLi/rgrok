![rgrok banner](https://user-images.githubusercontent.com/2946214/227126410-3e9dae19-d0c0-4a96-9040-1322e389c8db.png)

<div align="center">
  <h3>Poor man's ngrok — random-subdomain fork</h3>
</div>

## What?

rgrok is a multi-tenant HTTP/TCP reverse tunnel solution through remote port forwarding from the SSH protocol. It is a server-side fork of [pgrok](https://github.com/pgrok/pgrok); the upstream `pgrok` client binary continues to work against an `rgrokd` server (and vice versa for the wire-protocol-compatible parts).

The fork's defining differences from upstream:

- **Per-tunnel cryptographically random subdomains** by default — each `rgrok http` (or `pgrok http`) invocation gets a fresh unguessable URL. The upstream per-user stable-subdomain mode is still available behind `proxy.subdomain_strategy: stable`.
- **SQLite by default** — no external database required. Postgres and MySQL are still supported via `database.type`.

This is intended for small teams that need to expose the local development environment to the public internet, and you need to bring your own domain name and SSO provider. Gated by your SSO through OIDC protocol.

Think of this as a bare-bones alternative to commercial tunneling services. Trying to put this behind a production system will blow up your SLA. For individuals and production systems, just buy ngrok.

## Why?

Stable subdomains and SSO are two things too expensive elsewhere. Random unguessable URLs are nice on top.

Copy, paste, and run is the best UX for everyone.

## How?

Before you get started, make sure you have the following:

1. A domain name (e.g. `example.com`, this will be used as the example throughout this section).
1. A server (dedicated server, VPS) with a public IP address (e.g. `111.33.5.14`).
1. An SSO provider (e.g. Google, JumpCloud, Okta, GitLab, Keycloak) that allows you to create OIDC clients.
1. *(Optional)* A PostgreSQL or MySQL server if you don't want the default SQLite backend.

> [!NOTE]
> 1. All values used in this document are just examples, substitute based on your setup.
> 1. All examples in this document use HTTP for brevity, you may refer to our example walkthrough of [setting HTTPS with Caddy and Cloudflare](./docs/admin/https.md).

### Set up the server (`rgrokd`)

1. Add the following DNS records for your domain name:
    1. `A` record for `example.com` to `111.33.5.14` (with **DNS only** if using Cloudflare)
    1. `A` record for `*.example.com` to `111.33.5.14` (with **DNS only** if using Cloudflare)
1. Set up the server with the [single binary](./docs/admin/single-binary.md), [Docker](./docs/admin/docker.md#standalone-docker-container) or [Docker Compose](./docs/admin/docker.md#docker-compose).
1. Alter your network security policy (if applicable) to allow inbound requests to port `2222` from `0.0.0.0/0` (anywhere).
1. [Download and install Caddy 2](https://caddyserver.com/docs/install) on your server, and use the following Caddyfile config:
    ```caddyfile
    http://example.com {
        reverse_proxy * localhost:3320
    }

    http://*.example.com {
        reverse_proxy * localhost:3000
    }
    ```
1. Create a new OIDC client in your SSO with the **Redirect URI** to be `http://example.com/-/oidc/callback`.

### Set up the client (`rgrok`)

1. Go to http://example.com, authenticate with your SSO to obtain a token. With the default random-subdomain strategy, the URL is assigned per session (printed by the client on connect); with the stable strategy, your dashboard shows your fixed subdomain (e.g. `http://unknwon.example.com`).
1. Download the latest version of `rgrok` from the [Releases](https://github.com/EdwardJXLi/rgrok/releases) page. Upstream `pgrok` binaries also work — the wire protocol is unchanged.
1. Initialize a `rgrok.yml` file (assuming you want to forward requests to `http://localhost:3000`):
    ```sh
    rgrok init --remote-addr example.com:2222 --forward-addr http://localhost:3000 --token {YOUR_TOKEN}
    ```
    - By default, the config file is created under the [standard user configuration directory (`XDG_CONFIG_HOME`)](https://github.com/adrg/xdg):
        - macOS: `~/Library/Application Support/rgrok/rgrok.yml`
        - Linux: `~/.config/rgrok/rgrok.yml`
        - Windows: `%LOCALAPPDATA%\rgrok\rgrok.yml`
    - Use `--config` flag to specify a different path for the config file.
1. Launch the client by executing the `rgrok` or `rgrok http` command.
    - By default, `rgrok` expects the `rgrok.yml` is available under the standard user configuration directory, or under the home directory (`~/.rgrok/rgrok.yml`). Use `--config` flag to specify a different path for the config file.
    - Use the `--debug` flag to turn on debug logging.
    - Upon successful startup, you should see a log looks like:
        ```
        🎉 You're ready to go live at http://a8k3n2m4p9qr.example.com! remote=example.com:2222
        ```
1. Now visit the URL.

As a special case, the first argument of the `rgrok http` can be used to specify forward address, e.g.

```
rgrok http 8080
```

#### Raw TCP tunnels

> [!IMPORTANT]
> You need to alter the server network security policy (if applicable) to allow additional inbound requests to port range 10000-15000 from `0.0.0.0/0` (anywhere).

Use the `tcp` subcommand to tunnel raw TCP traffic:

```
rgrok tcp 5432
```

Upon successful startup, you should see a log looks like:

```
🎉 You're ready to go live at tcp://example.com:10086! remote=example.com:2222
```

The assigned TCP port on the server side is semi-stable, such that the same port number is used when still available.

#### Override config options

Following config options can be overridden through CLI flags for both `http` and `tcp` subcommands:

- `--remote-addr, -r` -> `remote_addr`
- `--forward-addr, -f` -> `forward_addr`
- `--token, -t` -> `token`

#### HTTP dynamic forwards

Typical HTTP reverse tunnel solutions only support forwarding requests to a single address, `rgrok` can be configured to have dynamic forward rules when tunneling HTTP requests.

For example, if your local frontend is running at `http://localhost:3000` but some gRPC endpoints need to talk to the backend directly at `http://localhost:8080`:

```yaml
dynamic_forwards: |
  /api http://localhost:8080
  /hook http://localhost:8080
```

Then all requests prefixed with the path `/api` and `/hook` will be forwarded to `http://localhost:8080` and all the rest are forwarded to the `forward_addr` (`http://localhost:3000`).

### Vanilla SSH

Because the standard SSH protocol is used for tunneling, you may well just use the vanilla SSH client.

1. Go to http://example.com, authenticate with your SSO to obtain the token. (Note: vanilla SSH can only learn the bound port, not the assigned random subdomain — you'll need the `rgrok` or upstream `pgrok` client for that.)
1. Launch the client by executing the `ssh -N -R 0::3000 example.com -p 2222` command:
    1. Enter the token as your password.
    1. Use the `-v` flag to turn on debug logging.
    1. Upon successful startup, you should see a log looks like:
        ```
        Allocated port 22487 for remote forward to :3000
        ```
1. Now visit the URL.

## Explain it to me

![rgrok network diagram](https://user-images.githubusercontent.com/2946214/229048941-cc12139d-f250-49fa-806d-19c27996ee09.png)

## Contributing

Please read through our [contributing guide](.github/contributing.md) and [set up your development environment](docs/dev/local_development.md).

## Credits

This is a fork of [pgrok](https://github.com/pgrok/pgrok) by Joe Chen. The upstream project wouldn't be possible without reading [function61/holepunch-server](https://github.com/function61/holepunch-server), [function61/holepunch-client](https://github.com/function61/holepunch-client), and [TCP/IP Port Forwarding](https://github.com/apache/mina-sshd/blob/master/docs/technical/tcpip-forwarding.md).

## License

This project is under the MIT License. See the [LICENSE](LICENSE) file for the full license text.
