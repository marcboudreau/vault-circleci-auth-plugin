# Vault CircleCI Authentication Plugin

This plugin allows a running CircleCI build to authenticate itself with a Vault
server in order to obtain a Vault token associated with policies mapped to the
given CircleCI project.

## Architecture

The authentication is achieved by providing details about the running CircleCI
build that the Vault server can verify using the CircleCI API.

## Configuration

The plugin must be configured with certain parameters prior to being able to
successfully handle login requests.  The plugin is configured by sending a POST
request to the `auth/circleci/config` endpoint.

Request parameters:

* **circleci_token** is a CircleCI personal API token that allows the plugin to make API calls to CircleCI.

* **base_url** (optional) is an alternate base URL where CircleCI API calls are sent. This parameter must end up with a slash.  Defaults to `https://circleci.com/api/v1.1/`.

* **vcs_type** is a value indicating the Version Control System type of the project builds being looked up in CircleCI's API.  Valid values are `github` and `bitbucket`.

* **owner** is the username or organization that owns the project in the VCS.

* **ttl** is a time duration used as the default *ttl* when no *ttl* is specified with the login request.  This value is larger than *max_ttl*, it will be capped at *max_ttl*.

* **max_ttl** is a time duration that sets the largest *ttl* that a new token can be assigned in this plugin.

* **attempt_cache_expiry** is the time duration that the backend caches login attempts.  Defaults to `18000s` (5 hours).

The plugin caches login attempts for and an approximate duration of **attempts_cache_time**.
The longer this duration, the greater the memory requirement will be.  However, in order
to prevent replay attacks, the plugin must cache the attempt for a duration that exceeds
the maximum CircleCI build duration possible.

## Policy Mapping

By default, this plugin doesn't associate any policies with the tokens that it creates,
except for the **default** policy.  To specify which policies are associated, a mapping
of project to policies must be provided by sending a POST request to the
`auth/circleci/map/projects/:project` endpoint.

Request parameters:

* **project** is the name of the CircleCI project being mapped to specific policies.

* **value** is a comma separated list of policy names to map to the specified project.

## Authentication

To create a token with this plugin, send a POST request to the `auth/circleci/login`
endpoint.

![Login Request](./docs/LoginRequest.png)

Request parameters:

* **project** is the name of the CircleCI project.

* **build_num** is the number of the running CircleCI build.

* **vcs_revision** is the VCS revision (commit hash) of the code which triggered the current build.

Upon receiving this request, the Vault server will serve it to this plugin,
which in turn will initiate a GET request to `:base_url/project/:vcs_type/:owner/:project/:build_num?circle-token=:circleci_token` 

![Verification Request](./docs/VerificationRequest.png)

Request parameters:

* **vcs_type** is the type of Version Control System for this project as specified in the plugin's configuration.

* **owner** is the user or organization that owns this project as specified in the plugin's configuration.

* **project** is the name of the CircleCI project as specified in the Login request.

* **build_num** is the number of the current CircleCI build as specified in the Login request.

* **circleci_token** is a CircleCI personal API token that gives access to the `project/` endpoint as specified in the plugin's configuration.

CircleCI will respond to this request with a response describing the specified build.

```json
{
  "vcs_url" : ...,
  "build_url" : ...,
  "build_num" : :build_num,
  "branch" : ...,
  "vcs_revision" : ":hash",
  "committer_name" : ...,
  "committer_email" : ...,
  "subject" : ...,
  "body" : ..., 
  "why" : ...,
  "dont_build" : ...,
  "queued_at" : ...,
  "start_time" : ...,
  "stop_time" : ...,
  "build_time_millis" : ...,
  "username" : ":user",
  "reponame" : ":project",
  "lifecycle" : "running",
  "outcome" : ...,
  "status" : "running",
  "retry_of" : ...,
  "steps" : [ ... ],
  ...
}
```

The plugin will verify that both the **lifecycle** key in the response is set
to `running`, and that the **vcs_revision** key is set to the same value as the
**vcs_revision** parameter that was specified in the Login request.  If these
conditions are met, a new token will be created with all of the mapped policies
associated to it, in addition to the **default** policy.

Finally the plugin will send a response to the initial Login request that
resembles the following:

```json
{
    "auth": {
        "client_token": "00000000-1111-2222-3333-444444444444",
        "policies": [
            "default",
            "mapped-policy"
        ],
        "metadata": {
        },
        "lease_duration": 300,
        "renewable": false
    }
}
```
