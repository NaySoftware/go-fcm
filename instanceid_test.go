package fcm

import (
	"testing"
)

func TestGenTopicUrl(t *testing.T) {
	expected := "https://iid.googleapis.com/iid/v1/DeviceToken/rel/topics/TopicNamE"
	result := generateSubToTopicUrl("DeviceToken", "TopicNamE")

	if result != expected {
		t.Error("Gen Topic Url Error")
	}
}

func TestGenInfoResult1(t *testing.T) {
	result1 := `{"applicationVersion":"1","connectDate":"2016-07-02","attestStatus":"NOT_ROOTED","application":"com.comp.company","scope":"*","authorizedEntity":"1234567891234","rel":{"topics":{"global":{"addDate":"2016-07-02"}}},"connectionType":"WIFI","appSigner":"c822c1b63853ed273b89687ac505f9faabcdefgh","platform":"ANDROID"}`

	resp, err := parseGetInfo([]byte(result1))
	if err != nil {
		t.Error("Parsing Error: ", err)
	}

	if resp == nil {
		t.Error("Parsing Error: Response is nil")
	}

}

func TestGenInfoResult2(t *testing.T) {
	result2 := `{
	"applicationVersion":"1",
	"connectDate":"2016-07-02",
	"attestStatus":"NOT_ROOTED",
	"application":"com.comp.company",
	"scope":"*",
	"authorizedEntity":"1234567891234",
	"rel":
	{
		"topics":
		{
			"global":{ "addDate":"2016-07-02" }
		}
	},
	"connectionType":"WIFI",
	"appSigner":"c822c1b63853ed273b89687ac505f9faabcdefgh",
	"platform":"ANDROID"
  }`

	resp, err := parseGetInfo([]byte(result2))
	if err != nil {
		t.Error("Parsing Error: ", err)
	}
	if resp == nil {
		t.Error("Parsing Error: Response is nil")
	}

}

func TestGenInfoResult3(t *testing.T) {
	result3 := `{"error":"No information found about this instance id."}`

	resp, err := parseGetInfo([]byte(result3))
	if err != nil {
		t.Error("Parsing Error: ", err)
	}
	if resp == nil {
		t.Error("Parsing Error: Response is nil")
	}

}

func TestParseApnsBatchToByte(t *testing.T) {
	batch1 := new(ApnsBatchRequest)
	batch1.App = "com.comp.company"
	batch1.Sandbox = true
	batch1.ApnsTokens = []string{
		"368dde283db539abc4a6419b1795b6131194703b816e4f624ffa12",
		"76b39c2b2ceaadee8400b8868c2f45325ab9831c1998ed70859d86",
	}

	batch2 := &ApnsBatchRequest{
		App:     "com.comp.company",
		Sandbox: true,
		ApnsTokens: []string{
			"368dde283db539abc4a6419b1795b6131194703b816e4f624ffa12",
			"76b39c2b2ceaadee8400b8868c2f45325ab9831c1998ed70859d86",
		},
	}

	resp, err := batch1.ToByte()
	if err != nil {
		t.Error("Parsing Error: ", err)
	}
	if resp == nil {
		t.Error("Parsing Error: batch1 is nil")
	}

	resp, err = batch2.ToByte()
	if err != nil {
		t.Error("Parsing Error: ", err)
	}
	if resp == nil {
		t.Error("Parsing Error: batch2 is nil")
	}
}

func TestExtractTopicName(t *testing.T) {
	topic := "/topics/news"
	expected := "news"

	if expected != extractTopicName(topic) {
		t.Error("Extracting topic name faild")
	}

	topic = "/TOPICS/alpha"
	expected = "alpha"

	if expected != extractTopicName(topic) {
		t.Error("Extracting topic name faild")
	}

	topic = "beta"
	expected = "beta"

	if expected != extractTopicName(topic) {
		t.Error("Extracting topic name faild")
	}

}
