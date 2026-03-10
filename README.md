UCR Learning Services contains infrastructure for deploying web services alongside the services to be deployed.

# Quickstart

To run the infrastructure with the Agent Arena service, you must ensure that the appropriate environment variables are set then run

```bash
docker compose up
```

from the root of this repository.
This will build and download the necessary container images and run them.

Instead of setting environment variables, you can create a `.env` file in the root directory of this repository and populate it with the necessary environment variable definitions.
As of the time of writing, the following is a `.env` template that includes all necessary environment variables:

```bash
OAUTH2_PROXY_CLIENT_SECRET=
OAUTH2_PROXY_CLIENT_ID=
OAUTH2_PROXY_COOKIE_SECRET=

EXTERNAL_HOST=
EXTERNAL_PORT=
```

# No-Docker Setup

Even without using docker, the docker-compose.yaml is useful as a guide.
Since each service may have its own set of environment variables that are important to set for the function of the entire system, [environment_setup](./environment_setup/) is a utility written to track environment variables needed by various parts of the infrastructure and services.
It looks for hosts.txt and env_vars.txt files recursively from a starting point to provide a unified source of all hosts/variables that need to be set.

Any file that ends in `.template` has one or more environment variable references that need to be dereferenced into a value to form the final version of the file, which should be named with the `.template` removed.
The references look like `${EXTERNAL_HOST}` and, given that the environment variable is set, `envsubst` may be used to swap in the values assigned to the variables.

The [compile-configurations.bash](./compile-configurations.bash) script is provided to search through the repository and perform this "compilation" of the files that need environment variable values substituted in.

# Code Architecture

## Folders

 - [./infrastructure/](./infrastructure/) contains configurations for infrastructure.
 - [./internal/](./internal/) contains source code for infrastructure and libraries used by the rest of the repository.
 - [./services/](./services/) contains source code for web services that leverage the infrastructure.

## Components 

- NGINX is the reverse proxy behind which all web services lie.
- [OAuth2-Proxy](https://github.com/oauth2-proxy/oauth2-proxy), which is configured [here](./infrastructure/oauth2-proxy/) handles user authentication.
- [Taxis](./internal/taxis), which is configured [here](./infrastructure/nginx/), wraps around an authentication (here OAuth2.0 Proxy) endpoint to append group headers.
- Services lie inside [./services](./services/). Each service contains a `deploy/` folder that has an NGINX config that may be included to use the service.

## Description

A web request from a user first hits NGINX,
which can then route the request through Taxis to add UID and group headers to the request.
The request is finally sent to the service with the added headers,
the response for which is sent to the user through NGINX. See the figure below



```
                                      AuthN                 
                         ┌──────────┐     ┌──────────────┐
                         │          ┼─────►              │
                         │  taxis   │     │ OAuth2 Proxy │
                         │          ◄─────┼              │
                         └──▲───┬───┘ UID └──────────────┘
                            │   │                           
               AuthN/AuthZ  │   │ UID + groups              
                            │   │     request + UID + groups
                        ┌───┼───▼────┐          ┌──────────┐
                        │            ┼──────────►          │
request────────────────►│  NGINX     │          │ service  │
                        │            ◄──────────┼          │
response◄───────────────┴────────────┘          └──────────┘
```
