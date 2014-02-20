
## Artifactor

Artifactor is a utility to fetch artifacts from a repository. The main advantage over Curl is that Artifactor
speeds up downloads where possibly by using binary patches rather than downloading the entire artifact.

### Usage

    artifactor --url <path to artifact> --version <version> --service <service name>
    artifactor --url <path to artifact> --output <path and filename>

### Conventions

    https://s3.amazonaws.com/my-artifacts/foo/foo-service-1.0-abc123.tar.gz
    https://s3.amazonaws.com/my-artifacts/foo/foo-service-1.0-abc123.tar.gz.manifest
    https://s3.amazonaws.com/my-artifacts/foo/foo-service-1.0-abc123.tar.gz.0.xd3

Or more generally:

    {{url-to-artifact}}
    {{url-to-artifact}}.manifest
    {{url-to-artifact}}.{{patch-index}}.xd3

Artifactor relies on a simple json manifest and xdelta3. The manifest tells Artifactor what patchs are available 
for the current version. From there Artifactor compares what it has cached locally to what''s possible. If an 
upgrade is possible, it will download the patch and apply it. If a patch-upgrade path isn''t available then
Artifactor simply downloads the whole Artifact.

At some future point it might make sense to be able to apply patches in squence. E.g. Our goal is to return
Artifact E. We have Artifact B. We could potentially patch from B->C, C->D and then D->E. We haven''t done any
testing to determin how viable this is. For now it''ll simply be a possibility that we don''t use.

### Manifest

    {
        "this-version"      : "{{this_version}}"
      , "this-version-sha1" : "..."
      , "patches"           : [

          { 
              "from-version"      : "{{from_version}}"
            , "from-version-sha1" : "..."
            , "patch-sha1"        : "..."
          } 

          // { ... } patch-1
          // { ... } patch-2
      
        ]
    }


### Artifact Storage

    /var/cache/artifactor/{{version}}
        artifact.manifest
        patch.{{from-version}}.xd3
        artifact

By default we store the last 3 artifacts. 

### Thoughts 

1. If the .manifest exists we can populate /cache and all that based on the manifest contents
2. If it does not, we simply put the file in --output

What does main() Look like?

1. Parse Options and set Configuration
2. Is there an Manifest for this Artifact
3. Is the Manifest's Patch Plan Viable?

if Artifactor.VerifyPermissions

if Artifactor.VerifyDependencies

art = NewArtifact(...)

if art.HasManifest

if art.HasViablePatchPath

CreateArtifactWorkspace

if getBestPatch

if applyPatch

if verifyPatchedFile










