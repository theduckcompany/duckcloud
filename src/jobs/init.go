package jobs

import (
	"github.com/theduckcompany/duckcloud/src/jobs/fsgc"
	"github.com/theduckcompany/duckcloud/src/service/inodes"
	"github.com/theduckcompany/duckcloud/src/tools"
	"go.uber.org/fx"
)

func StartJobs(lc fx.Lifecycle, inodes inodes.Service, tools tools.Tools) {
	fsgc.StartJob(lc, inodes, tools)
}
