package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// Provider returns a terraform.ResourceProvider.
func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"host": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALARDB_HOST", nil),
				Description: "The host address of the ScalarDB server.",
			},
			"port": {
				Type:        schema.TypeInt,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALARDB_PORT", 60051),
				Description: "The port of the ScalarDB server.",
			},
			"username": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("SCALARDB_USERNAME", nil),
				Description: "Username for ScalarDB authentication.",
			},
			"password": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("SCALARDB_PASSWORD", nil),
				Description: "Password for ScalarDB authentication.",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"scalardb_namespace": resourceScalarDBNamespace(),
			"scalardb_table":     resourceScalarDBTable(),
		},
		ConfigureContextFunc: providerConfigure,
	}
}

// providerConfigure configures the provider and returns a client.
func providerConfigure(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
	host := d.Get("host").(string)
	port := d.Get("port").(int)
	username := d.Get("username").(string)
	password := d.Get("password").(string)

	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	client := NewClient(host, port, username, password)

	return client, diags
}
