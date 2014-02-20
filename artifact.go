package main

import (
	"net/url"
)


type AnArtifact struct {
	ArtifactUrl *url.URL
	ManifestUrl *url.URL
	manifest *Manifest
}

type Manifest struct {
	ThisVersion string
	ThisVersionHash string
	Patches ManifestPatches
}


type ManifestPatch struct {
	PatchIndex int
	PatchHash string
	FromVersion string
	FromVersionHash string
}

type ManifestPatches []*ManifestPatch


