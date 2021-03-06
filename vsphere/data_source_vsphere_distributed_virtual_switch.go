package vsphere

import (
	"fmt"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/vim25/types"
)

func dataSourceVSphereDistributedVirtualSwitch() *schema.Resource {
	return &schema.Resource{
		Read: dataSourceVSphereDistributedVirtualSwitchRead,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The name of the distributed virtual switch. This can be a name or path.",
				Required:    true,
			},
			"datacenter_id": &schema.Schema{
				Type:        schema.TypeString,
				Description: "The managed object ID of the datacenter the DVS is in. This is required if the supplied path is not an absolute path containing a datacenter and there are multiple datacenters in your infrastructure.",
				Optional:    true,
			},
			"uplinks": &schema.Schema{
				Type:        schema.TypeList,
				Computed:    true,
				Description: "The uplink ports on this DVS.",
				Elem:        &schema.Schema{Type: schema.TypeString},
			},
		},
	}
}

func dataSourceVSphereDistributedVirtualSwitchRead(d *schema.ResourceData, meta interface{}) error {
	client := meta.(*VSphereClient).vimClient
	if err := validateVirtualCenter(client); err != nil {
		return err
	}

	name := d.Get("name").(string)
	var dc *object.Datacenter
	if dcID, ok := d.GetOk("datacenter_id"); ok {
		var err error
		dc, err = datacenterFromID(client, dcID.(string))
		if err != nil {
			return fmt.Errorf("cannot locate datacenter: %s", err)
		}
	}
	dvs, err := dvsFromPath(client, name, dc)
	if err != nil {
		return fmt.Errorf("error fetching distributed virtual switch: %s", err)
	}
	props, err := dvsProperties(dvs)
	if err != nil {
		return fmt.Errorf("error fetching DVS properties: %s", err)
	}

	d.SetId(props.Uuid)
	uplinkPolicy := props.Config.(*types.VMwareDVSConfigInfo).UplinkPortPolicy.(*types.DVSNameArrayUplinkPortPolicy)
	if err := flattenDVSNameArrayUplinkPortPolicy(d, uplinkPolicy); err != nil {
		return err
	}

	return nil
}
