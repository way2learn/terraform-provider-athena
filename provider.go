package athena

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Provider() *schema.Provider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"scheme": {
				Type:        schema.TypeString,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_SCHEME", "https"),
				Description: "ATHENA REST endpoint service http(s) scheme",
			},
			"address": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_ADDRESS", nil),
				Description: "ATHENA REST endpoint service host address",
			},
			"port": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_PORT", nil),
				Description: "ATHENA REST endpoint service port number",
			},
			"user": {
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_USER", nil),
				Description: "ATHENA REST endpoint user name",
			},
			"password": {
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_PASSWORD", nil),
				Description: "ATHENA REST endpoint password",
			},
			"verify_ssl": {
				Type:        schema.TypeBool,
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("ATHENA_VERIFY_SSL", true),
				Description: "Verify SSL certificates for ATHENA endpoints",
			},
		},
		ResourcesMap: map[string]*schema.Resource{
			"athena_ipam_record": resourceIPAMReservation(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"athena_ipam_policy": dataSourceIPAMPolicy(),
		},
		ConfigureFunc: configureProvider,
	}
}

type Config struct {
	scheme    string
	address   string
	port      string
	user      string
	password  string
	verifySSL bool
}

func configureProvider(d *schema.ResourceData) (interface{}, error) {
	return NewConfig(
		d.Get("scheme").(string),
		d.Get("address").(string),
		d.Get("port").(string),
		d.Get("user").(string),
		d.Get("password").(string),
		d.Get("verify_ssl").(bool),
	), nil
}

func NewConfig(scheme string, address string, port string, user string, password string, verifySSL bool) Config {
	return Config{
		scheme:    scheme,
		address:   address,
		port:      port,
		user:      user,
		password:  password,
		verifySSL: verifySSL,
	}
}
