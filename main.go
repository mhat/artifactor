package main

import (
	"fmt"
	"net/url"
	"os"
	"flag"
)

var ArtifactorOpts ArtifactorOptsStruct
type ArtifactorOptsStruct struct {
	ArtifactUrl string
	PathToCache string
	Verbose     bool
}

func init() {
	flag.StringVar(&ArtifactorOpts.ArtifactUrl, "url",     "",                "Artifact URL, most likely Maven, Azure or S3.")
	flag.StringVar(&ArtifactorOpts.PathToCache, "cache",   "/tmp/artifactor", "Where to cache downloads.")
	flag.BoolVar(&ArtifactorOpts.Verbose,       "verbose", false, "Verbose")
}



type Configuration struct {
	ArtifactorPath string
	XdeltaPath     string
	DownloadPath   string
}


func main () {
	flag.Parse()
	artifact_url, e := url.Parse(ArtifactorOpts.ArtifactUrl)
	manifest_url, e := url.Parse(ArtifactorOpts.ArtifactUrl + ".manifest")
	if (e != nil) {
		fmt.Println("Error: ", e)
	}

	art    := AnArtifact{ ArtifactUrl: artifact_url, ManifestUrl: manifest_url }
	config := &Configuration{
		ArtifactorPath: "/tmp/artifactor",
		DownloadPath: "/tmp/xxx",
		XdeltaPath: "/usr/local/bin/xdelta3"}

	runner := NewRunner(&art, config)
	runner.Run()
	os.Exit(0)
}

