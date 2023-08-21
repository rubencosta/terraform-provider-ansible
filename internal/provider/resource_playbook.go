package provider

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type playbook struct {
	Playbook              types.String     `tfsdk:"playbook"`
	InventoryHosts        []inventoryHost  `tfsdk:"inventory_hosts"`
	AnsiblePlaybookBinary types.String     `tfsdk:"ansible_playbook_binary"`
	InventoryGroups       []inventoryGroup `tfsdk:"inventory_groups"`
	Replayable            types.Bool       `tfsdk:"replayable"`
	IgnorePlaybookFailure types.Bool       `tfsdk:"ignore_playbook_failure"`
	Verbosity             types.Int64      `tfsdk:"verbosity"`
	Tags                  []types.String   `tfsdk:"tags"`
	CheckMode             types.Bool       `tfsdk:"check_mode"`
	DiffMode              types.Bool       `tfsdk:"diff_mode"`
	ForceHandlers         types.Bool       `tfsdk:"force_handlers"`
	ExtraVars             types.String     `tfsdk:"extra_vars"`
	ID                    types.String     `tfsdk:"id"`
	Cmd                   types.String     `tfsdk:"cmd"`
	TempInventoryDir      types.String     `tfsdk:"temp_inventory_dir"`
	AnsiblePlaybookStdout types.String     `tfsdk:"ansible_playbook_stdout"`
	AnsiblePlaybookStderr types.String     `tfsdk:"ansible_playbook_stderr"`
}
type inventoryHost struct {
	Name      types.String   `tfsdk:"name"`
	Groups    []types.String `tfsdk:"groups"`
	Variables types.String   `tfsdk:"variables"`
}
type inventoryGroup struct {
	Name      types.String   `tfsdk:"name"`
	Children  []types.String `tfsdk:"children"`
	Variables types.String   `tfsdk:"variables"`
}

type playbookResource struct{}

func NewPlaybookResource() resource.Resource {
	return &playbookResource{}
}

// Metadata should return the full name of the resource, such as
// examplecloud_thing.
func (p *playbookResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_playbook"
}

// Schema should return the schema for this resource.
func (p *playbookResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			// Required settings
			"playbook": schema.StringAttribute{
				Required:    true,
				Description: "Path to ansible playbook.",
			},
			"inventory_hosts": schema.ListNestedAttribute{
				Required: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the host.",
						},
						"groups": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of group names.",
						},
						"variables": schema.StringAttribute{
							Optional:    true,
							Description: "yaml encoded map of variables.",
						},
					},
				},
			},

			// Optional settings
			"ansible_playbook_binary": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("ansible-playbook"),
				Description: "Path to ansible-playbook executable (binary).",
			},

			"inventory_groups": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Required:    true,
							Description: "Name of the host.",
						},
						"children": schema.ListAttribute{
							Optional:    true,
							ElementType: types.StringType,
							Description: "List of group names.",
						},
						"variables": schema.StringAttribute{
							Optional:    true,
							Description: "yaml encoded map of variables.",
						},
					},
				},
			},

			"replayable": schema.BoolAttribute{
				Optional: true,
				Description: "" +
					"If 'true', the playbook will be executed on every 'terraform apply' and with that, the resource" +
					" will be recreated. " +
					"If 'false', the playbook will be executed only on the first 'terraform apply'. " +
					"Note, that if set to 'true', when doing 'terraform destroy', it might not show in the destroy " +
					"output, even though the resource still gets destroyed.",
			},

			"ignore_playbook_failure": schema.BoolAttribute{
				Optional: true,
				Description: "This parameter is good for testing. " +
					"Set to 'true' if the desired playbook is meant to fail, " +
					"but still want the resource to run successfully.",
			},

			"verbosity": schema.Int64Attribute{ // verbosity is between = (0, 6)
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
				Description: "A verbosity level between 0 and 6. " +
					"Set ansible 'verbose' parameter, which causes Ansible to print more debug messages. " +
					"The higher the 'verbosity', the more debug details will be printed.",
			},

			"tags": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "List of tags of plays and tasks to run.",
			},

			"check_mode": schema.BoolAttribute{
				Optional: true,
				Description: "If 'true', playbook execution won't make any changes but " +
					"only change predictions will be made.",
			},

			"diff_mode": schema.BoolAttribute{
				Optional: true,
				Description: "" +
					"If 'true', when changing (small) files and templates, differences in those files will be shown. " +
					"Recommended usage with 'check_mode'.",
			},

			// connection configs are handled with extra_vars
			"force_handlers": schema.BoolAttribute{
				Optional:    true,
				Description: "If 'true', run handlers even if a task fails.",
			},

			// become configs are handled with extra_vars --> these are also connection configs
			"extra_vars": schema.StringAttribute{
				Optional:    true,
				Description: "A string of json or yaml encoded map of additional variables as: { var-1 = {key-1 = value-1, key-2 = value-2, ... }, ... }.",
			},

			// computed
			"id": schema.StringAttribute{
				Computed: true,
			},

			// debug output
			"cmd": schema.StringAttribute{
				Computed:    true,
				Description: "The command used to run ansible-playbook",
			},

			"temp_inventory_dir": schema.StringAttribute{
				Computed:    true,
				Description: "Path to created temporary inventory dir.",
			},

			"ansible_playbook_stdout": schema.StringAttribute{
				Computed:    true,
				Description: "An ansible-playbook CLI stdout output.",
			},

			"ansible_playbook_stderr": schema.StringAttribute{
				Computed:    true,
				Description: "An ansible-playbook CLI stderr output.",
			},
		},
	}
}

// Create is called when the provider must create a new resource. Config
// and planned state values should be read from the
// CreateRequest and new state values set on the CreateResponse.
func (pr *playbookResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var p playbook

	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)

	if resp.Diagnostics.HasError() {
		return
	}

	p.ID = types.StringValue(time.Now().String())
	p.runPlaybook()

	resp.Diagnostics.Append(resp.State.Set(ctx, &p)...)
}

// Read is called when the provider must read resource values in order
// to update state. Planned state values should be read from the
// ReadRequest and new state values set on the ReadResponse.
func (pr *playbookResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var p playbook
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !p.Replayable.ValueBool() {
		return
	}
	resp.State.RemoveResource(ctx)
}

// Update is called to update the state of the resource. Config, planned
// state, and prior state values should be read from the
// UpdateRequest and new state values set on the UpdateResponse.
func (pr *playbookResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var s playbook
	resp.Diagnostics.Append(req.State.Get(ctx, &s)...)
	if resp.Diagnostics.HasError() {
		return
	}
	RemoveDir(s.TempInventoryDir.ValueString())

	var p playbook

	resp.Diagnostics.Append(req.Plan.Get(ctx, &p)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p.ID = types.StringValue(time.Now().String())
	p.runPlaybook()

	// persist the values to state
	resp.Diagnostics.Append(resp.State.Set(ctx, &p)...)
}

// Delete is called when the provider must delete the resource. Config
// values may be read from the DeleteRequest.
//
// If execution completes without error, the framework will automatically
// call DeleteResponse.State.RemoveResource(), so it can be omitted
// from provider logic.
func (pr *playbookResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var p playbook
	resp.Diagnostics.Append(req.State.Get(ctx, &p)...)
	RemoveDir(p.TempInventoryDir.ValueString())
}

func (p *playbook) runPlaybook() {
	args := []string{}

	if p.TempInventoryDir.IsNull() || p.TempInventoryDir.ValueString() == "" {
		p.TempInventoryDir = types.StringValue(buildPlaybookInventory("inventory-*", p.InventoryHosts, p.InventoryGroups))
	}
	log.Printf("Temp Inventory Dir: %s", p.TempInventoryDir.ValueString())
	args = append(args, "-i", p.TempInventoryDir.ValueString())

	verbose := CreateVerboseSwitch(int(p.Verbosity.ValueInt64()))
	if verbose != "" {
		args = append(args, verbose)
	}

	if p.ForceHandlers.ValueBool() {
		args = append(args, "--force-handlers")
	}

	if len(p.Tags) > 0 {
		tmpTags := []string{}

		for _, tag := range p.Tags {
			tmpTags = append(tmpTags, tag.ValueString())
		}

		tagsStr := strings.Join(tmpTags, ",")
		args = append(args, "--tags", tagsStr)
	}

	if p.CheckMode.ValueBool() {
		args = append(args, "--check")
	}

	if p.DiffMode.ValueBool() {
		args = append(args, "--diff")
	}

	if !p.ExtraVars.IsNull() {
		args = append(args, "-e", p.ExtraVars.String())
	}

	args = append(args, p.Playbook.ValueString())
	// set up the args
	log.Print("[ANSIBLE ARGS]:")
	log.Print(args)

	runAnsiblePlay := exec.Command(p.AnsiblePlaybookBinary.ValueString(), args...)
	p.Cmd = types.StringValue(runAnsiblePlay.String())
	runAnsiblePlayOut, runAnsiblePlayErr := runAnsiblePlay.CombinedOutput()
	p.AnsiblePlaybookStdout = types.StringValue(string(runAnsiblePlayOut))
	p.AnsiblePlaybookStderr = types.StringNull()

	if runAnsiblePlayErr != nil {
		p.AnsiblePlaybookStderr = types.StringValue(runAnsiblePlayErr.Error())
		playbookFailMsg := fmt.Sprintf("ERROR [ansible-playbook]: couldn't run ansible-playbook\n%s! "+
			"There may be an error within your playbook.\n%v",
			p.Playbook.ValueString(),
			runAnsiblePlayErr,
		)
		if !p.IgnorePlaybookFailure.ValueBool() {
			log.Fatal(playbookFailMsg)
		} else {
			log.Print(playbookFailMsg)
		}
	}

	log.Printf("LOG [ansible-playbook]: %s", runAnsiblePlayOut)

	// Wait for playbook execution to finish, then remove the temporary dir
	err := runAnsiblePlay.Wait()
	if err != nil {
		log.Printf("LOG [ansible-playbook]: didn't wait for playbook to execute: %v", err)
	}
}
