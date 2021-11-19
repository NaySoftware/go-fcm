package fcm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTopicHandle_1(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmMsgTo("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_2(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestTopicHandle_3(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(topicHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	data2 := map[string]string{
		"msg": "Hello bits",
	}

	c.NewFcmTopicMsg("/topics/topicName", data)

	c.SetMsgData(data2)
	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}
}

func TestRegIdHandle_1(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
		"token1",
		"token2",
	}

	c.NewFcmRegIdsMsg(ids, data)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func TestRegIdHandle_2(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
	chgUrl(srv)
	defer srv.Close()

	c := NewFcmClient("key")

	data := map[string]string{
		"msg": "Hello World",
		"sum": "Happy Day",
	}

	ids := []string{
		"token0",
	}

	xds := []string{
		"token1",
		"token2",
	}

	c.newDevicesList(ids)

	c.SetMsgData(data)

	c.AppendDevices(xds)

	res, err := c.Send()
	if err != nil {
		t.Error("Response Error : ", err)
	}
	if res == nil {
		t.Error("Res is nil")
	}

	if res.Success != 2 || res.Fail != 1 {
		t.Error("Parsing Success or Fail error")
	}
}

func chgUrl(ts *httptest.Server) {
	fcmServerUrl = ts.URL
}

func topicHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"message_id":6985435902064854329}`

	fmt.Fprintln(w, result)
}

func regIdHandle(w http.ResponseWriter, r *http.Request) {
	result := `{"multicast_id":1003859738309903334,"success":2,"failure":1,"canonical_ids":0,"results":[{"message_id":"0:1448128667408487%ecaaa23db3fd7efd"},{"message_id":"0:1468135657607438%ecafacddf9ff8ead"},{"error":"InvalidRegistration"}]}`
	fmt.Fprintln(w, result)

}

func TestFcmClient_SendWithPayloadCtx(t *testing.T) {
	type fields struct {
		ApiKey  string
		Message FcmMsg
	}
	type args struct {
		ctx    context.Context
		fcmMsg *FcmMsg
	}
	tests := []struct {
		name         string
		useTransport bool
		fields       fields
		args         args
		wantErr      bool
	}{
		{
			name:         "non nil context",
			useTransport: false,
			fields: fields{
				ApiKey: "key",
			},
			args: args{
				ctx: context.Background(),
				fcmMsg: &FcmMsg{
					Data: map[string]string{
						"msg": "Hello World",
						"sum": "Happy Day",
					},
					RegistrationIds: []string{"abc123"},
				},
			},
			wantErr: false,
		},
		{
			name:         "nil context",
			useTransport: false,
			fields: fields{
				ApiKey: "key",
			},
			args: args{
				ctx: nil,
				fcmMsg: &FcmMsg{
					Data: map[string]string{
						"msg": "Hello World",
						"sum": "Happy Day",
					},
					RegistrationIds: []string{"abc123"},
				},
			},
			wantErr: false,
		},
		{
			name:         "non nil context with transport",
			useTransport: true,
			fields: fields{
				ApiKey: "key",
			},
			args: args{
				ctx: context.Background(),
				fcmMsg: &FcmMsg{
					Data: map[string]string{
						"msg": "Hello World",
						"sum": "Happy Day",
					},
					RegistrationIds: []string{"abc123"},
				},
			},
			wantErr: false,
		},
		{
			name:         "nil context whout transport",
			useTransport: true,
			fields: fields{
				ApiKey: "key",
			},
			args: args{
				ctx: nil,
				fcmMsg: &FcmMsg{
					Data: map[string]string{
						"msg": "Hello World",
						"sum": "Happy Day",
					},
					RegistrationIds: []string{"abc123"},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		srv := httptest.NewServer(http.HandlerFunc(regIdHandle))
		chgUrl(srv)
		t.Run(tt.name, func(t *testing.T) {
			this := &FcmClient{}
			if !tt.useTransport {
				this = NewFcmClient(tt.fields.ApiKey)
				this.SetMsgData(tt.fields.Message)
				if this.HttpClient != nil {
					t.Errorf("FcmClient.HttpClient should be nil")
					return
				}
			} else {
				this = NewFcmClientWithTransport(tt.fields.ApiKey, nil, nil)
				this.SetMsgData(tt.fields.Message)
				if this.HttpClient == nil {
					t.Errorf("FcmClient.HttpClient should be not nil")
					return
				}
			}
			_, err := this.SendWithPayloadCtx(tt.args.ctx, tt.args.fcmMsg)
			if (err != nil) != tt.wantErr {
				t.Errorf("FcmClient.SendWithPayloadCtx() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}
