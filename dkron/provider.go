package dkron

import (
	"context"
	"errors"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider -
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				Optional:    false,
				DefaultFunc: schema.EnvDefaultFunc("DKRON_HOST", nil),
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"dkron_job": resourceDkronJob(),
		},
		DataSourcesMap:       map[string]*schema.Resource{},
		ConfigureContextFunc: providerConfigure,
	}
}

type DkronConfig struct {
	host string
}

func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)

	// validate
	if host == "" {
		return nil, diag.FromErr(errors.New("provider variable 'host' should not be empty"))
	}

	// create config
	config := DkronConfig{
		host: host,
	}

	return config, nil
}
