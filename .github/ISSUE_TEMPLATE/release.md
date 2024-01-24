---
name: Flowpipe Release
about: Flowpipe Release
title: "Flowpipe v<INSERT_VERSION_HERE>"
labels: release
---

#### Changelog

[Flowpipe v<INSERT_VERSION_HERE> Changelog](https://github.com/turbot/flowpipe/blob/v<INSERT_VERSION_HERE>/CHANGELOG.md)

## Checklist

### Pipe Fittings
- [ ] Pipe Fittings Changelog updated with correct version and date
- [ ] `pipe-fittings` tagged with correct final version (ensure you have a clean branch, otherwise the tag will be created on the wrong commit and difficult to revert)

### Flowpipe SDK Go
- [ ] `flowpipe-sdk-go` tagged with correct final version (ensure you have a clean branch, otherwise the tag will be created on the wrong commit and difficult to revert)

### Flowpipe

- [ ] Flowpipe Changelog updated and reviewed
- [ ] Raise Changelog update to `flowpipe.io`
- [ ] Update Flowipe dependency to `flowpipe-go-sdk` to use the relase tag
- [ ] Update Flowpipe dependency to `pipe-fittings` to use the relase tag
- [ ] Run release build. Do not tag `flowpipe repo``, the workflow will create the tag
- [ ] Update Changelog in the Release page (copy and paste from CHANGELOG.md)
- [ ] Test Linux install script
- [ ] Test Windows install
- [ ] Mark release as "latest" (workflow creates pre-release version)
- [ ] Merge PR in `@turbot/homebrew-tap` repo to update Turbot Homebrew Tap

### Post release check & admin
- [ ] Test Homebrew install
- [ ] Release branches merged to `main` (all three repos `flowpipe`, `flowpipe-go-sdk`, `pipe-fittings`)