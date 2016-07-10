package fcm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	// "github.com/jpillora/backoff"
)

const (
	fcm_server_url     = "https://fcm.googleapis.com/fcm/send"
	MAX_TTL            = 2419200
	Priority_HIGH      = "high"
	Priority_NORMAL    = "normal"
	retry_after_header = "Retry-After"
	error_key          = "error"
)

var (
	// minBackoff       = 1 * time.Second
	// maxBackoff       = 10 * time.Second
	retreyableErrors = map[string]bool{
		"Unavailable":         true,
		"InternalServerError": true,
	}

	// for testing purposes
	fcmServerUrl = fcm_server_url
)

type FcmClient struct {
	ApiKey  string
	Message FcmMsg
}

type FcmMsg struct {
	Data                  map[string]string   `json:"data,omitempty"`
	To                    string              `json:"to,omitempty"`
	RegistrationIds       []string            `json:"registration_ids,omitempty"`
	CollapseKey           string              `json:"collapse_key,omitempty"`
	Priority              string              `json:"priority,omitempty"`
	Notification          NotificationPayload `json:"notification,omitempty"`
	ContentAvailable      bool                `json:"content_available,omitempty"`
	DelayWhileIdle        bool                `json:"delay_while_idle,omitempty"`
	TimeToLive            int                 `json:"time_to_live,omitempty"`
	RestrictedPackageName string              `json:"restricted_package_name,omitempty"`
	DryRun                bool                `json:"dry_run,omitempty"`
	Condition             string              `json:"condition,omitempty"`
}

type FcmResponseStatus struct {
	Ok            bool
	StatusCode    int
	MulticastId   int                 `json:"multicast_id"`
	Success       int                 `json:"success"`
	Fail          int                 `json:"failure"`
	Canonical_ids int                 `json:"canonical_ids"`
	Results       []map[string]string `json:"results,omitempty"`
	MsgId         int                 `json:"message_id,omitempty"`
	Err           string              `json:"error,omitempty"`
	RetryAfter    string
}

type NotificationPayload struct {
	Title        string `json:"title,omitempty"`
	Body         string `json:"body,omitempty"`
	Icon         string `json:"icon,omitempty"`
	Sound        string `json:"sound,omitempty"`
	Badge        string `json:"badge,omitempty"`
	Tag          string `json:"tag,omitempty"`
	Color        string `json:"color,omitempty"`
	ClickAction  string `json:"click_action,omitempty"`
	BodyLocKey   string `json:"body_loc_key,omitempty"`
	BodyLocArgs  string `json:"body_loc_args,omitempty"`
	TitleLocKey  string `json:"title_loc_key,omitempty"`
	TitleLocArgs string `json:"title_loc_args,omitempty"`
}

type InstanceIdResult struct {
	Application        string                         `json:"application,omitempty"`
	AuthorizedEntity   string                         `json:"authorizedEntity,omitempty"`
	ApplicationVersion string                         `json:"applicationVersion,omitempty"`
	AppSigner          string                         `json:"appSigner,omitempty"`
	AttestStatus       string                         `json:"attestStatus,omitempty"`
	Platform           string                         `json:"platform,omitempty"`
	connectionType     string                         `json:"connectionType,omitempty"`
	connectDate        string                         `json:"connectDate,omitempty"`
	rel                map[string]InstanceIdRelTopics `json:"rel,omitempty"`
}

type InstanceIdRelTopics struct {
	TopicName map[string]TopicDate
}

type TopicDate map[string]string

func NewFcmClient(apiKey string) *FcmClient {
	fcmc := new(FcmClient)
	fcmc.ApiKey = apiKey

	return fcmc
}

func (this *FcmClient) NewFcmTopicMsg(to string, body map[string]string) *FcmClient {

	this.NewFcmMsgTo(to, body)

	return this
}

func (this *FcmClient) NewFcmMsgTo(to string, body map[string]string) *FcmClient {
	this.Message.To = to
	this.Message.Data = body

	return this
}

func (this *FcmClient) SetMsgData(body map[string]string) *FcmClient {

	this.Message.Data = body

	return this

}

func (this *FcmClient) NewFcmRegIdsMsg(list []string, body map[string]string) *FcmClient {
	this.newDevicesList(list)
	this.Message.Data = body

	return this

}

func (this *FcmClient) newDevicesList(list []string) *FcmClient {
	this.Message.RegistrationIds = make([]string, len(list))
	copy(this.Message.RegistrationIds, list)

	return this

}

func (this *FcmClient) AppendDevices(list []string) *FcmClient {

	this.Message.RegistrationIds = append(this.Message.RegistrationIds, list...)

	return this
}

func (this *FcmClient) apiKeyHeader() string {
	return fmt.Sprintf("key=%v", this.ApiKey)
}

func (this *FcmClient) sendOnce() (*FcmResponseStatus, error) {

	fcmRespStatus := new(FcmResponseStatus)

	jsonByte, err := this.Message.toJsonByte()
	if err != nil {
		fcmRespStatus.Ok = false
		return fcmRespStatus, err
	}

	request, err := http.NewRequest("POST", fcmServerUrl, bytes.NewBuffer(jsonByte))
	request.Header.Set("Authorization", this.apiKeyHeader())
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	response, err := client.Do(request)

	if err != nil {
		fcmRespStatus.Ok = false
		return fcmRespStatus, err
	}
	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	fcmRespStatus.StatusCode = response.StatusCode

	fcmRespStatus.RetryAfter = response.Header.Get(retry_after_header)

	if response.StatusCode == 200 && err == nil {

		fcmRespStatus.Ok = true

		eror := fcmRespStatus.parseStatusBody(body)
		if eror != nil {
			return fcmRespStatus, eror
		}

		return fcmRespStatus, nil

	} else {
		fcmRespStatus.Ok = false

		eror := fcmRespStatus.parseStatusBody(body)
		if eror != nil {
			return fcmRespStatus, eror
		}

		return fcmRespStatus, err
	}

	return fcmRespStatus, nil

}

func (this *FcmClient) Send() (*FcmResponseStatus, error) {
	return this.sendOnce()

}
func (this *FcmMsg) toJsonByte() ([]byte, error) {

	return json.Marshal(this)

}

func (this *FcmResponseStatus) parseStatusBody(body []byte) error {

	if err := json.Unmarshal([]byte(body), &this); err != nil {
		return err
	}
	return nil

}

func (this *FcmClient) SetPriorety(p string) {

	if p == Priority_HIGH {
		this.Message.Priority = Priority_HIGH
	} else {
		this.Message.Priority = Priority_NORMAL
	}
}

func (this *FcmClient) SetCollapseKey(val string) *FcmClient {

	this.Message.CollapseKey = val

	return this
}

func (this *FcmClient) SetNotificationPayload(payload *NotificationPayload) *FcmClient {

	this.Message.Notification = *payload

	return this
}

func (this *FcmClient) SetContentAvailable(isContentAvailable bool) *FcmClient {

	this.Message.ContentAvailable = isContentAvailable

	return this
}

func (this *FcmClient) SetDelayWhileIdle(isDelayWhileIdle bool) *FcmClient {

	this.Message.DelayWhileIdle = isDelayWhileIdle

	return this
}
func (this *FcmClient) SetTimeToLive(ttl int) *FcmClient {

	if ttl > MAX_TTL {

		this.Message.TimeToLive = MAX_TTL

	} else {

		this.Message.TimeToLive = ttl

	}
	return this
}

func (this *FcmClient) SetRestrictedPackageName(pkg string) *FcmClient {

	this.Message.RestrictedPackageName = pkg

	return this
}

func (this *FcmClient) SetDryRun(drun bool) *FcmClient {

	this.Message.DryRun = drun

	return this
}

func (this *FcmResponseStatus) PrintResults() {
	fmt.Println("Status Code   :", this.StatusCode)
	fmt.Println("Success       :", this.Success)
	fmt.Println("Fail          :", this.Fail)
	fmt.Println("Canonical_ids :", this.Canonical_ids)
	fmt.Println("Topic MsgId   :", this.MsgId)
	fmt.Println("Topic Err     :", this.Err)
	for i, val := range this.Results {
		fmt.Printf("Result(%d)> \n", i)
		for k, v := range val {
			fmt.Println("\t", k, " : ", v)
		}
	}
}

func (this *FcmResponseStatus) IsTimeout() bool {
	if this.StatusCode > 500 {
		return true
	} else if this.StatusCode == 200 {
		for _, val := range this.Results {
			for k, v := range val {
				if k == error_key && retreyableErrors[v] == true {
					return true
				}
			}
		}
	}

	return false
}


func (this *FcmResponseStatus) GetRetryAfterTime() (t time.Duration, e error) {
	t, e = time.ParseDuration(this.RetryAfter)
	return
}


func (this *FcmClient) SetCondition(condition string) *FcmClient {
	this.Message.Condition = condition
	return this
}

// func newBackoffHandler() *backoff.Backoff {
// 	b := &backoff.Backoff{
//
// 		Min:    minBackoff,
// 		Max:    maxBackoff,
// 		Jitter: true,
// 	}
//
// 	return b
// }

// func setMinBackoff(m time.Duration) {
// 	minBackoff = m
// }
//
// func setMaxBackoff(m time.Duration) {
// 	maxBackoff = m
// }
