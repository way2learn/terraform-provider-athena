package athena

import (
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/pkg/errors"
)

func resourceIPAMReservation() *schema.Resource {
	return &schema.Resource{
		Create: resourceIPAMReservationCreate,
		Read:   resourceIPAMReservationRead,
		Update: resourceIPAMReservationUpdate,
		Delete: resourceIPAMReservationDelete,
		Schema: map[string]*schema.Schema{
			"hostname": {
				Type:     schema.TypeString,
				Required: true,
			},
			"computed_hostname": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"policy_id": {
				Type:     schema.TypeInt,
				Required: true,
			},
			"workspace_url": {
				Type:     schema.TypeString,
				Computed: true,
				Optional: true,
			},
			"ip_address": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"netmask": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"gateway": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"network": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"subnet": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"primary_dns": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"secondary_dns": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"nic_label": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
				ForceNew: true,
			},
			"dns_suffix": {
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},
			"dns_search_suffix": {
				Type: schema.TypeList,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Optional: true,
				Computed: true,
			},
			"template_properties": {
				Type:     schema.TypeMap,
				Optional: true,
			},
		},
		Timeouts: &schema.ResourceTimeout{
			Create: schema.DefaultTimeout(5 * time.Minute),
			Delete: schema.DefaultTimeout(5 * time.Minute),
		},
	}
}

func bindIPAMReservationResource(d *schema.ResourceData, ipamRecord *IPAMReservation) error {
	log.Println("athena.bindIPAMReservationResource")

	if err := d.Set("computed_hostname", ipamRecord.Hostname); err != nil {
		return errors.WithMessage(err, "Cannot set name: "+ipamRecord.Hostname)
	}

	if err := d.Set("workspace_url", ipamRecord.Links.Workspace.Href); err != nil {
		return errors.WithMessage(err, "Cannot set workspace: "+ipamRecord.Links.Workspace.Href)
	}

	if err := d.Set("ip_address", ipamRecord.IPaddress); err != nil {
		return errors.WithMessage(err, "Cannot set IPAddress: "+ipamRecord.IPaddress)
	}

	if err := d.Set("netmask", ipamRecord.Netmask); err != nil {
		return errors.WithMessage(err, "Cannot set Netmask "+ipamRecord.Netmask)
	}

	if err := d.Set("primary_dns", ipamRecord.PrimaryDNS); err != nil {
		return errors.WithMessage(err, "Cannot set Primmary DNS: "+ipamRecord.PrimaryDNS)
	}

	if err := d.Set("secondary_dns", ipamRecord.SecondaryDNS); err != nil {
		return errors.WithMessage(err, "Cannot set Secondary DNS: "+ipamRecord.SecondaryDNS)
	}

	if err := d.Set("gateway", ipamRecord.Gateway); err != nil {
		return errors.WithMessage(err, "Cannot set Gateway: "+ipamRecord.Gateway)
	}

	if err := d.Set("network", ipamRecord.Network); err != nil {
		return errors.WithMessage(err, "Cannot set Network: "+ipamRecord.Network)
	}

	if err := d.Set("subnet", ipamRecord.Subnet); err != nil {
		return errors.WithMessage(err, "Cannot set Subnet: "+ipamRecord.Subnet)
	}

	if err := d.Set("nic_label", ipamRecord.NicLabel); err != nil {
		return errors.WithMessage(err, "Cannot set NicLabel: "+ipamRecord.NicLabel)
	}

	if err := d.Set("dns_suffix", ipamRecord.DNSSuffix); err != nil {
		return errors.WithMessage(err, "Cannot set DNSSuffix: "+ipamRecord.DNSSuffix)
	}

	ipamPolicyURLSplit := strings.Split(ipamRecord.Links.Policy.Href, "/")
	ipamPolicyID := ipamPolicyURLSplit[len(ipamPolicyURLSplit)-2]
	ipamPolicyIDInt, _ := strconv.Atoi(ipamPolicyID)
	if err := d.Set("policy_id", ipamPolicyIDInt); err != nil {
		return errors.WithMessage(err, "Cannot set policy")
	}

	return nil
}

func resourceIPAMReservationCreate(d *schema.ResourceData, m interface{}) error {
	log.Println("athena.resourceIPAMReservationCreate")

	var ipam_Suffixes []string
	for _, group := range d.Get("dns_search_suffix").([]interface{}) {
		ipam_Suffixes = append(ipam_Suffixes, group.(string))
	}

	config := m.(Config)

	newIPAMRecord := IPAMReservation{
		Hostname:           d.Get("hostname").(string),
		PolicyID:           d.Get("policy_id").(int),
		WorkspaceURL:       d.Get("workspace_url").(string),
		IPaddress:          d.Get("ip_address").(string),
		Netmask:            d.Get("netmask").(string),
		Subnet:             d.Get("subnet").(string),
		Gateway:            d.Get("gateway").(string),
		Network:            d.Get("network").(string),
		PrimaryDNS:         d.Get("primary_dns").(string),
		SecondaryDNS:       d.Get("secondary_dns").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		NicLabel:           d.Get("nic_label").(string),
		TemplateProperties: d.Get("template_properties").(map[string]interface{}),
	}

	ipamRecord, err := config.NewAthenaApiClient().CreateIPAMReservation(&newIPAMRecord)
	if err != nil {
		return err
	}
	d.SetId(strconv.Itoa(ipamRecord.ID))

	return bindIPAMReservationResource(d, ipamRecord)
}

func resourceIPAMReservationRead(d *schema.ResourceData, m interface{}) error {
	log.Println("athena.resourceIPAMReservationRead")

	config := m.(Config)

	id := d.Id()
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	ipamRecord, err := config.NewAthenaApiClient().GetIPAMReservation(intID)
	if err != nil {
		return err
	}

	return bindIPAMReservationResource(d, ipamRecord)
}

func resourceIPAMReservationUpdate(d *schema.ResourceData, m interface{}) error {
	log.Println("athena.resourceIPAMReservationUpdate")

	// Determine if a change is needed
	changed := (d.HasChange("hostname") ||
		d.HasChange("computed_hostname") ||
		d.HasChange("policy_id") ||
		d.HasChange("workspace_url")) ||
		d.HasChange("ip_address") ||
		d.HasChange("netmask") ||
		d.HasChange("subnet") ||
		d.HasChange("network") ||
		d.HasChange("gateway") ||
		d.HasChange("primary_dns") ||
		d.HasChange("secondary_dns") ||
		d.HasChange("dns_suffix") ||
		d.HasChange("nic_label") ||
		d.HasChange("template_properties")

	if !changed {
		return nil
	}

	var ipam_Suffixes []string
	for _, group := range d.Get("dns_search_suffix").([]interface{}) {
		ipam_Suffixes = append(ipam_Suffixes, group.(string))
	}

	// Make the API call to update the computer account
	config := m.(Config)

	// Create the desired IPAM Reservation
	id := d.Id()
	desiredIPAMRecord := IPAMReservation{
		Hostname:           d.Get("hostname").(string),
		PolicyID:           d.Get("policy_id").(int),
		WorkspaceURL:       d.Get("workspace_url").(string),
		IPaddress:          d.Get("ip_address").(string),
		Netmask:            d.Get("netmask").(string),
		Subnet:             d.Get("subnet").(string),
		Gateway:            d.Get("gateway").(string),
		Network:            d.Get("network").(string),
		PrimaryDNS:         d.Get("primary_dns").(string),
		SecondaryDNS:       d.Get("secondary_dns").(string),
		DNSSuffix:          d.Get("dns_suffix").(string),
		NicLabel:           d.Get("nic_label").(string),
		TemplateProperties: d.Get("template_properties").(map[string]interface{}),
	}

	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	ipamRecord, err := config.NewAthenaApiClient().UpdateIPAMReservation(intID, &desiredIPAMRecord)
	if err != nil {
		return err
	}

	return bindIPAMReservationResource(d, ipamRecord)
}

func resourceIPAMReservationDelete(d *schema.ResourceData, m interface{}) error {
	log.Println("athena.resourceIPAMReservationDelete")

	config := m.(Config)

	id := d.Id()
	intID, err := strconv.Atoi(id)
	if err != nil {
		return err
	}

	return config.NewAthenaApiClient().DeleteIPAMReservation(intID)
}
