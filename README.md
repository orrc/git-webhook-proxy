git-webhook-proxy
=================

Acts as a proxy for incoming webhooks between your Git hosting provider and your continuous integration server.

When a Git commit webhook is received, the repository in question will be mirrored locally (or updated, if it already exists), and then the webhook will be passed on to your CI server, where it can start a build, using the up-to-date local mirror.

Problem
-------
If you run a CI server which has multiple jobs using the same Git repository, you may find a lot of time is wasted with cloning the repository.  Especially if new jobs are created often (e.g. on-the-fly), or workspaces get cleaned out often.

The Git plugin for Jenkins allows the usage of a "reference repo" — essentially using the `git clone --reference` behaviour — which lets multiple jobs access Git objects via a single repository on local disk, saving on storage and network access.

However, this first requires that local reference repository is cloned, and subsequently kept up-to-date.
Ensuring the repository is always up-to-date and ready for your CI server to build from, even when webhooks arrive seconds after commits are pushed, can be difficult.

Solution
--------
This server generates such local repositories on demand — using `git clone --mirror` — transparent to the CI server, whenever a webhook is received.  When further hooks are received, the local mirror is updated from the remote (using `git remote update`).

This process blocks until the repository has been cloned or updated, and then the webhook is forwarded to the CI server, whose response will be returned to the original initiator of the webhook.

Download
--------
Binaries for select platforms are available from the [releases page](https://github.com/orrc/git-webhook-proxy/releases).

Building
--------
Once you have the [Go development tools](http://golang.org/doc/install) installed, you should be able to run:  
`go get github.com/orrc/git-webhook-proxy`

Configuration
-------------
### git-webhook-proxy
You should run git-webhook-proxy on the same machine as your CI server, or on a machine that has access to the same disk space as your CI server.

Running `git-webhook-proxy --help` will display the command line options.

You must explicitly specify the address(es) to listen on, e.g. `--listen 127.0.0.1:8000` for HTTP, or `--tls-listen :8443` for TLS.
To accept TLS connections, you must provide your TLS certificate in PEM format (if intermediate certificates are required, append them to your certificate file) and private key.

The interface and port given should be reachable from the public internet, if you want to receive webhooks from services like GitHub.

The directory to which Git repositories will be mirrored is set by the `--mirror-path` flag.

The URL to which incoming webhook requests should be forwarded, is configured with `--remote`.

### Webhooks
Set up webhooks as normal at your Git hosting provider.

e.g. For GitHub, use the Jenkins webhook type.

### Jenkins
For each job that should share a local repository, the job should be configured as normal, i.e. using the remote Git URL to clone from.

In addition, you should choose "Additional Behaviours > Add > Advanced clone behaviours" which will reveal some more options.
Set the "Path of the reference repo to use during clone" to the value of `--mirror-path` plus the repository directory name.

The repository directory name is determined from the Git clone URL, and has the form `<host>/<path>.git`.
e.g. The Git URL `git@github.com:example/code` has the name `github.com/example/code.git`.

So, with the `--mirror-path` of `/opt/git/mirrors`, the full path to enter into Jenkins would be `/opt/git/mirrors/github.com/example/code.git`.

Limitations
-----------
Currently, only requests with the exact path of `/git/notifyCommit?url=<repo_uri>` or `/github-webhook/` are processed.

These are the standard URL formats used by the Git and GitHub plugins for Jenkins respectively, making this tool a drop-in replacement if you use Jenkins.

In the future, this will be more flexible.

Licence
-------
    The MIT License (MIT)

    Copyright (c) 2014 Christopher Orr
