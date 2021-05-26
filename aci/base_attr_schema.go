package aci

import "github.com/hashicorp/terraform-plugin-sdk/helper/schema"

func GetBaseAttrSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"description": {
			Type:     schema.TypeString,
			Optional: true,
			Computed: true,
		},
		"annotation": {
			Type:     schema.TypeString,
			Optional: true,
			Default:  "orchestrator:terraform",
		},
	}
}

// AppendBaseAttrSchema adds the BaseAttr to any schema
func AppendBaseAttrSchema(attrs map[string]*schema.Schema) map[string]*schema.Schema {
	for key, value := range GetBaseAttrSchema() {
		attrs[key] = value
	}
	return attrs
}