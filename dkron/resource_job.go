package dkron

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func resourceDkronJob() *schema.Resource {
	return &schema.Resource{
		CreateContext: resourceDkronJobCreate,
		ReadContext:   resourceDkronJobRead,
		UpdateContext: resourceDkronJobUpdate,
		DeleteContext: resourceDkronJobDelete,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"parent_job": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
				Default:  nil,
			},
			"schedule": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"timezone": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"owner": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"retries": {
				Type:     schema.TypeInt,
				Required: false,
				Optional: true,
			},
			"owner_email": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"disabled": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
			},
			"concurrency": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"executor": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"command": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"timeout": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"cwd": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"shell": {
				Type:     schema.TypeBool,
				Required: false,
				Optional: true,
			},
			"mem_limit_kb": {
				Type:     schema.TypeString,
				Required: false,
				Optional: true,
			},
			"project": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"allowed_exitcodes": {
				Type:     schema.TypeString,
				Required: true,
				Optional: false,
			},
			"tags": {
				Type:     schema.TypeMap,
				Required: true,
				Optional: false,
			},
			"processors": {
				Type:     schema.TypeList,
				Required: true,
				Optional: false,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"type": {
							Type:         schema.TypeString,
							Required:     true,
							Optional:     false,
							ValidateFunc: validation.StringInSlice([]string{"files", "log", "syslog", "fluent"}, false),
						},
						"forward": {
							Type:     schema.TypeString,
							Required: false,
							Optional: true,
							Default:  nil,
						},
						"log_dir": {
							Type:     schema.TypeString,
							Required: false,
							Optional: true,
							Default:  nil,
						},
					},
				},
			},
			"dependent_jobs": {
				Type:     schema.TypeList,
				Required: false,
				Optional: true,
				Elem: &schema.Schema{
					Type: schema.TypeString,
				},
				Default: nil,
			},
		},
	}
}

type Job struct {
	Name           string                 `json:"name"`
	Schedule       string                 `json:"schedule"`
	Owner          string                 `json:"owner"`
	OwnerEmail     string                 `json:"owner_email"`
	Disabled       bool                   `json:"disabled"`
	Tags           map[string]interface{} `json:"tags"`
	DependentJobs  []interface{}          `json:"dependent_jobs"`
	Retries        int                    `json:"retries"`
	Processors     map[string]interface{} `json:"processors"`
	Concurrency    string                 `json:"concurrency"`
	Executor       string                 `json:"executor"`
	Timezone       string                 `json:"timezone"`
	ParentJob      string                 `json:"parent_job"`
	ExecutorConfig struct {
		Command          string `json:"command"`
		Timeout          string `json:"timeout"`
		Project          string `json:"project"`
		MemLimitKb       string `json:"mem_limit_kb"`
		Cwd              string `json:"cwd"`
		Shell            string `json:"shell"`
		AllowedExitCodes string `json:"allowed_exitcodes"`
	} `json:"executor_config"`
}

func resourceDkronJobCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	config := m.(DkronConfig)

	job := new(Job)
	job.Name = d.Get("name").(string)
	job.ParentJob = d.Get("parent_job").(string)
	job.Schedule = d.Get("schedule").(string)
	job.Timezone = d.Get("timezone").(string)
	job.Owner = d.Get("owner").(string)
	job.OwnerEmail = d.Get("owner_email").(string)
	job.Disabled = d.Get("disabled").(bool)
	job.Retries = d.Get("retries").(int)
	job.Executor = d.Get("executor").(string)
	job.ExecutorConfig.Command = d.Get("command").(string)
	job.ExecutorConfig.Timeout = d.Get("timeout").(string)
	job.ExecutorConfig.Project = d.Get("project").(string)
	job.ExecutorConfig.MemLimitKb = d.Get("mem_limit_kb").(string)
	job.ExecutorConfig.Cwd = d.Get("cwd").(string)
	job.ExecutorConfig.AllowedExitCodes = d.Get("allowed_exitcodes").(string)
	job.ExecutorConfig.Shell = "false"
	if d.Get("shell").(bool) {
		job.ExecutorConfig.Shell = "true"
	}
	job.Concurrency = d.Get("concurrency").(string)
	job.Tags = d.Get("tags").(map[string]interface{})
	job.DependentJobs = d.Get("dependent_jobs").([]interface{})
	job.Processors = make(map[string]interface{})

	processors := d.Get("processors").([]interface{})
	for _, p := range processors {
		processor := p.(map[string]interface{})
		if processor["forward"] == "" {
			delete(processor, "forward")
		}
		if processor["log_dir"] == "" {
			delete(processor, "log_dir")
		}
		job.Processors[processor["type"].(string)] = processor
	}

	endpoint := fmt.Sprintf("%s/v1/jobs", config.host)

	log.Printf("[INFO] Posting job creation to %s endpoint", endpoint)

	payload, err := json.Marshal(job)
	if err != nil {
		return diag.FromErr(err)
	}

	log.Printf("[INFO] Payload created. %s", payload)

	request, err := http.NewRequest("POST", endpoint, bytes.NewBuffer(payload))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Post request created. ")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Request sent. ")

	body, _ := ioutil.ReadAll(response.Body)
	log.Println("response Status:", response.Status)
	log.Println("response Headers:", response.Header)
	log.Println("response Body:", string(body))

	if response.StatusCode != 201 {
		return diag.FromErr(errors.New(string(body)))
	}

	jobResponse := new(Job)
	json.Unmarshal(body, &jobResponse)

	d.SetId(jobResponse.Name)

	return diags
}

func resourceDkronJobRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	config := m.(DkronConfig)
	endpoint := fmt.Sprintf("%s/v1/jobs/%s", config.host, d.Id())
	log.Printf("[INFO] Request to %s endpoint", endpoint)

	request, err := http.NewRequest("GET", endpoint, bytes.NewBufferString(""))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Request created. ")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Request sent. ")

	body, _ := ioutil.ReadAll(response.Body)
	log.Println("response Status:", response.Status)
	log.Println("response Headers:", response.Header)
	log.Println("response Body:", string(body))

	job := new(Job)
	json.Unmarshal(body, &job)

	d.Set("name", job.Name)
	d.Set("parent_job", job.ParentJob)
	d.Set("schedule", job.Schedule)
	d.Set("timezone", job.Timezone)
	d.Set("owner", job.Owner)
	d.Set("owner_email", job.OwnerEmail)
	d.Set("disalbed", job.Disabled)
	d.Set("retries", job.Retries)
	d.Set("executor", job.Executor)
	d.Set("command", job.ExecutorConfig.Command)
	d.Set("timeout", job.ExecutorConfig.Timeout)
	d.Set("project", job.ExecutorConfig.Project)
	d.Set("mem_limit_kb", job.ExecutorConfig.MemLimitKb)
	d.Set("cwd", job.ExecutorConfig.Cwd)
	d.Set("allowed_exitcodes", job.ExecutorConfig.AllowedExitCodes)
	d.Set("concurrency", job.Concurrency)
	d.Set("tags", job.Tags)
	d.Set("dependent_jobs", job.DependentJobs)
	processors := make([]interface{}, 0)
	for i, p := range job.Processors {
		processor := p.(map[string]interface{})
		processor["type"] = i
		processors = append(processors, processor)
	}
	d.Set("processors", processors)

	return diags
}

func resourceDkronJobUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	old, new := d.GetChange("name")
	if old != new {
		resourceDkronJobDelete(ctx, d, m)
	}
	return resourceDkronJobCreate(ctx, d, m)
}

func resourceDkronJobDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// Warning or errors can be collected in a slice type
	var diags diag.Diagnostics

	config := m.(DkronConfig)
	endpoint := fmt.Sprintf("%s/v1/jobs/%s", config.host, d.Id())
	log.Printf("[INFO] Request to %s endpoint", endpoint)

	request, err := http.NewRequest("DELETE", endpoint, bytes.NewBufferString(""))
	request.Header.Set("Content-Type", "application/json")
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Request created. ")

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return diag.FromErr(err)
	}
	log.Printf("[INFO] Request sent. ")

	body, _ := ioutil.ReadAll(response.Body)
	log.Println("response Status:", response.Status)
	log.Println("response Headers:", response.Header)
	log.Println("response Body:", string(body))

	return diags
}
