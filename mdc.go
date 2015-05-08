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

var acMetadataUrl = os.Getenv("AC_METADATA_URL")
var acAppName = os.Getenv("AC_APP_NAME")
var httpClient = &http.Client{}

func get(path string) []byte {
	req, err := http.NewRequest("GET", acMetadataUrl+"/acMetadata/v1/"+path, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Metadata-Flavor", "AppContainer")

	if resp, err := httpClient.Do(req); err != nil {
		panic(err)
	} else if resp.StatusCode != 200 {
		resp.Write(os.Stderr)
		os.Exit(1)
	} else if body, err := ioutil.ReadAll(resp.Body); err != nil {
		panic(err)
	} else {
		return body
	}

	panic("CAN'T HAPPEN")
}

func show(path string) {
	fmt.Println(string(get(path)))
}

func main() {
	if acMetadataUrl == "" {
		panic("AC_METADATA_URL environment variable unset!")
	}

	if acAppName == "" {
		panic("AC_APP_NAME environment variable unset!")
	}

	if len(os.Args) < 2 {
		usage(0)
	}

	switch os.Args[1] {
	case "help", "--help", "-help", "-h":
		usage(0)
	case "uuid":
		show("pod/uuid")
	case "annotation":
		if len(os.Args) < 3 {
			usage(1)
		}
		show("pod/annotations/" + os.Args[2])
	case "manifest":
		show("pod/manifest")
	case "image-id":
		show("apps/" + acAppName + "/image/id")
	case "image-manifest":
		show("apps/" + acAppName + "/image/manifest")
	case "app-annotation":
		if len(os.Args) < 3 {
			usage(1)
		}
		show("apps/" + acAppName + "/annotations/" + os.Args[2])
	default:
		usage(1)
	}
}
