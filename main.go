package main

// import (
// 	"github.com/hashicorp/terraform/plugin"
// 	"github.com/hashicorp/terraform/terraform"
// )

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/your-username/terraform-provider-athena/athena"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() terraform.ResourceProvider {
			return athena.Provider()
		},
	})
}
