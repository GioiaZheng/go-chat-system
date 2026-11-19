#!/usr/bin/env sh

# Launch an ephemeral Node 24 container wired to the current repository for frontend development.
docker run -it --rm -v "$(pwd):/src" -u "$(id -u):$(id -g)" --network host --workdir /src/webui node:24 /bin/bash
