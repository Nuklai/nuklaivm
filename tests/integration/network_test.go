// network_test.go
package integration

import (
	"context"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
)

var _ = ginkgo.Describe("network", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("can get network", func() {
		for _, inst := range instances {
			cli := inst.cli
			networkID, subnetID, chainID, err := cli.Network(context.Background())
			require.NoError(err)
			require.Equal(networkID, uint32(1))
			require.NotEqual(subnetID, ids.Empty)
			require.NotEqual(chainID, ids.Empty)
		}
	})
})
