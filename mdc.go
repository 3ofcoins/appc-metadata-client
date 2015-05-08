package main

import "fmt"
import "io/ioutil"
import "os"
import "net/http"
import "strings"

func usage(rv int) {
	fmt.Fprintln(os.Stderr, strings.Replace(`Usage:
    $0 uuid
    $0 annotation NAME
    $0 manifest
    $0 image-id
    $0 image-manifest
    $0 app-annotation NAME`, "$0", os.Args[0], -1))
	os.Exit(rv)
	panic("CAN'T HAPPEN")
}

type MDClient struct {
	ACMetadataURL, ACAppName         string
	uuid, appImageID                 string
	podAnnotations, appAnnotations   map[string]string
	podManifestJSON, appManifestJSON []byte
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

func (mdc *MDClient) Show(path string) {
	fmt.Println(string(mdc.Get(path)))
}

func (mdc *MDClient) UUID() string {
	if mdc.uuid == "" {
		mdc.uuid = strings.TrimSpace(string(mdc.Get("pod/uuid")))
	}
	return mdc.uuid
}

func (mdc *MDClient) PodAnnotation(name string) string {
	if mdc.podAnnotations == nil {
		mdc.podAnnotations = make(map[string]string)
	}

	if value, found := mdc.podAnnotations[name]; !found {
		value = string(mdc.Get("pod/annotations/" + name))
		mdc.podAnnotations[name] = value
		return value
	} else {
		return value
	}
}

func (mdc *MDClient) PodManifestJSON() string {
	if mdc.podManifestJSON == nil {
		mdc.podManifestJSON = mdc.Get("pod/manifest")
	}
	return string(mdc.podManifestJSON)
}

func (mdc *MDClient) AppImageID() string {
	if mdc.appImageID == "" {
		mdc.appImageID = strings.TrimSpace(string(mdc.Get("apps/" + mdc.ACAppName + "/image/id")))
	}
	return mdc.appImageID
}

func (mdc *MDClient) AppManifestJSON() string {
	if mdc.appManifestJSON == nil {
		mdc.appManifestJSON = mdc.Get("apps/" + mdc.ACAppName + "/image/manifest")
	}
	return string(mdc.appManifestJSON)
}

func (mdc *MDClient) AppAnnotation(name string) string {
	if mdc.appAnnotations == nil {
		mdc.appAnnotations = make(map[string]string)
	}

	if value, found := mdc.appAnnotations[name]; !found {
		value = string(mdc.Get("apps/" + mdc.ACAppName + "/annotations/" + name))
		mdc.appAnnotations[name] = value
		return value
	} else {
		return value
	}
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
		fmt.Println(mdc.PodAnnotation(os.Args[2]))
	case "manifest":
		fmt.Println(mdc.PodManifestJSON())
	case "image-id":
		fmt.Println(mdc.AppImageID())
	case "image-manifest":
		fmt.Println(mdc.AppManifestJSON())
	case "app-annotation":
		if len(os.Args) < 3 {
			usage(1)
		}
		fmt.Println(mdc.AppAnnotation(os.Args[2]))
	default:
		usage(1)
	}
}
