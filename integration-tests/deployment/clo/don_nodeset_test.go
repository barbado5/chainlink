package clo_test

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-go/testify/require"

	"github.com/smartcontractkit/chainlink/integration-tests/deployment/clo"
	"github.com/smartcontractkit/chainlink/integration-tests/deployment/clo/models"
)

// this is hacky, but there is no first class concept of a chain writer node in CLO
// in prod, probably better to make an explicit list of pubkeys if we can't add a category or product type
// sufficient for testing
var (
	writerFilter = func(n *models.Node) bool {
		return strings.Contains(n.Name, "Prod Keystone Cap One") && !strings.Contains(n.Name, "Boot")
	}

	assetFilter = func(n *models.Node) bool {
		return strings.Contains(n.Name, "Prod Keystone Asset") && !strings.Contains(n.Name, "Bootstrap")
	}

	wfFilter = func(n *models.Node) bool {
		return strings.Contains(n.Name, "Prod Keystone One") && !strings.Contains(n.Name, "Boot")
	}
)

func TestFilterNopNodes(t *testing.T) {
	t.Skipf("this test is for generating test data only")
	// use for generating keystone deployment test data
	// `./bin/fmscli --config ~/.fmsclient/prod.yaml login`
	// `./bin/fmscli --config ~/.fmsclient/prod.yaml get nodeOperators > /tmp/all-clo-nops.json`
	path := "/tmp/all-clo-nops.json"
	f, err := os.ReadFile(path)
	require.NoError(t, err)
	type cloData struct {
		Nops []*models.NodeOperator `json:"nodeOperators"`
	}
	var d cloData
	require.NoError(t, json.Unmarshal(f, &d))
	require.NotEmpty(t, d.Nops)
	allNops := d.Nops
	sort.Slice(allNops, func(i, j int) bool {
		return allNops[i].ID < allNops[j].ID
	})

	ksFilter := func(n *models.Node) bool {
		return writerFilter(n) || assetFilter(n) || wfFilter(n)
	}
	ksNops := clo.FilterNopNodes(allNops, ksFilter)
	require.NotEmpty(t, ksNops)
	b, err := json.MarshalIndent(ksNops, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile("testdata/keystone_nops.json", b, 0644)) // nolint: gosec
}
func TestDonNodeset(t *testing.T) {
	keystoneNops := loadTestNops(t, "testdata/keystone_nops.json")

	m := clo.CapabilityNodeSets(keystoneNops, map[string]clo.FilterFuncT[*models.Node]{
		"workflow":    wfFilter,
		"chainWriter": writerFilter,
		"asset":       assetFilter,
	})
	assert.Len(t, m, 3)
	assert.Len(t, m["workflow"], 10)
	assert.Len(t, m["chainWriter"], 10)
	assert.Len(t, m["asset"], 16)

	// can be used to derive the test data for the keystone deployment
	updateTestData := true
	if updateTestData {
		b, err := json.MarshalIndent(m["workflow"], "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile("testdata/workflow_nodes.json", b, 0644)) // nolint: gosec

		b, err = json.MarshalIndent(m["chainWriter"], "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile("testdata/chain_writer_nodes.json", b, 0644)) // nolint: gosec

		b, err = json.MarshalIndent(m["asset"], "", "  ")
		require.NoError(t, err)
		require.NoError(t, os.WriteFile("testdata/asset_nodes.json", b, 0644)) // nolint: gosec
	}

	gotWFNops := m["workflow"]
	sort.Slice(gotWFNops, func(i, j int) bool {
		return gotWFNops[i].ID < gotWFNops[j].ID
	})
	expectedWorkflowNops := loadTestNops(t, "testdata/workflow_nodes.json")
	assert.True(t, reflect.DeepEqual(gotWFNops, expectedWorkflowNops), "workflow nodes do not match")

	gotChainWriterNops := m["chainWriter"]
	sort.Slice(gotChainWriterNops, func(i, j int) bool {
		return gotChainWriterNops[i].ID < gotChainWriterNops[j].ID
	})
	expectedChainWriterNops := loadTestNops(t, "testdata/chain_writer_nodes.json")
	assert.True(t, reflect.DeepEqual(gotChainWriterNops, expectedChainWriterNops), "chain writer nodes do not match")
}

func loadTestNops(t *testing.T, path string) []*models.NodeOperator {
	f, err := os.ReadFile(path)
	require.NoError(t, err)
	var nodes []*models.NodeOperator
	require.NoError(t, json.Unmarshal(f, &nodes))
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	return nodes
}
