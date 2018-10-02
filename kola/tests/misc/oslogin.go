// Copyright 2018 CoreOS, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package misc

import (
	"github.com/coreos/mantle/kola/cluster"
	"github.com/coreos/mantle/kola/register"
	"github.com/coreos/mantle/platform/machine/gcloud"
)

func init() {
	register.Register(&register.Test{
		Run:         OsLoginEnabled,
		ClusterSize: 0,
		Platforms:   []string{"gce"},
		Name:        "coreos.gce.oslogin.enabled",
		Distros:     []string{"cl"},
	})
}

func OsLoginEnabled(c cluster.TestCluster) {
	m, err := c.Cluster.(*gcloud.Cluster).NewMachineWithMetadata(nil, map[string]string{"enable-oslogin": "TRUE"})
	if err != nil {
		c.Fatal(err)
	}
	sshPubKey, err := c.Cluster.SshPubKey()
	if err != nil {
		c.Fatal(err)
	}

	user, err := c.Cluster.(*gcloud.Cluster).Api().CreateServiceAccount("normalUser", sshPubKey, []string{"iam.serviceAccountUser"})
	if err != nil {
		c.Fatal(err)
	}

	client, err := c.Cluster.(*gcloud.Cluster).UserSSHClient(m.IP(), user)
	if err != nil {
		c.Fatal(err)
	}

	_, _, err = c.Cluster.(*gcloud.Cluster).SSHWithClient(m, client, "true")
	if err != nil {
		c.Fatal(err)
	}
}
