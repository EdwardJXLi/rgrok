# rgrok - cryptographically random pgrok

## What is rgrok?

rgrok is a server-side fork of [pgrok](https://github.com/pgrok/pgrok) that adds **cryptographically random subdomains (i.e. `http://a8k3n2m4.example.com`)** plus a handful of quality-of-life improvements (SQLite & MySQL backends, single-binary deploys, etc.).

Other than that, rgrok behaves just like pgrok — a multi-tenant HTTP/TCP reverse tunnel built on SSH remote port forwarding, gated by your SSO through OIDC protocol.

## Quick start (local)

```sh
git clone https://github.com/EdwardJXLi/rgrok.git
cd rgrok
nix-shell                # or install Go, pnpm, task, overmind manually
# write a minimal rgrokd.yml (see docs/dev/local_development.md)
overmind start           # rgrokd + Vite + mock OIDC, visit http://localhost:3320
```

Full local-dev walkthrough: [`docs/dev/local_development.md`](docs/dev/local_development.md).

## Production setup

The deployment story (DNS, reverse proxy, OIDC client, single binary / Docker / Docker Compose) is **identical to upstream pgrok** — just substitute `rgrok` / `rgrokd` for the binary names and `ghcr.io/EdwardJXLi/rgrokd` for the image.

Follow the upstream guide → **[pgrok README — How?](https://github.com/pgrok/pgrok#how)**

For rgrok-specific bits:

- Docker image / tags: [`docs/admin/docker.md`](docs/admin/docker.md)
- Single binary: [`docs/admin/single-binary.md`](docs/admin/single-binary.md)
- HTTPS with Caddy + Cloudflare: [`docs/admin/https.md`](docs/admin/https.md)
- Example config: [`rgrokd.example.yml`](rgrokd.example.yml)

## Client setup

**Just install the upstream `pgrok` client** the server-side changes are fully compatible with the vanilla client.

```sh
# macOS / Linux
brew install pgrok
# or download a binary from https://github.com/pgrok/pgrok/releases
```

Then point it at your rgrokd server (substitute `example.com` for your own domain and `{YOUR_TOKEN}` for the token shown in the rgrokd web UI after SSO login):

```sh
pgrok init \
  --remote-addr example.com:2222 \
  --forward-addr http://localhost:3000 \
  --token {YOUR_TOKEN}

pgrok http        # or `pgrok http 8080` to forward a different port
pgrok tcp 5432    # raw TCP tunnel
```

On startup you should see a log line like:

```
🎉 You're ready to go live at http://a8k3n2m4.example.com! remote=example.com:2222
```
