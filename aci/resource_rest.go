package aci

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/ciscoecosystem/aci-go-client/client"
	"github.com/ciscoecosystem/aci-go-client/container"
	"github.com/ciscoecosystem/aci-go-client/models"
	"github.com/ghodss/yaml"
	"github.com/hashicorp/terraform-plugin-sdk/helper/schema"
)

const Created = "created"
const Deleted = "deleted"

var class string

const ErrDistinguishedNameNotFound = "The Dn is not present in the content"
/*
func resourceAciRest() *schema.Resource {
	return &schema.Resource{
		Create: resourceAciRestCreate,
		Update: resourceAciRestUpdate,
		Read:   resourceAciRestRead,
		Delete: resourceAciRestDelete,

		SchemaVersion: 1,

		Schema: map[string]*schema.Schema{
			"path": {
				Type:     schema.TypeString,
				Required: true,
			},
			// we set it automatically if file config is provided
			"class_name": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"content": {
				Type:     schema.TypeMap,
				Optional: true,
			},
			"dn": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"payload": {
				Type:     schema.TypeString,
				Optional: true,
			},
		},
	}
}

func resourceAciRestCreate(d *schema.ResourceData, m interface{}) error {
	cont, err := PostAndSetStatus(d, m, "created, modified")
	if err != nil {
		return err
	}
	classNameIntf := d.Get("class_name")
	className := classNameIntf.(string)
	dn := models.StripQuotes(models.StripSquareBrackets(cont.Search(className, "attributes", "dn").String()))
	if dn == "{}" {
		d.SetId(GetDN(d, m))
	} else {
		d.SetId(dn)
	}
	return resourceAciRestRead(d, m)

}

func resourceAciRestUpdate(d *schema.ResourceData, m interface{}) error {
	cont, err := PostAndSetStatus(d, m, "modified")
	if err != nil {
		return err
	}
	classNameIntf := d.Get("class_name")
	className := classNameIntf.(string)
	dn := models.StripQuotes(models.StripSquareBrackets(cont.Search(className, "attributes", "dn").String()))
	if dn == "{}" {
		d.SetId(GetDN(d, m))
	} else {
		d.SetId(dn)
	}
	return resourceAciRestRead(d, m)
}

func resourceAciRestRead(d *schema.ResourceData, m interface{}) error {
	return nil
}

func resourceAciRestDelete(d *schema.ResourceData, m interface{}) error {
	cont, err := PostAndSetStatus(d, m, Deleted)
	if err != nil {
		errCode := models.StripQuotes(models.StripSquareBrackets(cont.Search("imdata", "error", "attributes", "code").String()))
		// Ignore errors of type "Cannot delete object of class"
		if errCode == "107" {
			return nil
		}
		return err
	}
	d.SetId("")
	return nil
}   */

func GetDN(d *schema.ResourceData, m interface{}) string {
	aciClient := m.(*client.Client)
	path := d.Get("path").(string)
	cont, _ := aciClient.GetViaURL(path)
	dn := models.StripQuotes(models.StripSquareBrackets(cont.Search("imdata", class, "attributes", "dn").String()))
    return dn
}

// PostAndSetStatus is used to post schema and set the status
func PostAndSetStatus(d *schema.ResourceData, m interface{}, status string) (*container.Container, error) {
	aciClient := m.(*client.Client)
	path := d.Get("path").(string)
	var cont *container.Container
	var err error
	method := "POST"

	if content, ok := d.GetOk("content"); ok {
		contentStrMap := toStrMap(content.(map[string]interface{}))

		if classNameIntf, ok := d.GetOk("class_name"); ok {
			className := classNameIntf.(string)
			cont, err = preparePayload(className, contentStrMap)
			if err != nil {
				return nil, err
			}

		} else {
			return nil, errors.New("the className is required when content is provided explicitly")
		}

	} else if payload, ok := d.GetOk("payload"); ok {
		payloadStr := payload.(string)
		if len(payloadStr) == 0 {
			return nil, fmt.Errorf("payload cannot be empty string")
		}
		yamlJsonPayload, err := yaml.YAMLToJSON([]byte(payloadStr))
		if err != nil {
			// It may be possible that the payload is in JSON
			jsonPayload, err := container.ParseJSON([]byte(payloadStr))
			if err != nil {
				return nil, fmt.Errorf("invalid format for yaml/JSON payload")
			}
			cont = jsonPayload
		} else {
			// we have valid yaml payload and we were able to convert it to json
			cont, err = container.ParseJSON(yamlJsonPayload)

			if err != nil {
				return nil, fmt.Errorf("failed to convert YAML to JSON")
			}
		}
		if err != nil {
			return nil, fmt.Errorf("unable to parse the payload to JSON. Please check your payload")
		}

	} else {
		return nil, fmt.Errorf("either of payload or content is required")
	}
	var output map[string]interface{}
	err_output := json.Unmarshal([]byte(cont.String()), &output)
	if err_output != nil {
		return nil, err_output
	}
	for key, _ := range output {
		class = key
	}

	if status == Deleted {
		cont.Set(status, class, "attributes", "status")
	}
	req, err := aciClient.MakeRestRequest(method, path, cont, true)
	if err != nil {
		return nil, err
	}
	respCont, _, err := aciClient.Do(req)
	if err != nil {
		return respCont, err
	}
	err = client.CheckForErrors(respCont, method, false)
	if err != nil {
		return respCont, err
	}
	return cont, nil
}