// Copyright (C) The Arvados Authors. All rights reserved.
//
// SPDX-License-Identifier: AGPL-3.0

package dispatchcloud

import (
	"git.curoverse.com/arvados.git/lib/cloud"
	"git.curoverse.com/arvados.git/sdk/go/arvados"
	"golang.org/x/crypto/ssh"
)

type instanceSetProxy struct {
	cloud.InstanceSet
}

func (is *instanceSetProxy) Create(it arvados.InstanceType, id cloud.ImageID, tags cloud.InstanceTags, pk ssh.PublicKey) (cloud.Instance, error) {
	// TODO: return if Create failed recently with a RateLimitError or QuotaError
	return is.InstanceSet.Create(it, id, tags, pk)
}

func (is *instanceSetProxy) Instances(tags cloud.InstanceTags) ([]cloud.Instance, error) {
	// TODO: return if Instances failed recently with a RateLimitError
	return is.InstanceSet.Instances(tags)
}
