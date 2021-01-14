module gitlab.com/gitlab-org/gitlab-shell

go 1.13

require (
	github.com/mattn/go-shellwords v0.0.0-20190425161501-2444a32a19f4
	github.com/otiai10/copy v1.0.1
	github.com/sirupsen/logrus v1.6.0
	github.com/stretchr/testify v1.5.1
	gitlab.com/gitlab-org/gitaly v1.68.0
	gitlab.com/gitlab-org/labkit v0.0.0-20200908084045-45895e129029
	google.golang.org/grpc v1.35.0
	gopkg.in/yaml.v2 v2.2.8
)

// go get tries to enforce semantic version compatibility via module paths.
// We can't upgrade to Gitaly v13.x.x from v1.x.x without using a manual override.
// See https://gitlab.com/gitlab-org/gitaly/-/issues/3177 for more details.
replace gitlab.com/gitlab-org/gitaly => gitlab.com/gitlab-org/gitaly v0.0.0-20201001041716-3f5e218def93
