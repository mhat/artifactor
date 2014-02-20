package main

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func (ar *ArtifactRunner) workspacePath () string {
	return filepath.Join(ar.config.ArtifactorPath, ar.artifact.manifest.ThisVersion)
}


func (ar *ArtifactRunner) artifactPathForPatch (patch *ManifestPatch) string {
	return filepath.Join(ar.config.ArtifactorPath, patch.FromVersion, "artifact")
}


// todo: seems sloppy
func (ar *ArtifactRunner) patchPathForPatch (patch *ManifestPatch) string {
	return filepath.Join(
		ar.config.ArtifactorPath,
		ar.artifact.manifest.ThisVersion,
		fmt.Sprintf("patch.%s.xd3", ar.targetPatch))
}


func CopyFile(src, dst string) bool {

	src_f, err := os.Open(src)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer src_f.Close()

	dst_f, err := os.Create(dst)
	if err != nil {
		fmt.Println(err)
		return false
	}
	defer dst_f.Close()

	io.Copy(src_f, dst_f)
	return true
}

func calculateArtifactSHA1(dest string) string {

	fh, err := os.Open(dest)
	if err != nil {
		fmt.Println("Wat: ", err)
		return ""
	}

	defer fh.Close()

	reader := bufio.NewReader(fh)
	hash   := sha1.New()
	io.Copy(hash, reader)

	return hex.EncodeToString(hash.Sum(nil))
}
