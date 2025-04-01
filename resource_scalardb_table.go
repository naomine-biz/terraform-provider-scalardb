package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceScalarDBTable() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceScalarDBTableCreate,
		ReadContext:   resourceScalarDBTableRead,
		DeleteContext: resourceScalarDBTableDelete,
		Schema: map[string]*schema.Schema{
			"namespace": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The namespace in which to create the table.",
			},
			"name": {
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
				Description: "The name of the table.",
			},
			"partition_key": {
				Type:        schema.TypeList,
				Required:    true,
				ForceNew:    true,
				Description: "The partition key columns of the table.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"clustering_key": {
				Type:        schema.TypeList,
				Optional:    true,
				ForceNew:    true,
				Description: "The clustering key columns of the table.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
			"column": {
				Type:        schema.TypeSet,
				Required:    true,
				ForceNew:    true,
				Description: "The columns of the table.",
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"name": {
							Type:        schema.TypeString,
							Required:    true,
							Description: "The name of the column.",
						},
						"type": {
							Type:     schema.TypeString,
							Required: true,
							ValidateFunc: validation.StringInSlice([]string{
								"INT", "BIGINT", "TEXT", "FLOAT", "DOUBLE", "BOOLEAN", "BLOB",
							}, false),
							Description: "The data type of the column.",
						},
					},
				},
			},
			"compaction_strategy": {
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "SizeTieredCompactionStrategy",
				Description: "The compaction strategy for the table.",
			},
			"clustering_order": {
				Type:        schema.TypeMap,
				Optional:    true,
				ForceNew:    true,
				Description: "The clustering order for the table.",
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
			},
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
	}
}

func resourceScalarDBTableCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	namespace := d.Get("namespace").(string)
	name := d.Get("name").(string)

	// Build columns map
	columns := make(map[string]map[string]interface{})
	for _, c := range d.Get("column").(*schema.Set).List() {
		column := c.(map[string]interface{})
		columnName := column["name"].(string)
		columnType := column["type"].(string)

		columns[columnName] = map[string]interface{}{
			"type": columnType,
		}
	}

	// Add primary key information
	partitionKey := make([]string, 0)
	for _, pk := range d.Get("partition_key").([]interface{}) {
		partitionKey = append(partitionKey, pk.(string))
		if col, ok := columns[pk.(string)]; ok {
			col["partition_key"] = true
		}
	}

	clusteringKey := make([]string, 0)
	for _, ck := range d.Get("clustering_key").([]interface{}) {
		clusteringKey = append(clusteringKey, ck.(string))
		if col, ok := columns[ck.(string)]; ok {
			col["clustering_key"] = true
		}
	}

	// Build options map
	options := make(map[string]interface{})
	options["compaction_strategy"] = d.Get("compaction_strategy").(string)

	if clusteringOrder, ok := d.GetOk("clustering_order"); ok {
		options["clustering_order"] = clusteringOrder.(map[string]interface{})
	}

	err := client.CreateTable(ctx, namespace, name, columns, options)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(fmt.Sprintf("%s.%s", namespace, name))

	return resourceScalarDBTableRead(ctx, d, m)
}

func resourceScalarDBTableRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	idParts := strings.Split(d.Id(), ".")
	if len(idParts) != 2 {
		return diag.Errorf("Invalid ID format: %s (expected namespace.table)", d.Id())
	}

	namespace := idParts[0]
	name := idParts[1]

	exists, err := client.TableExists(ctx, namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	if !exists {
		d.SetId("")
		return diags
	}

	_, _, err = client.GetTableSchema(ctx, namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	// In a real implementation, we would parse the columns and options
	// and update the state with the actual values.
	// For now, we'll just use the values from the configuration.

	d.Set("namespace", namespace)
	d.Set("name", name)

	return diags
}

func resourceScalarDBTableUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// In ScalarDB, table schema cannot be updated after creation.
	// So this is a no-op.
	return resourceScalarDBTableRead(ctx, d, m)
}

func resourceScalarDBTableDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	client := m.(*Client)

	var diags diag.Diagnostics

	idParts := strings.Split(d.Id(), ".")
	if len(idParts) != 2 {
		return diag.Errorf("Invalid ID format: %s (expected namespace.table)", d.Id())
	}

	namespace := idParts[0]
	name := idParts[1]

	err := client.DeleteTable(ctx, namespace, name)
	if err != nil {
		return diag.FromErr(err)
	}

	d.SetId("")

	return diags
}
