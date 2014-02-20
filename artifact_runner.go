package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	//"net/url"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
)

type ArtifactRunner struct {

	artifact *AnArtifact
	config   *Configuration
	targetPatch *ManifestPatch

	// State 
	UseStorage  bool
	UseXdelta   bool
	UseDownload bool

	HasManifest  bool
	HasPatchPlan bool
	HasWorkspace bool

	HasPatchDownload bool
	HasPatchVerified bool
	HasPatchedArtifact bool
	HasPatchedArtifactVerified bool
	HasArtifact bool
}


func NewRunner (art *AnArtifact, config *Configuration) *ArtifactRunner {
	return &ArtifactRunner{
		artifact:    art,
		config:      config,
		targetPatch: nil,

		UseStorage:   false,
		UseXdelta:    false,
		UseDownload:  false,
		HasManifest:  false,
		HasPatchPlan: false,
		HasWorkspace: false,
		HasPatchDownload: false,
		HasPatchedArtifact: false,
		HasPatchVerified: false,
		HasPatchedArtifactVerified: false,
		HasArtifact: false}
}

func (ar *ArtifactRunner) printStateFlags () {
	fmt.Printf("------------------------------ \n")
	fmt.Printf("UseStorage ................... %t\n", ar.UseStorage)
	fmt.Printf("UseXdelta .................... %t\n", ar.UseXdelta)
	fmt.Printf("UseDownload .................. %t\n", ar.UseDownload)
	fmt.Printf("HasManifest .................. %t\n", ar.HasManifest)
	fmt.Printf("HasPatchPlan ................. %t\n", ar.HasPatchPlan)
	fmt.Printf("HasWorkspace ................. %t\n", ar.HasWorkspace)
	fmt.Printf("HasPatchDownload ............. %t\n", ar.HasPatchDownload)
	fmt.Printf("HasPatchVerified ............. %t\n", ar.HasPatchVerified)
	fmt.Printf("HasPatchedArtifact ........... %t\n", ar.HasPatchedArtifact)
	fmt.Printf("HasPatchedArtifactVerified ... %t\n", ar.HasPatchedArtifactVerified)
	fmt.Printf("HasArtifact .................. %t\n", ar.HasArtifact)
	fmt.Printf("------------------------------ \n")
}


func (ar *ArtifactRunner) verifySystemTools () {

	// check to see if xdelta3 exists where we expect to find it
	info, err := os.Stat(ar.config.XdeltaPath)
	if err != nil {
		return
	}

	// check to see if xdelta3 is executable
	if info.Mode() & 0111 == 0 {
		return
	}

	ar.UseXdelta = true
}


func (ar *ArtifactRunner) verifySystemPaths () bool {

	// check to see if the file/directory exists
	info, err := os.Stat(ar.config.ArtifactorPath)
	if err != nil {
		return ar.UseStorage
	}


	// verify that we're looking at a directory an not a regular file
	if info.Mode().IsRegular() {
		return ar.UseStorage
	}

	// faster than doing it through perm-checking
	file, err := ioutil.TempFile(ar.config.ArtifactorPath, "art")
	if err != nil {
		return ar.UseStorage
	}
	os.Remove(file.Name())

	ar.UseStorage = true
	return ar.UseStorage
}

func (ar *ArtifactRunner) verifyDownloadPath () bool {

	_, err := os.Stat(ar.config.DownloadPath)
	if os.IsNotExist(err) {
		ar.UseDownload = true
		return ar.UseDownload
	}

	log.Fatal(err)
	os.Exit(1)
	return ar.UseDownload
}


func (ar *ArtifactRunner) getManifest () bool {

	response, err := http.Get(ar.artifact.ManifestUrl.String())
	if err != nil {
		return ar.HasManifest
	}
	defer response.Body.Close()

	if response.StatusCode != 200 {
		return ar.HasManifest
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return ar.HasManifest
	}

	// todo: should validate the manifest ...
	json.Unmarshal(body, &ar.artifact.manifest)
	ar.HasManifest = true

	return ar.HasManifest
}


func (ar *ArtifactRunner) verifyPatchPlan () bool {

	if ar.HasManifest == false {
		return ar.HasPatchPlan
	}

	if len(ar.artifact.manifest.Patches) <= 0 {
		return ar.HasPatchPlan
	}

	for _, patch := range ar.artifact.manifest.Patches {
		_, err := os.Stat(ar.artifactPathForPatch(patch))
		if err == nil {
			ar.targetPatch  = patch
			ar.HasPatchPlan = true
			break
		}
	}

	return ar.HasPatchPlan
}


func (ar *ArtifactRunner) createArtifactWorkspace () bool {

	err := os.Mkdir(ar.workspacePath(), 0755)
	if err == nil {
		ar.HasWorkspace = true
	}

	return ar.HasWorkspace
}


func (ar *ArtifactRunner) downloadPatch () bool {

	patch_url, _ := url.Parse(
		fmt.Sprintf("%s.%d.xd3", ar.artifact.ArtifactUrl.String(), ar.targetPatch.PatchIndex))

	// todo: patchPathForPatch seems sloppy
	if download(patch_url.String(), ar.patchPathForPatch(ar.targetPatch)) {
		ar.HasPatchDownload = true
		ar.verifyPatch()
	}

	return ar.HasPatchDownload
}

func (ar *ArtifactRunner) verifyPatch () bool {
	if calculateArtifactSHA1(ar.patchPathForPatch(ar.targetPatch)) == ar.targetPatch.PatchHash {
		ar.HasPatchVerified = true
	}
	return ar.HasPatchVerified
}

func (ar *ArtifactRunner) verifyPatchedArtifact () bool {

	thisArtifactPath := filepath.Join(ar.workspacePath(), "artifact")

	if calculateArtifactSHA1(thisArtifactPath) == ar.artifact.manifest.ThisVersionHash {
		ar.HasPatchedArtifactVerified = true
		ar.HasArtifact = true
	}

	return ar.HasPatchedArtifactVerified
}

func download (target_url, dest string) bool {

	response, err := http.Get(target_url)
	if err != nil {
		return false
	}

	defer response.Body.Close()

	out, err := os.Create(dest)
	if err != nil {
		return false
	}

	// todo: should handle redirects
	if response.StatusCode != 200 {
		return false
	}

	io.Copy(out, response.Body)

	return true
}


func (ar *ArtifactRunner) buildArtifactFromPatch () bool {

	pastArtifactPath := ar.artifactPathForPatch(ar.targetPatch)
	thisArtifactPath := filepath.Join(ar.workspacePath(), "artifact")
	patchPath        := ar.patchPathForPatch(ar.targetPatch)

	cmd     := exec.Command("/usr/bin/xdelta3", "-d", "-s", pastArtifactPath, patchPath, thisArtifactPath)
	cmd.Env  = append(os.Environ(), "GZIP=-9")
	err     := cmd.Run()

	if err != nil {
		fmt.Printf("Foo: %s\n", err)
		return ar.HasPatchedArtifact
	}

	ar.HasPatchedArtifact = true
	ar.verifyPatchedArtifact()

	return ar.HasPatchedArtifact
}

func (ar *ArtifactRunner) Run () {
	ar.printStateFlags()


	// verify that we can write to the output file; if not, it's a haulting error
	log.Printf("Verify Download Path\n")
	ar.verifyDownloadPath()

	// verify we have a place to store/cache files and that xdelta is present. if
	// it's not, well that's not the end of the world but we're limited to being
	// a curl clone
	log.Printf("Verify System Paths and Tools\n")
	ar.verifySystemPaths()
	ar.verifySystemTools()

	// if we can, do it
	if ar.UseStorage && ar.UseXdelta {
	        log.Printf("Get Manifest and Verify Patch Plan\n")
		ar.getManifest()
		ar.verifyPatchPlan()

		if ar.HasManifest {
			log.Printf("Create Artifact Workspace\n")
			ar.createArtifactWorkspace()

			if ar.HasPatchPlan && ar.HasWorkspace {
				log.Printf("Download Patch\n")
				ar.downloadPatch()
			}

			if ar.HasPatchDownload && ar.HasPatchVerified {
				log.Printf("Build Artifact From Patch\n")
				ar.buildArtifactFromPatch()
			} else {
				log.Printf("Plan B: Artifact / Optimize For Next Time\n")
				// setup for next time
				// todo: verify?
				dest := filepath.Join(ar.workspacePath(), "artifact")
				if download(ar.artifact.ArtifactUrl.String(), dest) {
					ar.HasArtifact = true
				}
			}

			// if it's all good, copy our artifact to the desired directory
			if ar.HasArtifact {
				log.Printf("Optimize Success\n")
				src := filepath.Join(ar.workspacePath(), "artifact")
				CopyFile(src, ar.config.DownloadPath)
			}
		}
	}

	// last ditch plan -- simply download the artifact directly to the desired direcotry
	if ! ar.HasArtifact {
		log.Printf("Plan D: When All Else Fails - Just Do It\n")
		log.Printf("Plan D: %s, %s\n", ar.artifact.ArtifactUrl.String(), ar.config.DownloadPath)
		if ! download(ar.artifact.ArtifactUrl.String(), ar.config.DownloadPath) {
			log.Printf("No Luck and End of the Line\n")
		}
	}

	ar.printStateFlags()

}

