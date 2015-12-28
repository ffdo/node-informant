package data

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	meshVpnStats = `
	{
	      "groups": {
	        "do01": {
	          "peers": {
	            "do01100": {
	              "established": 522401.615
	            },
	            "do01200": null
	          }
	        },
	        "do02": {
	          "peers": {
	            "do02100": null,
	            "do02200": null
	          }
	        }
	      }
	    }
	`
)

func TestParsingMeshVPNStats(t *testing.T) {
	assert := assert.New(t)

	meshVpn := MeshVPNStruct{}

	err := json.Unmarshal([]byte(meshVpnStats), &meshVpn)
	assert.Nil(err)

	assert.NotNil(meshVpn.Groups["do01"])
	assert.NotNil(meshVpn.Groups["do01"].Peers["do01100"])
	assert.Equal(float64(522401.615), meshVpn.Groups["do01"].Peers["do01100"].Established)
}
