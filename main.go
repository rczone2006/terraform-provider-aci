package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/plugin"
//	"github.com/rczone2006/terraform-provider-aci/aci"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: aci.Provider})
}
