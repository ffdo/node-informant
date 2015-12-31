package prometheus

import (
	"testing"

	"github.com/dereulenspiegel/node-informant/gluon-collector/config"
	"github.com/dereulenspiegel/node-informant/gluon-collector/data"
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

	nodeLabels := getLabels(nodeinfo, "metric")
	assert.Equal(4, len(nodeLabels))
	assert.Equal("1122", nodeLabels[0])
	assert.Equal("Testnode", nodeLabels[1])
	assert.Equal("fftest", nodeLabels[2])
	assert.Equal("metric", nodeLabels[3])
}
