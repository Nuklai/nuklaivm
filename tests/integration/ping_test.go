// Copyright (C) 2024, Nuklai. All rights reserved.
// See the file LICENSE for licensing terms.

// ping_test.go
package integration

import (
	"context"

	"github.com/onsi/ginkgo/v2"
	"github.com/stretchr/testify/require"
)

var _ = ginkgo.Describe("ping", func() {
	require := require.New(ginkgo.GinkgoT())

	ginkgo.It("can ping", func() {
		for _, inst := range instances {
			cli := inst.cli
			ok, err := cli.Ping(context.Background())
			require.NoError(err)
			require.True(ok)
		}
	})
})
