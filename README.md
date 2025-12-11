# headscale-oidc-sync

A Go application that synchronizes OIDC (via LDAP Provider) groups into the Headscale ACL file, enabling OIDC-authenticated users to be managed via group membership.

*TODO: Use API too*

## Purpose

This tool periodically queries your LDAP server for users and their group memberships, then updates the acl file used by Headscale. It ensures that Headscale ACL groups are kept in sync with LDAP, so access control policies can be managed directly in your directory service.

## Features

- Syncs LDAP groups into Headscale ACL format.
- Configurable LDAP filters and attributes.
- Cron-driven synchronization (configurable interval).
- Optional automatic reload of the Headscale container after ACL updates.
- Dockerized for easy deployment.

## Usage

1. Copy `.env.example` to `.env` and configure your LDAP and application settings.
2. Run the application using:
   - `make run` (for local binary)
   - `make docker-up` (for Docker Compose)

The tool will read users and groups from LDAP, update `acl.json`, and optionally reload the Headscale container.

## Configuration

Copy `.env.example` to `.env` and adjust the values to match your environment.

### Application Configuration

| Variable                      | Default Value                   | Description |
|------------------------------|---------------------------------|-------------|
| `APP_ENV`                    | `production`                    | Application environment (development, test, production) |
| `APP_GROUP_PREFIX`           | `headscale-`                    | Only groups with this prefix will be synced |
| `APP_ACL_JSON`               | `acl.json`                      | Path to the ACL file used by Headscale |
| `APP_IS_RELOAD_HEADSCALE`    | `true`                          | Whether to reload the Headscale container after ACL changes |
| `APP_HEADSCALE_CONTAINER_NAME`| `vpn-hs-headscale-1`            | Name of the Headscale Docker container |
| `APP_CRON_SCHEDULE`          | `@every 10m`                    | Cron schedule for sync jobs (e.g., `@every 10m`, `@daily`) |

### Log Configuration

| Variable            | Default Value   | Description |
|---------------------|-----------------|-------------|
| `LOG_LEVEL`         | `info`          | Logging level (debug, info, warn) |
| `LOG_FORMAT`        | `console`       | Log output format (console, json, text) |

### LDAP Configuration

| Variable                   | Default Value                          | Description |
|----------------------------|----------------------------------------|-------------|
| `LDAP_HOST`                | `ldap.example.com`                     | LDAP server hostname |
| `LDAP_PORT`                | `389`                                  | LDAP server port |
| `LDAP_PROTOCOL`            | `plain`                                | Protocol (plain, ssl, tls, starttls) |
| `LDAP_BIND_DN`             | `cn=admin,dc=example,dc=com`           | LDAP bind DN (service account) |
| `LDAP_BIND_PW`             | `password`                             | LDAP bind password |
| `LDAP_BASE_DN`             | `dc=example,dc=com`                    | Base DN for LDAP queries |
| `LDAP_GROUP_FILTER`        | `(&(objectClass=goauthentik.io/ldap/group))` | LDAP filter for groups |
| `LDAP_USER_FILTER`         | `(&(objectClass=person))`               | LDAP filter for users |
| `LDAP_ATTR_UID`            | `uid`                                  | LDAP attribute for user ID |
| `LDAP_ATTR_USERNAME`       | `cn`                                   | LDAP attribute for username |
| `LDAP_ATTR_EMAIL`          | `mail`                                 | LDAP attribute for email |
| `LDAP_ATTR_GROUPS`         | `memberOf`                             | LDAP attribute for user group memberships |

## Contributing

Pull requests are welcome! As I am still at the beginning of learning Go, please include detailed descriptions with your contributions.
Thank you for your help!