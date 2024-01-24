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
- [ ] Mark relese as "latest"

### Post release check & admin
- [ ] Test Linux install script
- [ ] Test Homebrew install
- [ ] Test Windows install
- [ ] Release branches merged to `main` (all three repos `flowpipe`, `flowpipe-go-sdk`, `pipe-fittings`)