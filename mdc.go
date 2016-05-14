package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/appc/spec/schema"
	"github.com/appc/spec/schema/types"
)

func usage(rv int) {
	fmt.Fprintln(os.Stderr, strings.Replace(`Usage:
    $0 uuid                          -- show pod UUID
    $0 annotation NAME [DEFAULT]     -- show pod's annotation
    $0 manifest                      -- show pod manifest JSON
    $0 image-id                      -- show current app image ID
    $0 image-manifest                -- show current app image manifest JSON
    $0 app-annotation NAME [DEFAULT] -- show current app's annotation
    $0 render PATH|-                 -- render template file or stdin to stdout
    $0 expand TEMPLATE-STRING        -- render template string to stdout`,
		"$0", filepath.Base(os.Args[0]), -1))
	os.Exit(rv)
	panic("CAN'T HAPPEN")
}

type MDClient struct {
	ACMetadataURL, ACAppName              string
	uuid, appImageID                      string
	podAnnotations, appAnnotations        types.Annotations
	podManifestJSON, appImageManifestJSON []byte
	podManifest                           *schema.PodManifest
	appImageManifest                      *schema.ImageManifest
}

func NewMDClient() *MDClient {
	rv := &MDClient{
		ACMetadataURL: os.Getenv("AC_METADATA_URL"),
		ACAppName:     os.Getenv("AC_APP_NAME"),
	}

	if rv.ACMetadataURL == "" {
		fmt.Fprintln(os.Stderr, "FATAL: No AC_METADATA_URL environment variable")
		os.Exit(1)
	}

	if rv.ACAppName == "" {
		fmt.Fprintln(os.Stderr, "FATAL: No AC_APP_NAME environment variable")
		os.Exit(1)
	}

	return rv
}

func (mdc *MDClient) Get(path string) []byte {
	req, err := http.NewRequest("GET", mdc.ACMetadataURL+"/acMetadata/v1/"+path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Metadata-Flavor", "AppContainer")

	if resp, err := (&http.Client{}).Do(req); err != nil {
		panic(err)
	} else if resp.StatusCode == 404 {
		return nil
	} else if resp.StatusCode != 200 {
		fmt.Fprintln(os.Stderr, "\nERROR: GET", path)
		resp.Write(os.Stderr)
		os.Exit(1)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		panic(err)
	} else {
		return body
	}

	panic("CAN'T HAPPEN")
}

func (mdc *MDClient) UUID() string {
	if mdc.uuid == "" {
		mdc.uuid = strings.TrimSpace(string(mdc.Get("pod/uuid")))
	}
	return mdc.uuid
}

func (mdc *MDClient) PodAnnotations() types.Annotations {
	if mdc.podAnnotations == nil {
		if err := json.Unmarshal(mdc.Get("pod/annotations"), &mdc.podAnnotations); err != nil {
			panic(err)
		}
	}
	return mdc.podAnnotations
}

func (mdc *MDClient) PodAnnotation(name string) string {
	v, _ := mdc.PodAnnotations().Get(name)
	return v
}

func (mdc *MDClient) HasPodAnnotation(name string) bool {
	_, v := mdc.PodAnnotations().Get(name)
	return v
}

func (mdc *MDClient) MustPodAnnotation(name string) string {
	v, found := mdc.PodAnnotations().Get(name)
	if !found {
		panic("Annotation not found: " + name)
	}
	return v
}

func (mdc *MDClient) PodAnnotationOr(name, defaultValue string) string {
	v, found := mdc.PodAnnotations().Get(name)
	if !found {
		return defaultValue
	}
	return v
}

func (mdc *MDClient) podManifestBytes() []byte {
	if mdc.podManifestJSON == nil {
		mdc.podManifestJSON = mdc.Get("pod/manifest")
	}
	return mdc.podManifestJSON
}

func (mdc *MDClient) PodManifestJSON() string {
	return string(mdc.podManifestBytes())
}

func (mdc *MDClient) PodManifest() *schema.PodManifest {
	if mdc.podManifest == nil {
		mdc.podManifest = &schema.PodManifest{}
		if err := json.Unmarshal(mdc.podManifestBytes(), mdc.podManifest); err != nil {
			panic(err)
		}
	}
	return mdc.podManifest
}

func (mdc *MDClient) AppImageID() string {
	if mdc.appImageID == "" {
		mdc.appImageID = strings.TrimSpace(string(mdc.Get("apps/" + mdc.ACAppName + "/image/id")))
	}
	return mdc.appImageID
}

func (mdc *MDClient) appImageManifestBytes() []byte {
	if mdc.appImageManifestJSON == nil {
		mdc.appImageManifestJSON = mdc.Get("apps/" + mdc.ACAppName + "/image/manifest")
	}
	return mdc.appImageManifestJSON
}

func (mdc *MDClient) AppImageManifestJSON() string {
	return string(mdc.appImageManifestBytes())
}

func (mdc *MDClient) AppImageManifest() *schema.ImageManifest {
	if mdc.appImageManifest == nil {
		mdc.appImageManifest = &schema.ImageManifest{}
		if err := json.Unmarshal(mdc.appImageManifestBytes(), mdc.appImageManifest); err != nil {
			panic(err)
		}
	}
	return mdc.appImageManifest
}

func (mdc *MDClient) AppAnnotations() types.Annotations {
	if mdc.appAnnotations == nil {
		if err := json.Unmarshal(mdc.Get("apps/"+mdc.ACAppName+"/annotations"), &mdc.appAnnotations); err != nil {
			panic(err)
		}
	}
	return mdc.appAnnotations
}

func (mdc *MDClient) AppAnnotation(name string) string {
	v, _ := mdc.AppAnnotations().Get(name)
	return v
}

func (mdc *MDClient) HasAppAnnotation(name string) bool {
	_, v := mdc.AppAnnotations().Get(name)
	return v
}

func (mdc *MDClient) MustAppAnnotation(name string) string {
	v, found := mdc.AppAnnotations().Get(name)
	if !found {
		panic("Annotation not found: " + name)
	}
	return v
}

func (mdc *MDClient) AppAnnotationOr(name, defaultValue string) string {
	v, found := mdc.AppAnnotations().Get(name)
	if !found {
		return defaultValue
	}
	return v
}

func main() {
	mdc := NewMDClient()

	if len(os.Args) < 2 {
		usage(0)
	}

	switch os.Args[1] {
	case "help", "--help", "-help", "-h":
		usage(0)
	case "uuid":
		fmt.Println(mdc.UUID())
	case "annotation":
		if len(os.Args) < 3 {
			usage(1)
		}
		if ann, found := mdc.PodAnnotations().Get(os.Args[2]); found {
			fmt.Println(ann)
		} else if len(os.Args) > 3 {
			fmt.Println(os.Args[3])
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: No such annotation: %#v\n", os.Args[2])
			os.Exit(1)
		}
	case "manifest":
		fmt.Println(mdc.PodManifestJSON())
	case "image-id":
		fmt.Println(mdc.AppImageID())
	case "image-manifest":
		fmt.Println(mdc.AppImageManifestJSON())
	case "app-annotation":
		if len(os.Args) < 3 {
			usage(1)
		}
		if ann, found := mdc.AppAnnotations().Get(os.Args[2]); found {
			fmt.Println(ann)
		} else if len(os.Args) > 3 {
			fmt.Println(os.Args[3])
		} else {
			fmt.Fprintf(os.Stderr, "ERROR: No such annotation: %#v\n", os.Args[2])
			os.Exit(1)
		}
	case "render":
		if len(os.Args) < 3 {
			usage(1)
		}
		path := os.Args[2]
		if path == "-" {
			path = "/dev/stdin"
		}
		if err := template.Must(template.ParseFiles(path)).Execute(os.Stdout, mdc); err != nil {
			panic(err)
		}
	case "expand":
		if len(os.Args) < 3 {
			usage(1)
		}
		if err := template.Must(template.New("").Parse(os.Args[2])).Execute(os.Stdout, mdc); err != nil {
			panic(err)
		}
	default:
		usage(1)
	}
}
