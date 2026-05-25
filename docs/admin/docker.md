# Deploy with Docker images

Visit [GitHub Container registry](https://github.com/EdwardJXLi/rgrok/pkgs/container/rgrokd) to see all available images and tags.

Image versions:
  - Every released version has its own tag, e.g. `ghcr.io/EdwardJXLi/rgrokd:1.1.4`.
  - The `latest` tag is an alias for the latest released version.
  - The `insiders` tag is the image version built from the latest `main` branch.

## Standalone Docker container

1. Pick a directory on the file system to store the configuration file (`rgrokd.yml`), e.g. `/srv/rgrokd`:
    ```sh
    mkdir -p /srv/rgrokd
    ```
1. Create the configuration file (`/srv/rgrokd/rgrokd.yml`):
    ```yaml
    external_url: "http://example.com"
    web:
      port: 3320
    proxy:
      port: 3000
      scheme: "http"
      domain: "example.com"
    sshd:
      port: 2222

    database:
      # Use "host.docker.internal" if your PostgreSQL is running locally on the same host.
      host: "localhost"
      port: 5432
      user: "REDACTED"
      password: "REDACTED"
      database: "rgrokd"

    identity_provider:
      type: "oidc"
      display_name: "Google"
      issuer: "https://accounts.google.com"
      client_id: "REDACTED"
      client_secret: "REDACTED"
      field_mapping:
        identifier: "email"
        display_name: "name"
        email: "email"
    # # The required domain name, "field_mapping.email" is required to set for this to work.
    #  required_domain: "example.com"
    ```
1. Start a Docker container.
   1. To only allow HTTP tunneling:
       ```sh
       docker run \
         --detach \
         --restart always \
         --volume /srv/rgrokd:/var/opt/rgrokd \
         --publish 3320:3320 \
         --publish 3000:3000 \
         --publish 2222:2222 \
         --name rgrokd \
         ghcr.io/EdwardJXLi/rgrokd:latest
       ```
   1. If you want to allow tunneling raw TCP traffic (this only works on Linux, but [expose port range in Docker is just too slow](https://github.com/moby/moby/issues/14288)):
       ```sh
       docker run \
         --detach \
         --restart always \
         --volume /srv/rgrokd:/var/opt/rgrokd \
         --network host \
         --name rgrokd \
         ghcr.io/EdwardJXLi/rgrokd:latest
       ```

### Upgrade

```sh
docker stop rgrokd
docker rm rgrokd
docker run ...
```

## Docker Compose

> **Note**: The [`docker-compose.yml`](../../docker-compose.yml) file lives under the repository root.

1. Create the directory to store the configuration file (`rgrokd.yml`):
    ```sh
    mkdir -p ./rgrokd
    ```
1. Create the configuration file (`./rgrokd/rgrokd.yml`):
    ```yaml
    external_url: "http://example.com"
    web:
      port: 3320
    proxy:
      port: 3000
      scheme: "http"
      domain: "example.com"
    sshd:
      port: 2222

    database:
      # This connects to the "postgres" service.
      host: "postgres"
      port: 5432
      # Make sure to match the value of the environment variable "POSTGRES_USER"
      user: "REDACTED"
      # Make sure to match the value of the environment variable "POSTGRES_PASSWORD"
      password: "REDACTED"
      database: "rgrokd"

    identity_provider:
      type: "oidc"
      display_name: "Google"
      issuer: "https://accounts.google.com"
      client_id: "REDACTED"
      client_secret: "REDACTED"
      field_mapping:
        identifier: "email"
        display_name: "name"
        email: "email"
    # # The required domain name, "field_mapping.email" is required to set for this to work.
    #  required_domain: "example.com"
    ```
1. Start the cluster:
    ```sh
    POSTGRES_USER=REDACTED \
    POSTGRES_PASSWORD=REDACTED \
    docker-compose up --detach
    ```
