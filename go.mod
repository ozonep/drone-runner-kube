module github.com/ozonep/drone-runner-kube

go 1.15

replace (
	github.com/googleapis/gnostic => github.com/googleapis/gnostic v0.4.0
	github.com/h2non/gock => gopkg.in/h2non/gock.v1 v1.0.14
	k8s.io/api => k8s.io/api v0.16.15
	k8s.io/apimachinery => k8s.io/apimachinery v0.16.15
	k8s.io/client-go => k8s.io/client-go v0.16.15
)

require (
	github.com/99designs/basicauth-go v0.0.0-20160802081356-2a93ba0f464d
	github.com/bmatcuk/doublestar v1.3.2
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
	github.com/ozonep/drone v0.0.0-20201014173059-5ce6010a32f9
	github.com/sirupsen/logrus v1.6.0
	golang.org/x/crypto v0.0.0-20201002170205-7f63de1d35b0
	golang.org/x/sync v0.0.0-20200625203802-6e8e738ad208
	gopkg.in/alecthomas/kingpin.v2 v2.2.6
	k8s.io/api v0.16.15
	k8s.io/apimachinery v0.16.15
	k8s.io/client-go v10.0.0+incompatible
)
