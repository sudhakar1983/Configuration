# Configuration



version=v1.0.0 && \
git tag $version && git push origin $version

go get github.com/sudhakar1983/Configuration@v1.0.0



go get github.com/sudhakar1983/Configuration/v2@v2.0.1


git tag -d $version

go mod graph | grep github.com/sudhakar1983/Configuration/v2

go mod why github.com/sudhakar1983/Configuration/v2