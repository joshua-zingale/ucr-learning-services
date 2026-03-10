Taxis (Greek τάξις) is a light-weight reverse proxy for an authentication
endpoint that augments the response with any groups to which the request's user
belongs.

Taxis is not designed to be infinitely scalable, but to permit easy modification
of group assignments via the editing of a single YAML-like file.

Taxis is **not** responsible for _authentication_, but for _authorization_.
Therefore, Taxis makes a request to an external authentication server (such as
[OAuth2 Proxy](https://github.com/oauth2-proxy/oauth2-proxy)) and augments its
response with group information.

# Functionality

After receiving a web request, Taxis will route the request first to an
authentication server. If the authentication server responds with a 202 code and
a header containing a user's ID, Taxis will respond with the initial request
both with the user's ID and any groups to which the user belongs. Otherwise, in
the event that the authentication server either responds with a status code that
is not 202 or fails to include the user's ID in the header, Taxis responds with
an error status.

## Running the Server

Use `taxis serve -h` to learn how to configure & run Taxis.

## Request Format

The format for a valid request to Taxis is determined by the authentication
server that is used because Taxis functions as a reverse proxy for the
authentication server.

## Response Format

If a request is authenticated by the authentication server, Taxis will respond
with two headers populated, one which contains the user's ID and one that
contains his groups. Assuming a default configuration of Taxis, these headers
may look like

```
X-Email: tom@univ.edu
X-Groups: cs100.instructor,cs287.student
```

## Assigning Groups

Taxis relies on a very limited subset of YAML as a file to assign groups to
users.

For example,

```
cs100:
    instructor:
        - bob@ucr.edu
    student:
        - tom@ucr.edu
        - dick@ucr.edu
        - harry@ucr.edu
cs9B: ...
```

This will assign bob@ucr.edu, for example, the group cs100.instructor

### YAML-like Format

The file must have map as the root element. Then, each value can either be a map
or a list of strings. Maps and lists of strings may only be specified using the
multi-line syntax as in the example.
