// Copyright 2020 CloudBolt Software
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this
// file, You can obtain one at https://mozilla.org/MPL/2.0/.

package athena

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const ApiVersion = "api/v3"
const ApiNamespace = "onefuse"
const WorkspaceResourceType = "workspaces"
const ModuleEndpointResourceType = "endpoints"
const ModulePolicyResourceType = "modulePolicies"
const ModuleDepoloymentResourceType = "moduleManagedObjects"
const IPAMReservationResourceType = "ipamReservations"
const IPAMPolicyResourceType = "ipamPolicies"
const JobStatusResourceType = "jobStatus"
const RenderTemplateType = "templateTester"
const JobSuccess = "Successful"
const JobFailed = "Failed"

type AthenaAPIClient struct {
	config *Config
}

type CustomName struct {
	Id        int
	Name      string
	DnsSuffix string
}

type LinkRef struct {
	Href  string `json:"href,omitempty"`
	Title string `json:"title,omitempty"`
}

type Workspace struct {
	Links *struct {
		Self LinkRef `json:"self,omitempty"`
	} `json:"_links,omitempty"`
	ID   int    `json:"id,omitempty"`
	Name string `json:"name,omitempty"`
}

type WorkspacesListResponse struct {
	Embedded struct {
		Workspaces []Workspace `json:"workspaces"`
	} `json:"_embedded"`
}

type IPAMReservation struct {
	Links *struct {
		Self        LinkRef `json:"self,omitempty"`
		Workspace   LinkRef `json:"workspace,omitempty"`
		Policy      LinkRef `json:"policy,omitempty"`
		JobMetadata LinkRef `json:"jobMetadata,omitempty"`
	} `json:"_links,omitempty"`
	ID                 int                    `json:"id,omitempty"`
	Hostname           string                 `json:"hostname,omitempty"`
	PolicyID           int                    `json:"policyId,omitempty"`
	Policy             string                 `json:"policy,omitempty"`
	WorkspaceURL       string                 `json:"workspace,omitempty"`
	IPaddress          string                 `json:"ipAddress,omitempty"`
	Gateway            string                 `json:"gateway,omitempty"`
	PrimaryDNS         string                 `json:"primaryDns"`
	SecondaryDNS       string                 `json:"secondaryDns"`
	Network            string                 `json:"network,omitempty"`
	Subnet             string                 `json:"subnet,omitempty"`
	DNSSuffix          string                 `json:"dnsSuffix,omitempty"`
	Netmask            string                 `json:"netmask,omitempty"`
	NicLabel           string                 `json:"nicLabel,omitempty"`
	TemplateProperties map[string]interface{} `json:"template_properties,omitempty"`
}

type IPAMPolicyResponse struct {
	Embedded struct {
		IPAMPolicies []IPAMPolicy `json:"ipamPolicies"`
	} `json:"_embedded"`
}

type IPAMPolicy struct {
	Links *struct {
		Self      LinkRef `json:"self,omitempty"`
		Workspace LinkRef `json:"workspace,omitempty"`
	} `json:"_links,omitempty"`
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}

type JobStatus struct {
	Links *struct {
		Self          LinkRef `json:"self,omitempty"`
		JobMetadata   LinkRef `json:"jobMetadata,omitempty"`
		ManagedObject LinkRef `json:"managedObject,omitempty"`
		Policy        LinkRef `json:"policy,omitempty"`
		Workspace     LinkRef `json:"workspace,omitempty"`
	} `json:"_links,omitempty"`
	ID                  int    `json:"id,omitempty"`
	JobStateDescription string `json:"jobStateDescription,omitempty"`
	JobState            string `json:"jobState,omitempty"`
	JobTrackingID       string `json:"jobTrackingId,omitempty"`
	JobType             string `json:"jobType,omitempty"`
	ErrorDetails        *struct {
		Code   int `json:"code,omitempty"`
		Errors *[]struct {
			Message string `json:"message,omitempty"`
		} `json:"errors,omitempty"`
	} `json:"errorDetails,omitempty"`
}

type RenderTemplateRequest struct {
	Template           string                 `json:"template,omitempty"`
	TemplateProperties map[string]interface{} `json:"template_properties,omitempty"`
}

func (c *Config) NewAthenaApiClient() *AthenaAPIClient {
	return &AthenaAPIClient{
		config: c,
	}
}

func buildPostRequest(config *Config, resourceType string, requestEntity interface{}) (*http.Request, error) {
	url := collectionURL(config, resourceType)

	jsonBytes, err := json.Marshal(requestEntity)
	if err != nil {
		return nil, errors.WithMessage(err, "athena.apiClient: Failed to marshal request body to JSON")
	}

	requestBody := string(jsonBytes)
	payload := strings.NewReader(requestBody)

	req, err := http.NewRequest("POST", url, payload)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Unable to create request POST %s %s", url, requestBody))
	}

	setHeaders(req, config)

	return req, nil
}

func buildPutRequest(config *Config, resourceType string, requestEntity interface{}, id int) (*http.Request, error) {
	url := itemURL(config, resourceType, id)

	jsonBytes, err := json.Marshal(requestEntity)
	if err != nil {
		return nil, errors.WithMessage(err, "athena.apiClient: Failed to marshal request body to JSON")
	}

	requestBody := string(jsonBytes)
	payload := strings.NewReader(requestBody)

	req, err := http.NewRequest("PUT", url, payload)
	if err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Unable to create request PUT %s %s", url, requestBody))
	}

	setHeaders(req, config)

	return req, nil
}

//Create IPAM Reservation

func (apiClient *AthenaAPIClient) CreateIPAMReservation(newIPAMRecord *IPAMReservation) (*IPAMReservation, error) {
	log.Println("athena.apiClient: CreateIPAMReservation")

	config := apiClient.config

	var err error
	if newIPAMRecord.WorkspaceURL, err = findWorkspaceURLOrDefault(config, newIPAMRecord.WorkspaceURL); err != nil {
		return nil, err
	}

	if newIPAMRecord.Policy == "" {
		if newIPAMRecord.PolicyID != 0 {
			newIPAMRecord.Policy = itemURL(config, WorkspaceResourceType, newIPAMRecord.PolicyID)
		} else {
			return nil, errors.New("athena.apiClient: IPAM Record Create requires a PolicyID or Policy URL")
		}
	} else {
		return nil, errors.New("athena.apiClient: IPAM Record Create requires a PolicyID or Policy URL")
	}

	var req *http.Request
	if req, err = buildPostRequest(config, IPAMReservationResourceType, newIPAMRecord); err != nil {
		return nil, err
	}

	ipamRecord := IPAMReservation{}

	_, err = handleAsyncRequestAndFetchManagdObject(req, config, &ipamRecord, "POST")
	if err != nil {
		return nil, err
	}
	return &ipamRecord, nil
}

//Get IPAM Reservation

func (apiClient *AthenaAPIClient) GetIPAMReservation(id int) (*IPAMReservation, error) {
	log.Println("athena.apiClient: GetIPAMReservation")

	config := apiClient.config

	url := itemURL(config, IPAMReservationResourceType, id)

	ipamRecord := IPAMReservation{}
	err := doGet(config, url, &ipamRecord)
	if err != nil {
		return nil, err
	}
	return &ipamRecord, err
}

//Update IPAM Record

func (apiClient *AthenaAPIClient) UpdateIPAMReservation(id int, updatedIPAMReservation *IPAMReservation) (*IPAMReservation, error) {
	log.Println("athena.apiClient: UpdateIPAMReservation")

	return nil, errors.New("athena.apiClient: Not implemented yet")
}

func (apiClient *AthenaAPIClient) DeleteIPAMReservation(id int) error {
	log.Println("athena.apiClient: DeleteIPAMReservation")

	config := apiClient.config

	url := itemURL(config, IPAMReservationResourceType, id)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to create request DELETE %s", url))
	}

	setHeaders(req, config)

	_, err = handleAsyncRequest(req, config, "DELETE")
	return err
}

// End IPAM

// End vRA Deployment

// Start IPAM Policies

func (apiClient *AthenaAPIClient) GetIPAMPolicy(id int) (*IPAMPolicy, error) {
	log.Println("athena.apiClient: GetIPAMPolicy")
	return nil, errors.New("athena.apiClient: Not implemented yet")
}

func (apiClient *AthenaAPIClient) GetIPAMPolicyByName(name string) (*IPAMPolicy, error) {
	log.Println("athena.apiClient: GetIPAMPolicyByName")

	config := apiClient.config

	ipamPolicies := IPAMPolicyResponse{}
	entity, err := findEntityByName(config, name, IPAMPolicyResourceType, &ipamPolicies, "IPAMPolicies", "")
	if err != nil {
		return nil, err
	}
	ipamPolicy := entity.(IPAMPolicy)
	return &ipamPolicy, nil
}

// End IPAM Policies
// Start Jobs

func GetJobStatus(id int, config *Config) (*JobStatus, error) {
	log.Println("athena.apiClient: GetJobStatus")

	url := itemURL(config, JobStatusResourceType, id)
	result := JobStatus{}

	err := doGet(config, url, &result)
	if err != nil {
		return nil, err
	}

	return &result, nil
}

// End Jobs

func handleAsyncRequestAndFetchManagdObject(req *http.Request, config *Config, responseObject interface{}, httpVerb string) (jobStatus *JobStatus, err error) {

	if jobStatus, err = handleAsyncRequest(req, config, httpVerb); err != nil {
		return
	}

	url := urlFromHref(config, jobStatus.Links.ManagedObject.Href)
	err = doGet(config, url, &responseObject)
	if err != nil {
		return nil, err
	}

	return jobStatus, nil
}

func handleAsyncRequest(req *http.Request, config *Config, httpVerb string) (jobStatus *JobStatus, err error) {

	client := getHttpClient(config)

	res, err := client.Do(req)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		return jobStatus, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to do request %s %s %s", httpVerb, req.URL, body))
	}

	body, err := readResponse(res)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		return jobStatus, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to read response body from %s %s %s", httpVerb, req.URL, body))
	}
	defer res.Body.Close()

	if err = json.Unmarshal(body, &jobStatus); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to unmarshal response %s", string(body)))
	}

	jobStatus, err = waitForJob(jobStatus.ID, config)
	if err != nil {
		return
	}

	if err = checkForJobErrors(jobStatus); err != nil {
		return nil, err
	}

	return jobStatus, nil
}

func doGet(config *Config, url string, v interface{}) (err error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to create request GET %s", url))
	}

	setHeaders(req, config)

	client := getHttpClient(config)
	res, err := client.Do(req)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to do request GET %s", url))
	}

	if err = checkForErrors(res); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Request failed GET %s", url))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to read response body from GET %s", url))
	}
	defer res.Body.Close()

	if err = json.Unmarshal(body, &v); err != nil {
		return errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to unmarshal response %s", string(body)))
	}

	return nil
}

func waitForJob(jobID int, config *Config) (jobStatus *JobStatus, err error) {
	jobStatusDescription := ""
	PollingTimeoutMS := 3600000
	PollingIntervalMS := 5000
	startTime := time.Now()
	for jobStatusDescription != JobSuccess && jobStatusDescription != JobFailed {
		jobStatus, err = GetJobStatus(jobID, config)
		if err != nil {
			return nil, err
		}

		jobStatusDescription = jobStatus.JobState
		log.Println(jobStatus)

		time.Sleep(time.Duration(PollingIntervalMS) * time.Millisecond)
		if time.Since(startTime) > (time.Duration(PollingTimeoutMS) * time.Millisecond) {
			return nil, errors.New("Timed out while waiting for job to complete.")
		}
	}
	return jobStatus, nil
}

func findWorkspaceURLOrDefault(config *Config, workspaceURL string) (string, error) {
	// Default workspace if it was not provided
	if workspaceURL == "" {
		workspaceID, err := findDefaultWorkspaceID(config)
		if err != nil {
			return "", errors.WithMessage(err, "athena.apiClient: Failed to find default workspace")
		}
		workspaceIDInt, err := strconv.Atoi(workspaceID)
		if err != nil {
			return "", errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to convert Workspace ID '%s' to integer", workspaceID))
		}

		workspaceURL = itemURL(config, WorkspaceResourceType, workspaceIDInt)
	}
	return workspaceURL, nil
}

// Start Render Template

func (apiClient *AthenaAPIClient) RenderTemplate(template string, templateProperties map[string]interface{}) (*RenderTemplateResponse, error) {
	// this API endpoint is a POST, but only so we can pass in a body to be rendered by the templating engine
	// it behaves mostly like a GET, and doesn't create an object, just returns the rendered value.
	log.Println("athena.apiClient: RenderTemplate")

	config := apiClient.config

	requestBody := RenderTemplateRequest{
		Template:           template,
		TemplateProperties: templateProperties,
	}

	var err error

	var req *http.Request
	if req, err = buildPostRequest(config, RenderTemplateType, requestBody); err != nil {
		return nil, err
	}

	client := getHttpClient(config)

	res, err := client.Do(req)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to do request POST %s %s", req.URL, body))
	}

	if err = checkForErrors(res); err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Request failed POST %s %s", req.URL, body))
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		body, _ := ioutil.ReadAll(req.Body)
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to read response body from POST %s %s", req.URL, body))
	}
	defer res.Body.Close()

	renderTemplateResponse := RenderTemplateResponse{}
	if err = json.Unmarshal(body, &renderTemplateResponse); err != nil {
		return nil, errors.WithMessage(err, fmt.Sprintf("athena.apiClient: Failed to unmarshal response %s", string(body)))
	}
	renderedTemplate := renderTemplateResponse

	return &renderedTemplate, nil
}

// End Render Template

func findDefaultWorkspaceID(config *Config) (workspaceID string, err error) {
	fmt.Println("athena.findDefaultWorkspaceID")

	filter := "filter=name.exact:Default"
	url := fmt.Sprintf("%s?%s", collectionURL(config, WorkspaceResourceType), filter)

	req, clientErr := http.NewRequest("GET", url, nil)
	if clientErr != nil {
		err = errors.WithMessage(clientErr, fmt.Sprintf("athena.findDefaultWorkspaceID: Failed to make request GET %s", url))
		return
	}

	setHeaders(req, config)

	client := getHttpClient(config)
	res, clientErr := client.Do(req)
	if clientErr != nil {
		err = errors.WithMessage(clientErr, fmt.Sprintf("athena.findDefaultWorkspaceID: Failed to do request GET %s", url))
		return
	}

	body, err := readResponse(res)
	if err != nil {
		return
	}
	defer res.Body.Close()

	var data WorkspacesListResponse
	json.Unmarshal(body, &data)

	workspaces := data.Embedded.Workspaces
	if len(workspaces) == 0 {
		err = errors.WithMessage(clientErr, "athena.findDefaultWorkspaceID: Failed to find default workspace!")
		return
	}
	workspaceID = strconv.Itoa(workspaces[0].ID)
	return
}

func findEntityByName(config *Config, name string, resourceType string, collectionResponse interface{},
	embeddedStructFieldName string, additionalFilters string) (interface{}, error) {

	url := fmt.Sprintf("%s?filter=name:%s%s", collectionURL(config, resourceType), name, additionalFilters)

	err := doGet(config, url, &collectionResponse)
	if err != nil {
		return nil, err
	}

	embeddedField := reflect.Indirect(reflect.ValueOf(collectionResponse)).FieldByName("Embedded")
	embedded := embeddedField.Interface()

	collectionField := reflect.Indirect(reflect.ValueOf(embedded)).FieldByName(embeddedStructFieldName)

	if collectionField.Len() < 1 {
		return nil, errors.New(fmt.Sprintf("athena.apiClient: Could not find %s '%s'!", resourceType, name))
	}

	entity := collectionField.Index(0).Interface()

	return entity, err
}

func getHttpClient(config *Config) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: !config.verifySSL},
	}
	return &http.Client{Transport: tr}
}

func readResponse(res *http.Response) (bytes []byte, err error) {
	err = checkForErrors(res)
	if err != nil {
		return
	}

	bytes, err = ioutil.ReadAll(res.Body)
	return
}

func checkForErrors(res *http.Response) error {
	if res.StatusCode >= 500 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		return errors.New(string(b))
	} else if res.StatusCode >= 400 {
		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return err
		}
		defer res.Body.Close()
		return errors.New(string(b))
	}
	return nil
}

func checkForJobErrors(jobStatus *JobStatus) error {
	if jobStatus.JobState != JobSuccess {
		return errors.New(fmt.Sprintf("Job %s (%d) failed with message %v", jobStatus.JobType, jobStatus.ID, *jobStatus.ErrorDetails.Errors))
	}
	return nil
}

func setStandardHeaders(req *http.Request) {
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "*/*")
	req.Header.Add("Cache-Control", "no-cache")
	req.Header.Add("accept-encoding", "gzip, deflate")
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("cache-control", "no-cache")
}

func setHeaders(req *http.Request, config *Config) {
	setStandardHeaders(req)
	req.Header.Add("Host", fmt.Sprintf("%s:%s", config.address, config.port))
	req.Header.Add("SOURCE", "Terraform")
	req.SetBasicAuth(config.user, config.password)
}

func collectionURL(config *Config, resourceType string) string {
	baseURL := fmt.Sprintf("%s://%s:%s", config.scheme, config.address, config.port)
	endpoint := path.Join(ApiVersion, ApiNamespace, resourceType)
	return fmt.Sprintf("%s/%s/", baseURL, endpoint)
}

func urlFromHref(config *Config, href string) string {
	return fmt.Sprintf("%s://%s:%s%s", config.scheme, config.address, config.port, href)
}

func itemURL(config *Config, resourceType string, id int) string {
	idString := strconv.Itoa(id)
	baseURL := collectionURL(config, resourceType)
	return fmt.Sprintf("%s%s/", baseURL, idString)
}
