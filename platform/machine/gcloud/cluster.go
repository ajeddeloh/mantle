// Copyright 2015 CoreOS, Inc.
// Copyright 2015 The Go Authors.
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

package gcloud

import (
	"os"
	"path/filepath"

	"golang.org/x/crypto/ssh/agent"

	"github.com/coreos/pkg/capnslog"

	ctplatform "github.com/coreos/container-linux-config-transpiler/config/platform"
	"github.com/coreos/mantle/platform"
	"github.com/coreos/mantle/platform/api/gcloud"
	"github.com/coreos/mantle/platform/conf"
)

type Cluster struct {
	*platform.BaseCluster
	api *gcloud.API
}

const (
	Platform platform.Name = "gcloud"
)

var (
	plog = capnslog.NewPackageLogger("github.com/coreos/mantle", "platform/machine/gcloud")
)

//hack remove this
func (c Cluster) Api() *gcloud.API {
	return c.api
}

func NewCluster(opts *gcloud.Options, rconf *platform.RuntimeConfig) (platform.Cluster, error) {
	api, err := gcloud.New(opts)
	if err != nil {
		return nil, err
	}

	bc, err := platform.NewBaseCluster(opts.Options, rconf, Platform, ctplatform.GCE)
	if err != nil {
		return nil, err
	}

	gc := &Cluster{
		BaseCluster: bc,
		api:         api,
	}

	return gc, nil
}

func (gc *Cluster) NewMachine(userdata *conf.UserData) (platform.Machine, error) {
	return gc.NewMachineWithMetadata(userdata, nil)
}

// Calling in parallel is ok
func (gc *Cluster) NewMachineWithMetadata(userdata *conf.UserData, metadata map[string]string) (platform.Machine, error) {
	conf, err := gc.RenderUserData(userdata, map[string]string{
		"$public_ipv4":  "${COREOS_GCE_IP_EXTERNAL_0}",
		"$private_ipv4": "${COREOS_GCE_IP_LOCAL_0}",
	})
	if err != nil {
		return nil, err
	}

	var keys []*agent.Key
	if !gc.RuntimeConf().NoSSHKeyInMetadata {
		keys, err = gc.Keys()
		if err != nil {
			return nil, err
		}
	}

	instance, err := gc.api.CreateInstance(conf.String(), keys, metadata)
	if err != nil {
		return nil, err
	}

	intip, extip := gcloud.InstanceIPs(instance)

	gm := &machine{
		gc:    gc,
		name:  instance.Name,
		intIP: intip,
		extIP: extip,
	}

	gm.dir = filepath.Join(gc.RuntimeConf().OutputDir, gm.ID())
	if err := os.Mkdir(gm.dir, 0777); err != nil {
		gm.Destroy()
		return nil, err
	}

	confPath := filepath.Join(gm.dir, "user-data")
	if err := conf.WriteFile(confPath); err != nil {
		gm.Destroy()
		return nil, err
	}

	if gm.journal, err = platform.NewJournal(gm.dir); err != nil {
		gm.Destroy()
		return nil, err
	}

	if err := platform.StartMachine(gm, gm.journal); err != nil {
		gm.Destroy()
		return nil, err
	}

	gc.AddMach(gm)

	return gm, nil
}
