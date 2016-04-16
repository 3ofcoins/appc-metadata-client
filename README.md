App Container Metadata Client
=============================

This is a simple command line client for
[App Container](https://github.com/appc/spec)
[Metadata Service](https://github.com/appc/spec/blob/master/SPEC.md#app-container-metadata-service).

It is intended to be included in App Container Images to easily render
configuration files based on data available from the metadata service
(in particular, from annotations).

Building
--------

Best way is to use BSD Make (GNU Make won't work; I'll be happy to
accept a pull request that won't break BSD Make): type `make` to build
the `mdc` binary.

You can also use `go build`, if you specify `GOPATH` manually:

    env GOPATH=`pwd`/vendor go build -o mdc

If you prefer to just use your regular `GOPATH` with whatever version
of appc/spec that lives there, go ahead, but please note that it is
not supported and if anything breaks, try with vendor GOPATH before
filing a report:

    go get
    go build -o mdc

Usage
-----

Inside the appc pod, when `AC_METADATA_URL` and `AC_APP_NAME`
environment variables point to a running metadata service, you can use
the following commands to access the metadata:

    mdc uuid                    -- show pod UUID
    mdc annotation NAME         -- show pod's annotation
    mdc manifest                -- show pod manifest JSON
    mdc image-id                -- show current app image ID
    mdc image-manifest          -- show current app image manifest JSON
    mdc app-annotation NAME     -- show current app's annotation
    mdc render PATH|-           -- render template file or stdin to stdout
    mdc expand TEMPLATE-STRING  -- render template string to stdout

Template Rendering
------------------

You can use mdc to render templates that include values from the
metadata service. The `render` command will read a template from file
(of from standard input if `-` is given), and `expand` command will
render template from command line argument.

The templates use
[Go `text/template` syntax](https://golang.org/pkg/text/template/). The
following methods are supported:


 - `{{.ACMetadataURL}}`, `{{.ACAppName}}` – values of environment
   variables `AC_METADATA_URL` and `AC_APP_NAME`
 - `{{.UUID}}` – pod's UUID
 - `{{.PodAnnotation "name"}}` – pod's annotation value, empty string if does not exist
 - `{{.PodAnnotationOr "name" "default"}}` – pod's annotation value, "default" if does not exist
 - `{{.MustPodAnnotation "name"}}` – pod's annotation value, panics if does not exist
 - `{{.HasPodAnnotation "name"}}` – true if pod has an annotation of that name
 - `{{.PodManifest}}` – [PodManifest](https://godoc.org/github.com/appc/spec/schema#PodManifest) object
 - `{{.AppImageID}}` – ID of current app's image
 - `{{.AppImageManifest}}` – [ImageManifest](https://godoc.org/github.com/appc/spec/schema#ImageManifest) object for current app's image
 - `{{.AppAnnotation "name"}}`,
   `{{.AppAnnotationOr "name" "default"}}`,
   `{{.MustAppAnnotation "name"}}`,
   `{{.HasAppAnnotation "name"}}`– same as `…PodAnnotation…`, but for app annotations

### Example template

    # Rendered for pod {{.UUID}}
    DATABASES = {
    'default': {
        'ENGINE': 'django.db.backends.postgresql_psycopg2',

        'NAME': "{{.PodAnnotation "postgresql/dbname"}}",
        'USER': "{{.PodAnnotation "postgresql/user"}}",
        'PASSWORD': "{{.PodAnnotation "postgresql/password"}}",
        'HOST': "{{.PodAnnotation "postgresql/host"}}",
        'PORT': "{{.PodAnnotation "postgresql/port"}}",

        # If you're using Postgres, we recommend turning on autocommit
        'OPTIONS': { 'autocommit': True, }
        }
    }

Testing [![Build Status](https://travis-ci.org/3ofcoins/appc-metadata-client.svg?branch=master)](https://travis-ci.org/3ofcoins/appc-metadata-client)
-------

With BSD Make, run `make test` to run the test suite; otherwise, run:

    env GOPATH=`pwd`/vendor go test

Tests are running automatically at [Travis CI](https://travis-ci.org/3ofcoins/appc-metadata-client)

TODO
----

 - [ ] Gracefully handle nonexistent annotations
 - [ ] Implement identity service
 - [ ] Improve the test suite
