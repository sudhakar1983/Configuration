# Configuration


git add .
git commit -m "version=v1.0.1 - checksum"

version=v1.0.1 && \
git tag $version && git push origin $version

go get github.com/sudhakar1983/ServerConfig/@v1.0.1



git tag -d $version

go mod graph | grep github.com/sudhakar1983/ServerConfig

go mod why github.com/sudhakar1983/ServerConfig