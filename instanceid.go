package fcm

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	instance_id_info_with_details_srv_url = "https://iid.googleapis.com/iid/info/%s?details=true"
	instance_id_info_no_details_srv_url   = "https://iid.googleapis.com/iid/info/%s"
	subscribe_instanceid_to_topic_srv_url = "https://iid.googleapis.com/iid/v1/%s/rel/topics/%s"

	batch_add_srv_url = "https://iid.googleapis.com/iid/v1:batchAdd"
	batch_rem_srv_url = "https://iid.googleapis.com/iid/v1:batchRemove"

	apns_batch_import_srv_url = "https://iid.googleapis.com/iid/v1:batchImport"

	apns_token_key = "apns_token"
	status_key     = "status"
	reg_token_key  = "registration_token"

	topics = "/topics/"
)

var (
	batchErrors = map[string]bool{
		"NOT_FOUND":        true,
		"INVALID_ARGUMENT": true,
		"INTERNAL":         true,
		"TOO_MANY_TOPICS":  true,
	}
)

type InstanceIdInfoResponse struct {
	Application        string                                  `json:"application,omitempty"`
	AuthorizedEntity   string                                  `json:"authorizedEntity,omitempty"`
	ApplicationVersion string                                  `json:"applicationVersion,omitempty"`
	AppSigner          string                                  `json:"appSigner,omitempty"`
	AttestStatus       string                                  `json:"attestStatus,omitempty"`
	Platform           string                                  `json:"platform,omitempty"`
	ConnectionType     string                                  `json:"connectionType,omitempty"`
	ConnectDate        string                                  `json:"connectDate,omitempty"`
	Error              string                                  `json:"error,omitempty"`
	Rel                map[string]map[string]map[string]string `json:"rel,omitempty"`
}

type SubscribeResponse struct {
	Error      string `json:"error,omitempty"`
	Status     string
	StatusCode int
}

type BatchRequest struct {
	To        string   `json:"to,omitempty"`
	RegTokens []string `json:"registration_tokens,omitempty"`
}

type BatchResponse struct {
	Error      string              `json:"error,omitempty"`
	Results    []map[string]string `json:"results,omitempty"`
	Status     string
	StatusCode int
}

type ApnsBatchRequest struct {
	App        string   `json:"application,omitempty"`
	Sandbox    bool     `json:"sandbox,omitempty"`
	ApnsTokens []string `json:"apns_tokens,omitempty"`
}

type ApnsBatchResponse struct {
	Results    []map[string]string `json:"results,omitempty"`
	Error      string              `json:"error,omitempty"`
	Status     string
	StatusCode int
}

func (this *FcmClient) GetInfo(withDetails bool, instanceIdToken string) (*InstanceIdInfoResponse, error) {

	var request_url string = generateGetInfoUrl(instance_id_info_no_details_srv_url, instanceIdToken)

	if withDetails == true {
		request_url = generateGetInfoUrl(instance_id_info_with_details_srv_url, instanceIdToken)
	}

	request, err := http.NewRequest("GET", request_url, nil)
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	infoResponse, err := parseGetInfo(body)
	if err != nil {
		return nil, err
	}

	return infoResponse, nil
}

func parseGetInfo(body []byte) (*InstanceIdInfoResponse, error) {

	info := new(InstanceIdInfoResponse)

	if err := json.Unmarshal([]byte(body), &info); err != nil {
		return nil, err
	}

	return info, nil

}

func (this *InstanceIdInfoResponse) PrintResults() {
	fmt.Println("Error     : ", this.Error)
	fmt.Println("App       : ", this.Application)
	fmt.Println("Auth      : ", this.AuthorizedEntity)
	fmt.Println("Ver       : ", this.ApplicationVersion)
	fmt.Println("Sig       : ", this.AppSigner)
	fmt.Println("Att       : ", this.AttestStatus)
	fmt.Println("Platform  : ", this.Platform)
	fmt.Println("Connection: ", this.ConnectionType)
	fmt.Println("ConnDate  : ", this.ConnectDate)
	fmt.Println("Rel       : ")
	for k, v := range this.Rel {
		fmt.Println(k, " --> ")
		for k2, v2 := range v {
			fmt.Println("\t", k2, "\t|")
			fmt.Println("\t\t", "addDate", " : ", v2["addDate"])
		}
	}
}

func generateGetInfoUrl(srv string, instanceIdToken string) string {
	return fmt.Sprintf(srv, instanceIdToken)
}

func (this *FcmClient) SubscribeToTopic(instanceIdToken string, topic string) (*SubscribeResponse, error) {

	request, err := http.NewRequest("POST", generateSubToTopicUrl(instanceIdToken, topic), nil)
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	subResponse, err := parseSubscribeResponse(body, response)
	if err != nil {
		return nil, err
	}

	return subResponse, nil
}

func parseSubscribeResponse(body []byte, resp *http.Response) (*SubscribeResponse, error) {

	subResp := new(SubscribeResponse)

	subResp.Status = resp.Status
	subResp.StatusCode = resp.StatusCode

	if err := json.Unmarshal(body, &subResp); err != nil {
		return nil, err
	}
	return subResp, nil
}

func (this *SubscribeResponse) PrintResults() {

	fmt.Println("Response Status: ", this.Status)
	fmt.Println("Response Code  : ", this.StatusCode)
	if this.StatusCode != 200 {
		fmt.Println("Error          : ", this.Error)
	}

}

func generateSubToTopicUrl(instaceId string, topic string) string {
	Tmptopic := strings.ToLower(topic)
	if strings.Contains(Tmptopic, "/topics/") {
		tmp := strings.Split(topic, "/")
		topic = tmp[len(tmp)-1]
	}
	return fmt.Sprintf(subscribe_instanceid_to_topic_srv_url, instaceId, topic)
}

func (this *FcmClient) BatchSubscribeToTopic(tokens []string, topic string) (*BatchResponse, error) {

	jsonByte, err := generateBatchRequest(tokens, topic)
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", batch_add_srv_url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result, err := generateBatchResponse(body)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("Parsing response error")
	}
	result.Status = response.Status
	result.StatusCode = response.StatusCode

	return result, nil
}

func (this *FcmClient) BatchUnsubscribeFromTopic(tokens []string, topic string) (*BatchResponse, error) {

	jsonByte, err := generateBatchRequest(tokens, topic)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	request, err := http.NewRequest("POST", batch_rem_srv_url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		fmt.Println(err)
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		fmt.Println(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	result, err := generateBatchResponse(body)
	if err != nil {
		return nil, err
	}
	if result == nil {
		return nil, errors.New("Parsing response error")
	}
	result.Status = response.Status
	result.StatusCode = response.StatusCode

	return result, nil
}

func (this *BatchResponse) PrintResults() {
	fmt.Println("Error       : ", this.Error)
	fmt.Println("Status      : ", this.Status)
	fmt.Println("Status Code : ", this.StatusCode)
	for i, val := range this.Results {
		if batchErrors[val["error"]] == true {
			fmt.Println("ID: ", i, " | ", val["error"])
		}
	}
}

func generateBatchRequest(tokens []string, topic string) ([]byte, error) {
	envelope := new(BatchRequest)
	envelope.To = topics + extractTopicName(topic)
	envelope.RegTokens = make([]string, len(tokens))
	copy(envelope.RegTokens, tokens)

	return json.Marshal(envelope)

}

func extractTopicName(inTopic string) (result string) {
	Tmptopic := strings.ToLower(inTopic)
	if strings.Contains(Tmptopic, "/topics/") {
		tmp := strings.Split(inTopic, "/")
		result = tmp[len(tmp)-1]
		return
	}

	result = inTopic
	return
}

func generateBatchResponse(resp []byte) (*BatchResponse, error) {
	result := new(BatchResponse)

	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil

}

func (this *FcmClient) ApnsBatchImportRequest(apnsReq *ApnsBatchRequest) (*ApnsBatchResponse, error) {

	jsonByte, err := apnsReq.ToByte()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", apns_batch_import_srv_url, bytes.NewBuffer(jsonByte))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	result, err := parseApnsBatchResponse(body)
	if err != nil {
		return nil, err
	}

	if result == nil {
		return nil, errors.New("Parsing Request error")
	}

	result.Status = response.Status
	result.StatusCode = response.StatusCode

	return result, nil
}

func (this *ApnsBatchRequest) ToByte() ([]byte, error) {
	data, err := json.Marshal(this)
	if err != nil {
		return nil, err
	}

	return data, nil
}

func parseApnsBatchResponse(resp []byte) (*ApnsBatchResponse, error) {

	result := new(ApnsBatchResponse)
	if err := json.Unmarshal(resp, &result); err != nil {
		return nil, err
	}

	return result, nil

}

func (this *ApnsBatchResponse) PrintResults() {
	fmt.Println("Status     : ", this.Status)
	fmt.Println("StatusCode : ", this.StatusCode)
	fmt.Println("Error      : ", this.Error)
	for i, val := range this.Results {
		fmt.Println(i, ":")
		fmt.Println("\tAPNS Token", val[apns_token_key])
		fmt.Println("\tStatus    ", val[status_key])
		fmt.Println("\tReg Token  ", val[reg_token_key])
	}
}
