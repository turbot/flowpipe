Don't push this the devcontainer Docker image to GitHub yet.

Build this locally and it will work.

## Troubleshooting

https://stackoverflow.com/questions/74707530/docker-buildx-fails-to-show-result-in-image-list


```
# build both images
docker buildx build --platform linux/arm64,linux/amd64 .
# load just one platform
docker buildx build --load --platform linux/amd64 -t my-image-tag .
```