# Vault CircleCI Authentication Plugin

This plugin allows a running CircleCI build to authenticate itself with a Vault
server.

## Architecture

The authentication is achieved by having the running CircleCI build provide
pieces of data that the Vault server can verify using the CircleCI API.

The authentication is started by sending a POST request to the `auth/circleci/login`
endpoint.

![Login Request](./docs/LoginRequest.png)

Request parameters:

* **vcs_type** is the type of version control system for this project.  CircleCI supports `github` and `bitbucket`.

* **user** is the user that owns the repository of the project.  For repositories owned by an organization, set this parameter to the organization name.

* **project** is the name of the project or repository in the VCS.

* **build_num** is the number of the running build.

* **hash** is the VCS revision of the code which the current build is using.

Upon receiving this request, the Vault server will serve it to this plugin,
which in turn will initiate a GET request to `https://circleci.com/api/v1.1/project/:vcs_type/:user/:project/:build_num?circle-token=:token` 

![Verification Request](./docs/VerificationRequest.png)

Request parameters:

* **vcs_type** is the type of version control system for this project as specified in the Login request.

* **user** is the user (or organization) that owns this project as specified in the Login request.

* **project** is the name of the project in the VCS as specified in the Login request.

* **build_num** is the number of the current build as specified in the Login request.

* **token** is a CircleCI personal API token that gives access to the `project/` endpoint.

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

The plugin will verify that both the **lifecycle** and **status** keys in the
response are set to `running`, and that the **vcs_revision** key is set to the
same value as the **hash** parameter that was specified in the Login request.
If these conditions are met, a new token will be created.  Then, the plugin
will check if any specific policies have been mapped to the specified project.
If so, those policies will be attached to the new token.

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
