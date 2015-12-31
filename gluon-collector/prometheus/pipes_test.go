package prometheus

import (
	"testing"

	"github.com/ffdo/node-informant/gluon-collector/config"
	"github.com/ffdo/node-informant/gluon-collector/data"
	cfg "github.com/olebedev/config"
	"github.com/stretchr/testify/assert"
)

var (
	testConfig = `
  prometheus:
    namelabel: true
    sitecodelabel: true
  `
)

func TestCreationOfPrometheusLabels(t *testing.T) {
	assert := assert.New(t)
	nodeinfo := data.NodeInfo{
		Hostname: "Testnode",
		System: data.SystemStruct{
			SiteCode: "fftest",
		},
		NodeId: "1122",
	}
	var err error
	config.Global, err = cfg.ParseYaml(testConfig)
	assert.Nil(err)

	prmcfg, err := config.Global.Get("prometheus")
	assert.Nil(err)
	assert.NotNil(prmcfg)
	assert.Equal([]string{"1122", "Testnode", "fftest"}, getLabels(nodeinfo))
}
