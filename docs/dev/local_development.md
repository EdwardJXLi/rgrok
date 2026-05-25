# Set up your development environment

rgrok is built and runs as a single binary and meant to be cross platform. Therefore, you should be able to develop rgrok on any major platform you prefer.

The fastest path is the included Nix shell — `nix-shell` (or `direnv allow`) at the repo root drops you into an environment with all dependencies, and the default SQLite backend means no database setup is needed at all.

## Step 1: Install dependencies

The development of rgrok has the following dependencies:

- [Git](https://git-scm.com/book/en/v2/Getting-Started-Installing-Git) (v2 or higher)
- [Go](https://go.dev/doc/install) (v1.25 or higher)
- [pnpm](https://pnpm.io/installation) (v10 or higher)
- [Task](https://taskfile.dev/installation/) (v3)
- [Overmind](https://github.com/DarthSim/overmind#installation) (v2)
- *(Optional)* [PostgreSQL](https://wiki.postgresql.org/wiki/Detailed_installation_guides) (v10+) or MySQL — only needed if you want to test those backends or run the integration suite.

On macOS via Homebrew:

```bash
brew install git go pnpm go-task overmind
# Optional, only for the postgres backend / integration tests:
# brew install postgresql@15 && brew services start postgresql@15
```

Or on any platform with Nix:

```bash
nix-shell    # everything above (and postgres for integration tests) is provided
```

## Step 2: (Optional) Initialize a PostgreSQL database

Only needed if you set `database.type: postgres` in your config. With the default `sqlite` backend you can skip this step entirely.

```bash
createuser --superuser rgrokd
psql -c "ALTER USER rgrokd WITH PASSWORD 'rgrokd';"
createdb --owner=rgrokd --encoding=UTF8 --template=template0 rgrokd
```

## Step 3: Get the code

```bash
# HTTPS
git clone --depth 10 https://github.com/EdwardJXLi/rgrok.git

# or SSH
git clone --depth 10 git@github.com:EdwardJXLi/rgrok.git
```

> [!NOTE]
> The repository has Go modules enabled, please clone to somewhere outside of your `$GOPATH`.

## Step 4: Initialize `rgrokd.yml`

Create a `rgrokd.yml` file under the repository root and put the following configuration:

```yaml
external_url: "http://localhost:3320"
web:
  port: 3320
proxy:
  port: 3000
  scheme: "http"
  domain: "localhost:3000"
sshd:
  port: 2222

database:
  type: "sqlite"
  path: "./rgrokd.db"

identity_provider:
  type: "oidc"
  display_name: "OIDC"
  issuer: "http://localhost:9833"
  client_id: "winnerwinner"
  client_secret: "chickendinner"
  field_mapping:
    identifier: "email"
    display_name: "name"
    email: "email"
```

If you'd rather use the postgres backend you set up in Step 2, swap the `database:` block for:

```yaml
database:
  type: "postgres"
  host: "localhost"
  port: 5432
  user: "rgrokd"
  password: "rgrokd"
  database: "rgrokd"
```

## Step 5: Start the servers

The following command will start processes defined in the [`Procfile`](../../Procfile) and automatically recompile and restart these servers if related files are changed:

```bash
overmind start
```

Then, visit http://localhost:3320!

Few things to note:

- The web, proxy and SSHD servers of rgrokd are started
- No need to access the Vite server for the rgrokd web app as all requests to it are proxied by the rgrokd web server
- A [mock OIDC server](../../integration-tests/oidc-server/) is started for your convenience
