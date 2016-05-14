package main

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"text/template"
)

import "os"

var pod_uuid = "26E56A04-F590-11E4-A66F-D7B3DD9DA696"
var pod_manifest = `{
    "acVersion": "0.5.1",
    "acKind": "PodManifest",
    "apps": [
        {
            "name": "reduce-worker",
            "image": {
                "name": "example.com/reduce-worker",
                "id": "sha512-8d3fffddf79e9a232ffd19f9ccaa4d6b37a6a243dbe0f23137b108a043d9da13121a9b505c804956b22e93c7f93969f4a7ba8ddea45bf4aab0bebc8f814e0990"
            },
            "annotations": [{"name": "foo", "value": "baz"}]
        },
        {
            "name": "backup",
            "image": {
                "name": "example.com/worker-backup",
                "id": "sha512-d603c29df0214c9b6681ed591871d40cc4bfabf9914383ce95ada0f2333defa7e97e21ca347e1d8dfde0b3edfe703688729cd25cec895a9a5b5c856da2f031fe",
                "labels": [{"name": "version", "value": "latest"}]
            }
        }
    ],
    "annotations": [{"name": "ip-address", "value": "10.1.2.3"}]
}`
var image_manifest = `{
    "acKind": "ImageManifest",
    "acVersion": "0.5.1",
    "name": "example.com/reduce-worker",
    "labels": [
        {"name": "version", "value": "1.0.0"},
        {"name": "arch", "value": "amd64"},
        {"name": "os", "value": "linux"}
    ],
    "app": {
        "exec": ["/usr/bin/reduce-worker", "--quiet"],
        "user": "100",
        "group": "300",
        "eventHandlers": [
            {"name": "pre-start", "exec": ["/usr/bin/data-downloader"]},
            {"name": "post-stop", "exec": ["/usr/bin/deregister-worker", "--verbose"]}
        ],
        "environment": [
            {"name": "REDUCE_WORKER_DEBUG", "value": "true"}
        ],
        "isolators": [
            {"name": "resource/cpu", "value": {"limit": "20"}},
            {"name": "resource/memory", "value": {"limit": "1G"}},
            {"name": "os/linux/capabilities-revoke-set", "value": {"set": ["CAP_NET_BIND_SERVICE", "CAP_SYS_ADMIN"]}}
        ],
        "ports": [{"name": "health", "port": 4000, "protocol": "tcp", "socketActivated": true}]
    },
    "dependencies": [
        {
            "app": "example.com/reduce-worker-base",
            "imageID": "sha512-7fa909434c9683e9db38a56a35f83e838a2df25b9c6c13dd3d9ce25ec6463b3cac338c94289528cf5f8b9e70e9bcdf59246fe05e7b91489ee5fb9cb0c7db92cd",
            "labels": [
                {"name": "os", "value": "linux"},
                {"name": "env", "value": "canary"}
            ]
        }
    ],
    "pathWhitelist": ["/etc/ca/example.com/crt", "/usr/bin/map-reduce-worker", "/opt/libs/reduce-toolkit.so", "/etc/reduce-worker.conf", "/etc/systemd/system/"],
    "annotations": [
        {"name": "authors", "value": "Carly Container <carly@example.com>, Nat Network <[nat@example.com](mailto:nat@example.com)>"},
        {"name": "created", "value": "2014-10-27T19:32:27.67021798Z"},
        {"name": "documentation", "value": "https://example.com/docs"},
        {"name": "homepage", "value": "https://example.com"}
    ]
}`

func serveMetadata(w http.ResponseWriter, r *http.Request) {
	if hdr, ok := r.Header["Metadata-Flavor"]; !ok || len(hdr) != 1 || hdr[0] != "AppContainer" {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Metadata-Flavor header missing or invalid"))
		return
	}

	switch r.URL.Path {
	case "/acMetadata/v1/pod/uuid":
		w.Write([]byte(pod_uuid))
	case "/acMetadata/v1/pod/manifest":
		w.Write([]byte(pod_manifest))
	case "/acMetadata/v1/pod/annotations":
		w.Write([]byte(`[{"name": "ip-address", "value": "10.1.2.3"}]`))
	case "/acMetadata/v1/apps/reduce-worker/image/id":
		w.Write([]byte("sha512-8d3fffddf79e9a232ffd19f9ccaa4d6b37a6a243dbe0f23137b108a043d9da13121a9b505c804956b22e93c7f93969f4a7ba8ddea45bf4aab0bebc8f814e0990"))
	case "/acMetadata/v1/apps/reduce-worker/image/manifest":
		w.Write([]byte(image_manifest))
	case "/acMetadata/v1/apps/reduce-worker/annotations":
		w.Write([]byte(`[
        {"name": "foo", "value": "baz"},
        {"name": "authors", "value": "Carly Container <carly@example.com>, Nat Network <[nat@example.com](mailto:nat@example.com)>"},
        {"name": "created", "value": "2014-10-27T19:32:27.67021798Z"},
        {"name": "documentation", "value": "https://example.com/docs"},
        {"name": "homepage", "value": "https://example.com"}]`))
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

var mds = httptest.NewServer(http.HandlerFunc(serveMetadata))

func init() {
	if err := os.Setenv("AC_METADATA_URL", mds.URL); err != nil {
		panic(err)
	}

	if err := os.Setenv("AC_APP_NAME", "reduce-worker"); err != nil {
		panic(err)
	}
}

func TestMDCApi(t *testing.T) {
	mdc := NewMDClient()

	if mdc.ACAppName != "reduce-worker" {
		t.Error("Invalid ACAppName", mdc.ACAppName)
	}

	if uuid := mdc.UUID(); uuid != pod_uuid {
		t.Error("Invalid UUID:", uuid)
	}

	if mdc.PodManifestJSON() != pod_manifest {
		t.Error("Invalid pod manifest")
	}

	val, found := mdc.PodAnnotations().Get("ip-address")
	if !found {
		t.Error("ip-address pod annotation not found")
	}
	if val != "10.1.2.3" {
		t.Error("Invalid annotation value:", val)
	}

	_, found = mdc.PodAnnotations().Get("whatever")
	if found {
		t.Error("whatever pod annotation found")
	}

	if id := mdc.AppImageID(); id != "sha512-8d3fffddf79e9a232ffd19f9ccaa4d6b37a6a243dbe0f23137b108a043d9da13121a9b505c804956b22e93c7f93969f4a7ba8ddea45bf4aab0bebc8f814e0990" {
		t.Error("Invalid app image ID:", id)
	}

	if mdc.AppImageManifestJSON() != image_manifest {
		t.Error("Invalid pod manifest")
	}

	val, found = mdc.AppAnnotations().Get("foo")
	if !found {
		t.Error("App annotation foo not found")
	}
	if val != "baz" {
		t.Error("Invalid annotation value:", val)
	}

	_, found = mdc.AppAnnotations().Get("bar")
	if found {
		t.Error("Nonexistent app annotation bar found")
	}
}

const templateText = `
I am a {{.ACAppName}}, {{.UUID}}, running {{.AppImageID}}
My IP address is {{.PodAnnotation "ip-address"}} {{.PodAnnotationOr "ip-address" "0.0.0.0"}}
My whatever is "{{.PodAnnotation "whatever"}}" "{{.PodAnnotationOr "whatever" "whatevs"}}"
My "foo" is "{{.AppAnnotation "foo"}}" "{{.AppAnnotationOr "foo" "quux"}}"
My "bar" is "{{.AppAnnotation "bar"}}" "{{.AppAnnotationOr "bar" "quux"}}"
`

const templateExpected = `
I am a reduce-worker, 26E56A04-F590-11E4-A66F-D7B3DD9DA696, running sha512-8d3fffddf79e9a232ffd19f9ccaa4d6b37a6a243dbe0f23137b108a043d9da13121a9b505c804956b22e93c7f93969f4a7ba8ddea45bf4aab0bebc8f814e0990
My IP address is 10.1.2.3 10.1.2.3
My whatever is "" "whatevs"
My "foo" is "baz" "baz"
My "bar" is "" "quux"
`

func TestTemplateRendering(t *testing.T) {
	mdc := NewMDClient()
	out := &bytes.Buffer{}

	tmpl := template.Must(template.New("appc-metadata-client").Parse(templateText))
	if err := tmpl.Execute(out, mdc); err != nil {
		t.Error("Error rendering template:", err)
	} else if actual := out.String(); actual != templateExpected {
		t.Error("Rendered template: got %#v, but expected %#v", actual, templateExpected)
	}
}
