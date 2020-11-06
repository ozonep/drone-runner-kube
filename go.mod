module github.com/ozonep/drone-runner-kube

go 1.15

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.15
	k8s.io/api => k8s.io/api v0.17.13
	k8s.io/apimachinery => k8s.io/apimachinery v0.17.13
)

require (
	github.com/99designs/basicauth-go v0.0.0-20160802081356-2a93ba0f464d
	github.com/bmatcuk/doublestar v1.3.3
	github.com/buildkite/yaml v2.1.0+incompatible
	github.com/coreos/go-semver v0.3.0
	github.com/dchest/uniuri v0.0.0-20200228104902-7aecb25e1fe5
	github.com/docker/distribution v2.7.1+incompatible
	github.com/docker/go-units v0.4.0
	github.com/ghodss/yaml v1.0.0
	github.com/golang/mock v1.4.4
	github.com/google/go-cmp v0.5.2
	github.com/gosimple/slug v1.9.0
	github.com/hashicorp/go-multierror v1.1.0
	github.com/joho/godotenv v1.3.0
	github.com/kelseyhightower/envconfig v1.4.0
	github.com/mattn/go-isatty v0.0.12
	github.com/natessilva/dag v0.0.0-20180124060714-7194b8dcc5c4
	github.com/ozonep/drone v0.0.0-20201106101057-9383c6ce9b43
	github.com/sirupsen/logrus v1.7.0
	golang.org/x/crypto v0.0.0-20201016220609-9e8e0b390897
	golang.org/x/sync v0.0.0-20201020160332-67f06af15bc9
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.17.13
	k8s.io/apimachinery v0.17.13
	k8s.io/client-go v0.17.13
	k8s.io/utils v0.0.0-20201104234853-8146046b121e // indirect
)
