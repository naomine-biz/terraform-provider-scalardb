package main

import (
	"context"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func resourceScalarDBNamespace() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalarDBNamespaceCreate,
		ReadContext:   resourceScalarDBNamespaceRead,
		DeleteContext: resourceScalarDBNamespaceDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the namespace.",
			},
			"replication_factor": {
				Type:        schema.TypeInt,
				Optional:    true,
				Default:     1,
				ForceNew:    true,
				Description: "The replication factor for the namespace.",
			},
			"strategy_class": {
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "SimpleStrategy",
				ForceNew:    true,
				Description: "The replication strategy class for the namespace.",
			},
			"durable_writes": {
				Type:        schema.TypeBool,
				Optional:    true,
				Default:     true,
				ForceNew:    true,
				Description: "Whether to use durable writes for the namespace.",
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceScalarDBNamespaceCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	name := d.Get("name").(string)

	options := make(map[string]interface{})
	options["replication_factor"] = d.Get("replication_factor").(int)
	options["strategy_class"] = d.Get("strategy_class").(string)
	options["durable_writes"] = d.Get("durable_writes").(bool)

	err := client.CreateNamespace(ctx, name, options)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(name)

	return resourceScalarDBNamespaceRead(ctx, d, m)
}

func resourceScalarDBNamespaceRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	name := d.Id()

	exists, err := client.NamespaceExists(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if !exists {
		d.SetId("")
		return diags
	}

	// In a real implementation, we would fetch the namespace details from ScalarDB
	// and update the state with the actual values.
	// For now, we'll just use the values from the configuration.

	return diags
}

func resourceScalarDBNamespaceUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// In ScalarDB, namespace properties like replication_factor cannot be updated
	// after creation. So this is a no-op.
	return resourceScalarDBNamespaceRead(ctx, d, m)
}

func resourceScalarDBNamespaceDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	name := d.Id()

	err := client.DeleteNamespace(ctx, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
