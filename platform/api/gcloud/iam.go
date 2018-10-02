// Copyright 2016 CoreOS, Inc.
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
	"fmt"
	"math/rand"

	"google.golang.org/api/iam/v1"
	//"google.golang.org/api/oslogin/v1beta"
)

func (a *API) CreateServiceAccount(name, sshKey string, roles []string) (string, error) {
	acctId := fmt.Sprintf("kola-%d", rand.Uint64())
	project := fmt.Sprintf("projects/%s", a.options.Project)
	fmt.Println(acctId)
	acct, err := a.iam.Projects.ServiceAccounts.Create(project, &iam.CreateServiceAccountRequest{
		AccountId: acctId,
		ServiceAccount: &iam.ServiceAccount{
			DisplayName: name,
		},
	}).Do()
	if err != nil {
		return "", err
	}
	// todo defer cleanup

	// plan: create key, reauth with it, add ssh key with oslogin api derived from that login

	fmt.Println(acct.Email)
	policyPath := fmt.Sprintf("projects/%s/serviceAccounts/%s", a.options.Project, acct.Email)
	for _, role := range roles {
		_, err = a.iam.Projects.ServiceAccounts.SetIamPolicy(policyPath, &iam.SetIamPolicyRequest{
			Policy: &iam.Policy{
				Bindings: []*iam.Binding{
					{
						Members: []string{fmt.Sprintf("serviceAccount:%s", acct.Email)},
						Role:    fmt.Sprintf("roles/%s", role),
					},
				},
			},
		}).Do()
		if err != nil {
			return "", fmt.Errorf("Error adding roles: %v", err)
		}
	}
	// todo defer cleanup

	/*info, err := a.oslogin.Users.ImportSshPublicKey("users/" + acct.Email, &oslogin.SshPublicKey{
		Key: sshKey,
	}).Do()
	if err != nil {
		return "", fmt.Errorf("Error importing ssh key:", err)
	}*/
	info, err := a.oslogin.Users.GetLoginProfile("users/" + acct.Email).Do()
	if err != nil {
		return "", err
	}

	return info.PosixAccounts[0].Username, nil
}
