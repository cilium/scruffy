# Scruffy

Scruffy is a tool designed to garbage collect unused tags on quay.io. This
tool assumes that the tags that should be kept in such repositories are only the
ones that match a commit SHA for any of the branches provided by
`--stable-branches`.

## Features

- [X] Garbage collect unused tags on quay.io

## Missing features

- [ ] None :-)

# Build

```bash
make scruffy
```

# Usage

1. Login into quay.io with a bot account, for example `ciliumbot`. **Note** Your organization should create its own bot account.
2. Go to the organization that holds the images that you want to delete, for example [Quay Cilium Organization](https://quay.io/organization/cilium?tab=applications)
3. Invite that bot account, for example `ciliumbot`, as an admin of that organization.
4. Go to [OAuth Applications](https://quay.io/organization/cilium?tab=applications)
5. Create a new Application with the name "GC CI Images"
6. In the Application settings generate a new Token with the permissions "Read/Write to any accessible repositories".
7. Copy the access token in a safe place until so that it can be used as a
   GitHub secret for the GitHub action.
8. Remove the bot account from the list of owners / admin of that organization.
9. Verify that the bot account only has access to the repositories that it
   should delete tags from by checking https://quay.io/repository/ while logged
   in with that bots account.
10. Set the token as a secret in the GitHub repository, `QUAY_TOKEN` and deploy
    the GitHub action.

# GitHub action

```yaml
name: Scruffy
on:
  workflow_dispatch:
  schedule:
    # Run the GC every monday at 9am
    - cron: "0 9 * * 1"

permissions: read-all

jobs:
  scruffy:
    # if: github.repository == '<my-org>/<my-repo>'
    name: scruffy
    runs-on: ubuntu-20.04
    steps:
      - uses: actions/checkout@5a4ac9002d0be2fb38bd78e4b4dbde5606d7042f
      - uses: docker://quay.io/cilium/scruffy:v0.0.1@sha256:15e3926d8e74aa6a278cc07fb61d5888322fabdae49637384dc6a3fb32452969
        with:
          entrypoint: scruffy
          args: --git-repository=./
        env:
          QUAY_TOKEN: ${{ secrets.QUAY_TOKEN }}
```
