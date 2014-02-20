package main

import (
	//"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"
)


func CreateHttpServerForTesting (t *testing.T) {

	http.Handle("/", http.FileServer(http.Dir("./test/data")))
	err := http.ListenAndServe("localhost:12345", nil)
        if err != nil {
	        t.Error("ListenAndServe: " + err.Error())
        }
}


func TestRunnerOnArtifactV1 (t *testing.T) {

	go CreateHttpServerForTesting(t)

	// create some temp directories
	temp_path, _       := ioutil.TempDir("./test/", "tmp")
	artifactor_path, _ := ioutil.TempDir(temp_path, "artifact-")
	download_path,   _ := ioutil.TempDir(temp_path, "download-")
	download_path       = filepath.Join(download_path, "version.tgz")

	log.Printf("Temp Path:       %s\n", temp_path)
	log.Printf("Artifactor Path: %s\n", artifactor_path)
	log.Printf("Download Path:   %s\n", download_path)

	artifact_url, _ := url.Parse("http://localhost:12345/v1.tgz")
	manifest_url, _ := url.Parse("http://localhost:12345/v1.tgz.manifest")

	artifact := AnArtifact{ ArtifactUrl: artifact_url, ManifestUrl: manifest_url }
	config   := Configuration{
		ArtifactorPath: artifactor_path,
		DownloadPath:   download_path,
		XdeltaPath:     "/usr/local/bin/xdelta3"}

	runner := NewRunner(&artifact, &config)
	runner.Run()

	// Now for the Actual Test!

	// 1. download_path should exist

	_, err := os.Stat(download_path)
	if os.IsNotExist(err) {
		t.Error("Should have downloaded an artifact!")
		return
	}

	// 2. should have a manifest
	if (runner.artifact.manifest == nil) {
		t.Error("Should have a manifest!")
		return
	}

	// 2. should have a artifact for future use
	_, err = os.Stat(filepath.Join(runner.workspacePath(), "artifact"))
	if os.IsNotExist(err) {
		t.Error("Should have an artifact ready to go!")
		return
	}




}

